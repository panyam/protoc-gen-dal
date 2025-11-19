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

import "testing"

// testFieldMapping is a test implementation of FieldWithStrategy
type testFieldMapping struct {
	name                     string
	toTargetRenderStrategy   FieldRenderStrategy
	fromTargetRenderStrategy FieldRenderStrategy
}

func (t *testFieldMapping) GetToTargetRenderStrategy() FieldRenderStrategy {
	return t.toTargetRenderStrategy
}

func (t *testFieldMapping) GetFromTargetRenderStrategy() FieldRenderStrategy {
	return t.fromTargetRenderStrategy
}

func TestClassifyFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []*testFieldMapping
		want   *ClassifiedFields[*testFieldMapping]
	}{
		{
			name:   "empty list",
			fields: []*testFieldMapping{},
			want: &ClassifiedFields[*testFieldMapping]{
				ToTargetInline: []*testFieldMapping{},
				ToTargetSetter: []*testFieldMapping{},
				ToTargetLoop:   []*testFieldMapping{},
				FromTargetInline: []*testFieldMapping{},
				FromTargetSetter: []*testFieldMapping{},
				FromTargetLoop:   []*testFieldMapping{},
			},
		},
		{
			name: "all inline fields",
			fields: []*testFieldMapping{
				{name: "field1", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
				{name: "field2", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
			},
			want: &ClassifiedFields[*testFieldMapping]{
				ToTargetInline: []*testFieldMapping{
					{name: "field1", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
					{name: "field2", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
				},
				ToTargetSetter: []*testFieldMapping{},
				ToTargetLoop:   []*testFieldMapping{},
				FromTargetInline: []*testFieldMapping{
					{name: "field1", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
					{name: "field2", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
				},
				FromTargetSetter: []*testFieldMapping{},
				FromTargetLoop:   []*testFieldMapping{},
			},
		},
		{
			name: "mixed strategies",
			fields: []*testFieldMapping{
				{name: "inline", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
				{name: "setter", toTargetRenderStrategy: StrategySetterSimple, fromTargetRenderStrategy: StrategySetterWithError},
				{name: "loop", toTargetRenderStrategy: StrategyLoopRepeated, fromTargetRenderStrategy: StrategyLoopMap},
			},
			want: &ClassifiedFields[*testFieldMapping]{
				ToTargetInline: []*testFieldMapping{
					{name: "inline", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
				},
				ToTargetSetter: []*testFieldMapping{
					{name: "setter", toTargetRenderStrategy: StrategySetterSimple, fromTargetRenderStrategy: StrategySetterWithError},
				},
				ToTargetLoop: []*testFieldMapping{
					{name: "loop", toTargetRenderStrategy: StrategyLoopRepeated, fromTargetRenderStrategy: StrategyLoopMap},
				},
				FromTargetInline: []*testFieldMapping{
					{name: "inline", toTargetRenderStrategy: StrategyInlineValue, fromTargetRenderStrategy: StrategyInlineValue},
				},
				FromTargetSetter: []*testFieldMapping{
					{name: "setter", toTargetRenderStrategy: StrategySetterSimple, fromTargetRenderStrategy: StrategySetterWithError},
				},
				FromTargetLoop: []*testFieldMapping{
					{name: "loop", toTargetRenderStrategy: StrategyLoopRepeated, fromTargetRenderStrategy: StrategyLoopMap},
				},
			},
		},
		{
			name: "all setter variants",
			fields: []*testFieldMapping{
				{name: "simple", toTargetRenderStrategy: StrategySetterSimple, fromTargetRenderStrategy: StrategySetterSimple},
				{name: "transform", toTargetRenderStrategy: StrategySetterTransform, fromTargetRenderStrategy: StrategySetterTransform},
				{name: "withError", toTargetRenderStrategy: StrategySetterWithError, fromTargetRenderStrategy: StrategySetterWithError},
				{name: "ignoreError", toTargetRenderStrategy: StrategySetterIgnoreError, fromTargetRenderStrategy: StrategySetterIgnoreError},
			},
			want: &ClassifiedFields[*testFieldMapping]{
				ToTargetInline: []*testFieldMapping{},
				ToTargetSetter: []*testFieldMapping{
					{name: "simple", toTargetRenderStrategy: StrategySetterSimple, fromTargetRenderStrategy: StrategySetterSimple},
					{name: "transform", toTargetRenderStrategy: StrategySetterTransform, fromTargetRenderStrategy: StrategySetterTransform},
					{name: "withError", toTargetRenderStrategy: StrategySetterWithError, fromTargetRenderStrategy: StrategySetterWithError},
					{name: "ignoreError", toTargetRenderStrategy: StrategySetterIgnoreError, fromTargetRenderStrategy: StrategySetterIgnoreError},
				},
				ToTargetLoop: []*testFieldMapping{},
				FromTargetInline: []*testFieldMapping{},
				FromTargetSetter: []*testFieldMapping{
					{name: "simple", toTargetRenderStrategy: StrategySetterSimple, fromTargetRenderStrategy: StrategySetterSimple},
					{name: "transform", toTargetRenderStrategy: StrategySetterTransform, fromTargetRenderStrategy: StrategySetterTransform},
					{name: "withError", toTargetRenderStrategy: StrategySetterWithError, fromTargetRenderStrategy: StrategySetterWithError},
					{name: "ignoreError", toTargetRenderStrategy: StrategySetterIgnoreError, fromTargetRenderStrategy: StrategySetterIgnoreError},
				},
				FromTargetLoop: []*testFieldMapping{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFields(tt.fields)

			// Check ToTarget classifications
			if len(got.ToTargetInline) != len(tt.want.ToTargetInline) {
				t.Errorf("ToTargetInline count = %d, want %d", len(got.ToTargetInline), len(tt.want.ToTargetInline))
			}
			if len(got.ToTargetSetter) != len(tt.want.ToTargetSetter) {
				t.Errorf("ToTargetSetter count = %d, want %d", len(got.ToTargetSetter), len(tt.want.ToTargetSetter))
			}
			if len(got.ToTargetLoop) != len(tt.want.ToTargetLoop) {
				t.Errorf("ToTargetLoop count = %d, want %d", len(got.ToTargetLoop), len(tt.want.ToTargetLoop))
			}

			// Check FromTarget classifications
			if len(got.FromTargetInline) != len(tt.want.FromTargetInline) {
				t.Errorf("FromTargetInline count = %d, want %d", len(got.FromTargetInline), len(tt.want.FromTargetInline))
			}
			if len(got.FromTargetSetter) != len(tt.want.FromTargetSetter) {
				t.Errorf("FromTargetSetter count = %d, want %d", len(got.FromTargetSetter), len(tt.want.FromTargetSetter))
			}
			if len(got.FromTargetLoop) != len(tt.want.FromTargetLoop) {
				t.Errorf("FromTargetLoop count = %d, want %d", len(got.FromTargetLoop), len(tt.want.FromTargetLoop))
			}
		})
	}
}

func TestAddRenderStrategies(t *testing.T) {
	tests := []struct {
		name                     string
		toTargetConversionType   ConversionType
		fromTargetConversionType ConversionType
		sourceIsPointer          bool
		targetIsPointer          bool
		isRepeated               bool
		isMap                    bool
		toTargetHasConverter     bool
		fromTargetHasConverter   bool
		wantToTarget             FieldRenderStrategy
		wantFromTarget           FieldRenderStrategy
	}{
		{
			name:                     "simple assignment, non-pointer",
			toTargetConversionType:   ConvertByAssignment,
			fromTargetConversionType: ConvertByAssignment,
			sourceIsPointer:          false,
			targetIsPointer:          false,
			wantToTarget:             StrategyInlineValue,
			wantFromTarget:           StrategyInlineValue,
		},
		{
			name:                     "assignment with pointer",
			toTargetConversionType:   ConvertByAssignment,
			fromTargetConversionType: ConvertByAssignment,
			sourceIsPointer:          true,
			targetIsPointer:          true,
			wantToTarget:             StrategySetterSimple,
			wantFromTarget:           StrategySetterSimple,
		},
		{
			name:                     "repeated with converter",
			toTargetConversionType:   ConvertByTransformerWithError,
			fromTargetConversionType: ConvertByTransformerWithError,
			isRepeated:               true,
			toTargetHasConverter:     true,
			fromTargetHasConverter:   true,
			wantToTarget:             StrategyLoopRepeated,
			wantFromTarget:           StrategyLoopRepeated,
		},
		{
			name:                     "map with converter",
			toTargetConversionType:   ConvertByTransformerWithError,
			fromTargetConversionType: ConvertByTransformerWithError,
			isMap:                    true,
			toTargetHasConverter:     true,
			fromTargetHasConverter:   true,
			wantToTarget:             StrategyLoopMap,
			wantFromTarget:           StrategyLoopMap,
		},
		{
			name:                     "transformer with error",
			toTargetConversionType:   ConvertByTransformerWithError,
			fromTargetConversionType: ConvertByTransformerWithError,
			sourceIsPointer:          false,
			targetIsPointer:          false,
			wantToTarget:             StrategySetterWithError,
			wantFromTarget:           StrategySetterWithError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToTarget, gotFromTarget := AddRenderStrategies(
				tt.toTargetConversionType,
				tt.fromTargetConversionType,
				tt.sourceIsPointer,
				tt.targetIsPointer,
				tt.isRepeated,
				tt.isMap,
				tt.toTargetHasConverter,
				tt.fromTargetHasConverter,
			)

			if gotToTarget != tt.wantToTarget {
				t.Errorf("ToTarget strategy = %v, want %v", gotToTarget, tt.wantToTarget)
			}
			if gotFromTarget != tt.wantFromTarget {
				t.Errorf("FromTarget strategy = %v, want %v", gotFromTarget, tt.wantFromTarget)
			}
		})
	}
}
