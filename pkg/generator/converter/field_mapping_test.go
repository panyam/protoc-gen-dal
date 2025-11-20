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

func TestBuildSameTypeMapping(t *testing.T) {
	tests := []struct {
		name            string
		sourceKind      string
		targetKind      string
		fieldName       string
		excludeMessages bool
		wantFound       bool
		wantToCode      string
		wantFromCode    string
	}{
		{
			name:         "same scalar types",
			sourceKind:   "int64",
			targetKind:   "int64",
			fieldName:    "Count",
			wantFound:    true,
			wantToCode:   "src.Count",
			wantFromCode: "src.Count",
		},
		{
			name:       "different types",
			sourceKind: "int64",
			targetKind: "string",
			fieldName:  "Count",
			wantFound:  false,
		},
		{
			name:            "same message types with excludeMessages=true",
			sourceKind:      "message",
			targetKind:      "message",
			fieldName:       "Author",
			excludeMessages: true,
			wantFound:       false,
		},
		{
			name:            "same message types with excludeMessages=false",
			sourceKind:      "message",
			targetKind:      "message",
			fieldName:       "Author",
			excludeMessages: false,
			wantFound:       true,
			wantToCode:      "src.Author",
			wantFromCode:    "src.Author",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := &FieldMapping{}
			found := BuildSameTypeMapping(tt.sourceKind, tt.targetKind, tt.fieldName, tt.excludeMessages, mapping)

			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}

			if !tt.wantFound {
				return
			}

			if mapping.ToTargetCode != tt.wantToCode {
				t.Errorf("ToTargetCode = %q, want %q", mapping.ToTargetCode, tt.wantToCode)
			}

			if mapping.FromTargetCode != tt.wantFromCode {
				t.Errorf("FromTargetCode = %q, want %q", mapping.FromTargetCode, tt.wantFromCode)
			}

			if mapping.ToTargetConversionType != ConvertByAssignment {
				t.Errorf("ToTargetConversionType = %v, want %v", mapping.ToTargetConversionType, ConvertByAssignment)
			}

			if mapping.FromTargetConversionType != ConvertByAssignment {
				t.Errorf("FromTargetConversionType = %v, want %v", mapping.FromTargetConversionType, ConvertByAssignment)
			}
		})
	}
}

func TestBuildNumericTypeMapping(t *testing.T) {
	tests := []struct {
		name         string
		sourceKind   string
		targetKind   string
		fieldName    string
		wantFound    bool
		wantToCode   string
		wantFromCode string
	}{
		{
			name:         "int64 to int32",
			sourceKind:   "int64",
			targetKind:   "int32",
			fieldName:    "Count",
			wantFound:    true,
			wantToCode:   "int32(src.Count)",
			wantFromCode: "int64(src.Count)",
		},
		{
			name:         "uint32 to uint64",
			sourceKind:   "uint32",
			targetKind:   "uint64",
			fieldName:    "Age",
			wantFound:    true,
			wantToCode:   "uint64(src.Age)",
			wantFromCode: "uint32(src.Age)",
		},
		{
			name:       "non-numeric source",
			sourceKind: "string",
			targetKind: "int64",
			fieldName:  "Count",
			wantFound:  false,
		},
		{
			name:       "non-numeric target",
			sourceKind: "int64",
			targetKind: "string",
			fieldName:  "Count",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := &FieldMapping{}
			found := BuildNumericTypeMapping(tt.sourceKind, tt.targetKind, tt.fieldName, mapping)

			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}

			if !tt.wantFound {
				return
			}

			if mapping.ToTargetCode != tt.wantToCode {
				t.Errorf("ToTargetCode = %q, want %q", mapping.ToTargetCode, tt.wantToCode)
			}

			if mapping.FromTargetCode != tt.wantFromCode {
				t.Errorf("FromTargetCode = %q, want %q", mapping.FromTargetCode, tt.wantFromCode)
			}

			if mapping.ToTargetConversionType != ConvertByAssignment {
				t.Errorf("ToTargetConversionType = %v, want %v", mapping.ToTargetConversionType, ConvertByAssignment)
			}

			if mapping.FromTargetConversionType != ConvertByAssignment {
				t.Errorf("FromTargetConversionType = %v, want %v", mapping.FromTargetConversionType, ConvertByAssignment)
			}
		})
	}
}

// Note: More comprehensive tests for BuildKnownTypeMapping, BuildMapFieldMapping,
// BuildRepeatedFieldMapping, and BuildMessageToMessageMapping require complex protogen.Field
// mocking. These functions are tested through integration tests when GORM and Datastore
// generators use them.
