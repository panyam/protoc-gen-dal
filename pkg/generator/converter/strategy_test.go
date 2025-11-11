// Copyright 2025 Sri Panyam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"testing"
)

func TestDetermineRenderStrategy(t *testing.T) {
	tests := []struct {
		name           string
		conversionType ConversionType
		chars          FieldCharacteristics
		want           FieldRenderStrategy
	}{
		{
			name:           "repeated message field",
			conversionType: ConvertByTransformerWithError,
			chars: FieldCharacteristics{
				IsRepeated:         true,
				HasMessageElements: true,
			},
			want: StrategyLoopRepeated,
		},
		{
			name:           "map with message values",
			conversionType: ConvertByTransformerWithError,
			chars: FieldCharacteristics{
				IsMap:            true,
				HasMessageValues: true,
			},
			want: StrategyLoopMap,
		},
		{
			name:           "pointer field with assignment",
			conversionType: ConvertByAssignment,
			chars: FieldCharacteristics{
				IsPointer: true,
			},
			want: StrategySetterSimple,
		},
		{
			name:           "pointer field with transformer",
			conversionType: ConvertByTransformer,
			chars: FieldCharacteristics{
				IsPointer: true,
			},
			want: StrategySetterTransform,
		},
		{
			name:           "pointer field with error",
			conversionType: ConvertByTransformerWithError,
			chars: FieldCharacteristics{
				IsPointer: true,
			},
			want: StrategySetterWithError,
		},
		{
			name:           "pointer field with ignorable error",
			conversionType: ConvertByTransformerWithIgnorableError,
			chars: FieldCharacteristics{
				IsPointer: true,
			},
			want: StrategySetterIgnoreError,
		},
		{
			name:           "non-pointer direct assignment",
			conversionType: ConvertByAssignment,
			chars: FieldCharacteristics{
				IsPointer: false,
			},
			want: StrategyInlineValue,
		},
		{
			name:           "non-pointer transformer without error",
			conversionType: ConvertByTransformer,
			chars: FieldCharacteristics{
				IsPointer: false,
			},
			want: StrategyInlineValue,
		},
		{
			name:           "non-pointer transformer with error",
			conversionType: ConvertByTransformerWithError,
			chars: FieldCharacteristics{
				IsPointer: false,
			},
			want: StrategySetterWithError,
		},
		{
			name:           "non-pointer transformer with ignorable error",
			conversionType: ConvertByTransformerWithIgnorableError,
			chars: FieldCharacteristics{
				IsPointer: false,
			},
			want: StrategySetterIgnoreError,
		},
		{
			name:           "repeated primitive (not message)",
			conversionType: ConvertByAssignment,
			chars: FieldCharacteristics{
				IsRepeated:         true,
				HasMessageElements: false, // primitives
			},
			want: StrategyInlineValue, // direct assignment, can be inline
		},
		{
			name:           "map with primitive values (not message)",
			conversionType: ConvertByAssignment,
			chars: FieldCharacteristics{
				IsMap:            true,
				HasMessageValues: false, // primitives
			},
			want: StrategyInlineValue, // direct assignment, can be inline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineRenderStrategy(tt.conversionType, tt.chars)
			if got != tt.want {
				t.Errorf("DetermineRenderStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyFieldsForRendering(t *testing.T) {
	fields := []*ClassifiedField{
		{
			SourceField:    "Name",
			TargetField:    "Name",
			RenderStrategy: StrategyInlineValue,
		},
		{
			SourceField:    "Age",
			TargetField:    "Age",
			RenderStrategy: StrategyInlineValue,
		},
		{
			SourceField:    "Email",
			TargetField:    "Email",
			RenderStrategy: StrategySetterSimple,
		},
		{
			SourceField:    "CreatedAt",
			TargetField:    "CreatedAt",
			RenderStrategy: StrategySetterTransform,
		},
		{
			SourceField:    "Author",
			TargetField:    "Author",
			RenderStrategy: StrategySetterWithError,
		},
		{
			SourceField:    "Contributors",
			TargetField:    "Contributors",
			RenderStrategy: StrategyLoopRepeated,
		},
		{
			SourceField:    "Departments",
			TargetField:    "Departments",
			RenderStrategy: StrategyLoopMap,
		},
	}

	inlineFields, setterFields, loopFields := ClassifyFieldsForRendering(fields)

	// Check inline fields
	if len(inlineFields) != 2 {
		t.Errorf("Expected 2 inline fields, got %d", len(inlineFields))
	}
	if inlineFields[0].SourceField != "Name" {
		t.Errorf("Expected first inline field to be Name, got %s", inlineFields[0].SourceField)
	}
	if inlineFields[1].SourceField != "Age" {
		t.Errorf("Expected second inline field to be Age, got %s", inlineFields[1].SourceField)
	}

	// Check setter fields
	if len(setterFields) != 3 {
		t.Errorf("Expected 3 setter fields, got %d", len(setterFields))
	}
	expectedSetters := []string{"Email", "CreatedAt", "Author"}
	for i, field := range setterFields {
		if field.SourceField != expectedSetters[i] {
			t.Errorf("Expected setter field %d to be %s, got %s", i, expectedSetters[i], field.SourceField)
		}
	}

	// Check loop fields
	if len(loopFields) != 2 {
		t.Errorf("Expected 2 loop fields, got %d", len(loopFields))
	}
	if loopFields[0].SourceField != "Contributors" {
		t.Errorf("Expected first loop field to be Contributors, got %s", loopFields[0].SourceField)
	}
	if loopFields[1].SourceField != "Departments" {
		t.Errorf("Expected second loop field to be Departments, got %s", loopFields[1].SourceField)
	}
}
