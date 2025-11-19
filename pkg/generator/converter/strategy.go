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

// ConversionType represents the user's intent for how a field should be converted.
//
// This is determined by the generator based on:
// - Proto annotations (custom converters)
// - Field types (same type = assignment, different = transform)
// - Built-in conversion rules (Timestamp ↔ int64, numeric casts)
//
// ConversionType is user-facing domain logic, not an implementation detail.
type ConversionType int

const (
	// ConvertIgnore means skip this field in generated converter (caller handles it)
	ConvertIgnore ConversionType = iota

	// ConvertByAssignment means direct assignment: out.Field = src.Field
	ConvertByAssignment

	// ConvertByTransformer means call converter function without error checking
	// Example: out.Field = timestampToInt64(src.Field)
	ConvertByTransformer

	// ConvertByTransformerWithError means call converter and check error
	// Example: out.Field, err = AuthorToAuthorGORM(src.Field, nil, nil)
	ConvertByTransformerWithError

	// ConvertByTransformerWithIgnorableError means call converter but ignore error
	// Example: out.Field, _ = converter(src.Field)
	ConvertByTransformerWithIgnorableError
)

// FieldRenderStrategy represents how to render a field conversion in the template.
//
// This is an implementation detail derived from ConversionType plus field characteristics
// (pointer, repeated, map, etc.). The template uses this to decide where and how to emit code.
//
// FieldRenderStrategy is internal, not exposed to users.
type FieldRenderStrategy int

const (
	// StrategyInlineValue means render in struct literal initialization
	// Only for non-pointer, non-error fields
	// Example: out := Type{ Field: src.Field, Age: int64(src.Age) }
	StrategyInlineValue FieldRenderStrategy = iota

	// StrategySetterSimple means render as simple assignment statement
	// Used for pointer fields with direct assignment (with nil check)
	// Example: if src.Field != nil { out.Field = src.Field }
	StrategySetterSimple

	// StrategySetterTransform means render as assignment with transformation (no error)
	// Example: out.CreatedAt = timestampToInt64(src.CreatedAt)
	StrategySetterTransform

	// StrategySetterWithError means render as assignment with error checking
	// Example: out.Author, err = AuthorToAuthorGORM(src.Author, nil, nil); if err != nil { return }
	StrategySetterWithError

	// StrategySetterIgnoreError means render as assignment ignoring error
	// Example: out.Field, _ = converter(src.Field)
	StrategySetterIgnoreError

	// StrategyLoopRepeated means render as loop over repeated field
	// Example: for i, item := range src.Authors { ... }
	StrategyLoopRepeated

	// StrategyLoopMap means render as loop over map field
	// Example: for key, val := range src.Departments { ... }
	StrategyLoopMap
)

// FieldCharacteristics captures the properties of a field that affect rendering strategy.
type FieldCharacteristics struct {
	IsPointer          bool // Whether field is a pointer type
	IsRepeated         bool // Whether field is a repeated (slice)
	IsMap              bool // Whether field is a map
	HasMessageElements bool // For repeated: are elements messages?
	HasMessageValues   bool // For map: are values messages?
}

// DetermineRenderStrategy decides how to render a field conversion in the template.
//
// This is the core strategy selection logic that maps user intent (ConversionType)
// to implementation (FieldRenderStrategy) based on field characteristics.
//
// Parameters:
//   - conversionType: the user's intent for this conversion
//   - chars: characteristics of the field being converted
//
// Returns:
//   - FieldRenderStrategy indicating how to render this conversion
func DetermineRenderStrategy(conversionType ConversionType, chars FieldCharacteristics) FieldRenderStrategy {
	// Special case: collections with message types need loops
	if chars.IsRepeated && chars.HasMessageElements {
		return StrategyLoopRepeated
	}
	if chars.IsMap && chars.HasMessageValues {
		return StrategyLoopMap
	}

	// Pointer fields cannot be inline (need nil checks)
	if chars.IsPointer {
		switch conversionType {
		case ConvertByAssignment:
			return StrategySetterSimple
		case ConvertByTransformer:
			return StrategySetterTransform
		case ConvertByTransformerWithError:
			return StrategySetterWithError
		case ConvertByTransformerWithIgnorableError:
			return StrategySetterIgnoreError
		default:
			return StrategySetterSimple
		}
	}

	// Non-pointer fields can be inline if they don't need error handling
	switch conversionType {
	case ConvertByAssignment, ConvertByTransformer:
		// Simple assignment or transformation without error can be inline
		return StrategyInlineValue
	case ConvertByTransformerWithError:
		// Error handling requires statement (not inline)
		return StrategySetterWithError
	case ConvertByTransformerWithIgnorableError:
		// Error ignoring requires statement (not inline)
		return StrategySetterIgnoreError
	default:
		// Default to inline for safety
		return StrategyInlineValue
	}
}

// ClassifiedField contains all information needed to render a field conversion.
//
// This structure separates concerns:
// - ConversionType: what conversion to apply (user intent)
// - RenderStrategy: how to render it (implementation detail)
// - Pre-generated expressions: the actual code to emit
type ClassifiedField struct {
	// Field names
	SourceField string
	TargetField string

	// Conversion strategy (user intent)
	ConversionType ConversionType

	// Render strategy (implementation detail)
	RenderStrategy FieldRenderStrategy

	// Pre-generated code expressions
	InlineExpr string // For StrategyInlineValue: "src.Field" or "int64(src.Age)"
	SetterExpr string // For setter strategies: "src.Field" or "converter(src.Field, nil, nil)"

	// Metadata for rendering
	HasNilCheck bool // Whether to wrap setter in "if src.Field != nil"
	IsPointer   bool // Whether target field is pointer
	IsRepeated  bool // Whether field is repeated
	IsMap       bool // Whether field is map

	// For loop-based strategies
	ConverterFunc     string // Name of converter function (e.g., "AuthorToAuthorGORM")
	ElementType       string // Element/value type (e.g., "AuthorGORM")
	SourceElementType string // Source element/value type (e.g., "Author")
}

// ClassifyFieldsForRendering separates classified fields into rendering groups.
//
// This groups fields by their render strategy to simplify template logic:
// - InlineFields: go in struct literal
// - SetterFields: go in post-construction statements
// - LoopFields: go in loop blocks
//
// Parameters:
//   - fields: all classified fields for a converter
//
// Returns:
//   - inlineFields: fields to render in struct literal
//   - setterFields: fields to render as setter statements
//   - loopFields: fields to render as loops
func ClassifyFieldsForRendering(fields []*ClassifiedField) (inlineFields, setterFields, loopFields []*ClassifiedField) {
	for _, field := range fields {
		switch field.RenderStrategy {
		case StrategyInlineValue:
			inlineFields = append(inlineFields, field)
		case StrategySetterSimple, StrategySetterTransform, StrategySetterWithError, StrategySetterIgnoreError:
			setterFields = append(setterFields, field)
		case StrategyLoopRepeated, StrategyLoopMap:
			loopFields = append(loopFields, field)
		}
	}
	return inlineFields, setterFields, loopFields
}
