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

// FieldWithStrategy represents any field mapping that has render strategies.
// This interface allows generic field classification for both GORM and Datastore.
type FieldWithStrategy interface {
	GetToTargetRenderStrategy() FieldRenderStrategy
	GetFromTargetRenderStrategy() FieldRenderStrategy
}

// ClassifiedFields holds fields grouped by their render strategy.
// Used by both GORM and Datastore generators for template rendering.
type ClassifiedFields[T FieldWithStrategy] struct {
	// ToTarget direction (API → Target)
	ToTargetInline []T
	ToTargetSetter []T
	ToTargetLoop   []T

	// FromTarget direction (Target → API)
	FromTargetInline []T
	FromTargetSetter []T
	FromTargetLoop   []T
}

// ClassifyFields groups field mappings by their render strategy.
// This determines how fields are rendered in converter templates:
// - Inline: Direct assignment in struct literal
// - Setter: Post-construction assignment statements
// - Loop: Separate loop blocks for collections
//
// Generic function works with any type that implements FieldWithStrategy.
func ClassifyFields[T FieldWithStrategy](fields []T) *ClassifiedFields[T] {
	result := &ClassifiedFields[T]{
		ToTargetInline: make([]T, 0),
		ToTargetSetter: make([]T, 0),
		ToTargetLoop:   make([]T, 0),
		FromTargetInline: make([]T, 0),
		FromTargetSetter: make([]T, 0),
		FromTargetLoop:   make([]T, 0),
	}

	for _, field := range fields {
		// Classify for ToTarget direction (API → GORM/Datastore)
		switch field.GetToTargetRenderStrategy() {
		case StrategyInlineValue:
			result.ToTargetInline = append(result.ToTargetInline, field)
		case StrategySetterSimple, StrategySetterTransform,
			StrategySetterWithError, StrategySetterIgnoreError:
			result.ToTargetSetter = append(result.ToTargetSetter, field)
		case StrategyLoopRepeated, StrategyLoopMap:
			result.ToTargetLoop = append(result.ToTargetLoop, field)
		}

		// Classify for FromTarget direction (GORM/Datastore → API)
		switch field.GetFromTargetRenderStrategy() {
		case StrategyInlineValue:
			result.FromTargetInline = append(result.FromTargetInline, field)
		case StrategySetterSimple, StrategySetterTransform,
			StrategySetterWithError, StrategySetterIgnoreError:
			result.FromTargetSetter = append(result.FromTargetSetter, field)
		case StrategyLoopRepeated, StrategyLoopMap:
			result.FromTargetLoop = append(result.FromTargetLoop, field)
		}
	}

	return result
}

// AddRenderStrategies calculates and sets render strategies for a field mapping.
// This determines HOW to render the conversion (inline, setter, loop) based on
// WHAT conversion to apply (ConversionType) and field characteristics.
//
// This function is generic - it works by accepting pointer interfaces to allow
// modification of either GORM FieldMappingData or Datastore FieldMapping.
func AddRenderStrategies(
	toTargetConversionType, fromTargetConversionType ConversionType,
	sourceIsPointer, targetIsPointer bool,
	isRepeated, isMap bool,
	toTargetHasConverter, fromTargetHasConverter bool,
) (toTargetStrategy, fromTargetStrategy FieldRenderStrategy) {

	// Build field characteristics for ToTarget direction (API → Target)
	chars := FieldCharacteristics{
		IsPointer:          sourceIsPointer,
		IsRepeated:         isRepeated,
		IsMap:              isMap,
		HasMessageElements: isRepeated && toTargetHasConverter,
		HasMessageValues:   isMap && toTargetHasConverter,
	}

	toTargetStrategy = DetermineRenderStrategy(toTargetConversionType, chars)

	// Build characteristics for FromTarget direction (Target → API)
	charsFrom := chars
	charsFrom.IsPointer = targetIsPointer
	charsFrom.HasMessageElements = isRepeated && fromTargetHasConverter
	charsFrom.HasMessageValues = isMap && fromTargetHasConverter

	fromTargetStrategy = DetermineRenderStrategy(fromTargetConversionType, charsFrom)

	return toTargetStrategy, fromTargetStrategy
}
