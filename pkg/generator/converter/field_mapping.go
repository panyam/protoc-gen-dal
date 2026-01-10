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
	"fmt"
	"log"

	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/registry"
	"google.golang.org/protobuf/compiler/protogen"
)

// FieldMapping contains all data needed to convert a field between source and target types.
// This is the unified field mapping structure used by both GORM and Datastore generators.
type FieldMapping struct {
	// Source and target field names
	SourceField string
	TargetField string

	// Conversion code for direct transformations (empty if using converter functions)
	ToTargetCode   string // API → Target conversion code
	FromTargetCode string // Target → API conversion code

	// Conversion and render strategies
	ToTargetConversionType   ConversionType      // How to convert source → target (user intent)
	FromTargetConversionType ConversionType      // How to convert target → source (user intent)
	ToTargetRenderStrategy   FieldRenderStrategy // How to render source → target (implementation)
	FromTargetRenderStrategy FieldRenderStrategy // How to render target → source (implementation)

	// Converter function names for nested message conversions
	ToTargetConverterFunc   string // e.g., "AuthorToAuthorGORM"
	FromTargetConverterFunc string // e.g., "AuthorFromAuthorGORM"

	// Pointer characteristics
	SourceIsPointer bool // Whether source field is a pointer type (needs nil check)
	TargetIsPointer bool // Whether target field is a pointer type (affects assignment)

	// Oneof characteristics
	SourceIsOneofMember bool // Whether source field is part of a oneof (requires getter access)

	// Collection characteristics
	IsRepeated bool // Whether this is a repeated field (needs loop-based conversion)
	IsMap      bool // Whether this is a map field (needs loop-based conversion)

	// Element types for repeated/map fields
	TargetElementType string // For repeated/map: Go type of target element/value (e.g., "AuthorGORM")
	SourceElementType string // For repeated/map: Go type of source element/value (e.g., "Author")
	MapKeyType        string // For map fields: Go type of map key (e.g., "string", "int32", "bool")

	// Package information
	SourcePkgName string // Source package name (e.g., "api" or "testapi") - needed for type references
}

// sourceFieldAccess generates the correct source field access expression.
// For oneof members, it returns "src.GetFieldName()" (getter method required).
// For regular fields, it returns "src.FieldName" (direct field access).
func sourceFieldAccess(fieldName string, isOneofMember bool) string {
	if isOneofMember {
		return fmt.Sprintf("src.Get%s()", fieldName)
	}
	return fmt.Sprintf("src.%s", fieldName)
}

// Implement FieldWithStrategy interface for FieldMapping
func (f *FieldMapping) GetToTargetRenderStrategy() FieldRenderStrategy {
	return f.ToTargetRenderStrategy
}

func (f *FieldMapping) GetFromTargetRenderStrategy() FieldRenderStrategy {
	return f.FromTargetRenderStrategy
}

// MapFieldMappingParams contains parameters for map field mapping.
type MapFieldMappingParams struct {
	SourceField   *protogen.Field
	TargetField   *protogen.Field
	Registry      *registry.ConverterRegistry
	MsgRegistry   *common.MessageRegistry
	FieldName     string
}

// RepeatedFieldMappingParams contains parameters for repeated field mapping.
type RepeatedFieldMappingParams struct {
	SourceField   *protogen.Field
	TargetField   *protogen.Field
	Registry      *registry.ConverterRegistry
	MsgRegistry   *common.MessageRegistry
	FieldName     string
}

// MessageFieldMappingParams contains parameters for message field mapping.
type MessageFieldMappingParams struct {
	SourceField   *protogen.Field
	TargetField   *protogen.Field
	SourceKind    string
	TargetKind    string
	Registry      *registry.ConverterRegistry
	MsgRegistry   *common.MessageRegistry
	FieldName     string
	IsRepeated    bool
	IsMap         bool
}

// BuildMapFieldMapping handles map field conversions.
// Returns true if this is a map field with primitive values (complete - should return early).
// Returns false if this is a map with message values (needs further processing for converter lookup).
// Modifies mapping in place with conversion details.
func BuildMapFieldMapping(params MapFieldMappingParams, mapping *FieldMapping) bool {
	if !params.SourceField.Desc.IsMap() {
		return false
	}

	// Map fields: check the value type to determine conversion
	mapEntry := params.SourceField.Message
	keyField := mapEntry.Fields[0]   // key field is always index 0
	valueField := mapEntry.Fields[1] // value field is always index 1
	valueKind := valueField.Desc.Kind().String()

	// Extract map key type (string, int32, int64, uint32, uint64, bool, etc.)
	mapping.MapKeyType = common.ProtoKindToGoType(keyField.Desc.Kind().String())

	if valueKind == "message" {
		// map<K, MessageType> - needs loop-based converter for values
		sourceMsg := valueField.Message
		targetMapEntry := params.TargetField.Message
		targetValueField := targetMapEntry.Fields[1]
		var targetMsg *protogen.Message
		var targetFieldMsg *protogen.Message = targetValueField.Message

		// Check if target is a well-known type first
		if _, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); !isWellKnown {
			// Use MessageRegistry to resolve source → target mapping
			targetMsg = params.MsgRegistry.LookupTargetMessage(sourceMsg)
			if targetMsg == nil {
				targetMsg = targetFieldMsg
			}
		}

		if sourceMsg != nil && targetMsg != nil {
			sourceTypeName := string(sourceMsg.Desc.Name())
			targetTypeName := params.MsgRegistry.GetStructName(targetMsg)

			// Check if converter exists for this nested type
			if params.Registry.HasConverter(sourceTypeName, targetTypeName) {
				mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
				mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
				mapping.TargetElementType = targetTypeName
				mapping.SourceElementType = sourceTypeName
				mapping.ToTargetConversionType = ConvertByTransformerWithError
				mapping.FromTargetConversionType = ConvertByTransformerWithError
				return true
			}
		}
		// No converter found - return false to continue processing
		return false
	}

	// map<K, primitive> - direct assignment (copy entire map)
	// Use sourceFieldAccess to handle oneof member access correctly
	mapping.ToTargetCode = sourceFieldAccess(params.FieldName, mapping.SourceIsOneofMember)
	mapping.FromTargetCode = fmt.Sprintf("src.%s", params.FieldName)
	mapping.ToTargetConversionType = ConvertByAssignment
	mapping.FromTargetConversionType = ConvertByAssignment
	return true // Complete - return early
}

// BuildRepeatedFieldMapping handles repeated field conversions.
// Returns true if this is a repeated field with primitive elements (complete - should return early).
// Returns false if this is a repeated field with message elements (needs further processing).
// Modifies mapping in place with conversion details.
func BuildRepeatedFieldMapping(params RepeatedFieldMappingParams, mapping *FieldMapping) bool {
	if !params.SourceField.Desc.IsList() {
		return false
	}

	// Repeated fields: check the element type to determine conversion
	elementKind := params.SourceField.Desc.Kind().String()

	if elementKind == "message" {
		// []MessageType - needs loop-based converter for elements
		sourceMsg := params.SourceField.Message
		var targetMsg *protogen.Message
		var targetFieldMsg *protogen.Message = params.TargetField.Message

		// Check if target is a well-known type first
		if _, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); !isWellKnown {
			// Use MessageRegistry to resolve source → target mapping
			targetMsg = params.MsgRegistry.LookupTargetMessage(sourceMsg)
			if targetMsg == nil {
				targetMsg = targetFieldMsg
			}
		}

		if sourceMsg != nil && targetMsg != nil {
			sourceTypeName := string(sourceMsg.Desc.Name())
			targetTypeName := params.MsgRegistry.GetStructName(targetMsg)

			// Check if converter exists for this nested type
			if params.Registry.HasConverter(sourceTypeName, targetTypeName) {
				mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
				mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
				mapping.TargetElementType = targetTypeName
				mapping.SourceElementType = sourceTypeName
				mapping.ToTargetConversionType = ConvertByTransformerWithError
				mapping.FromTargetConversionType = ConvertByTransformerWithError
				return true
			}
		}
		// No converter found - return false to continue processing
		return false
	}

	// []primitive - direct assignment (copy entire slice)
	// Use sourceFieldAccess to handle oneof member access correctly
	mapping.ToTargetCode = sourceFieldAccess(params.FieldName, mapping.SourceIsOneofMember)
	mapping.FromTargetCode = fmt.Sprintf("src.%s", params.FieldName)
	mapping.ToTargetConversionType = ConvertByAssignment
	mapping.FromTargetConversionType = ConvertByAssignment
	return true // Complete - return early
}

// BuildMessageToMessageMapping handles message→message conversions including google.protobuf.Any.
// Returns: 0 = not applicable, 1 = conversion found, -1 = skip field (no converter).
// Modifies mapping in place with conversion details.
func BuildMessageToMessageMapping(params MessageFieldMappingParams, mapping *FieldMapping) int {
	if params.SourceKind != "message" || params.TargetKind != "message" {
		return 0 // Not applicable
	}

	sourceMsg := params.SourceField.Message
	targetFieldMsg := params.TargetField.Message
	var targetMsg *protogen.Message

	// For regular (non-repeated, non-map) message fields, resolve target via MessageRegistry
	if !params.IsRepeated && !params.IsMap {
		// Check if target is a well-known type first
		if _, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); !isWellKnown {
			// Use MessageRegistry to resolve source → target mapping
			targetMsg = params.MsgRegistry.LookupTargetMessage(sourceMsg)
			if targetMsg == nil {
				targetMsg = targetFieldMsg
			}
		}
	}

	// Check if target is google.protobuf.Any (well-known type)
	if wkt, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); isWellKnown {
		if wkt.ProtoFullName == "google.protobuf.Any" {
			// Message → google.protobuf.Any serialization
			sourcePkgName := common.ExtractPackageName(sourceMsg)
			sourceTypeName := fmt.Sprintf("*%s.%s", sourcePkgName, sourceMsg.Desc.Name())

			if params.IsRepeated {
				// For repeated Any fields: []Message → [][]byte
				mapping.SourceElementType = string(sourceMsg.Desc.Name())
				mapping.TargetElementType = "[]byte"
				mapping.ToTargetConverterFunc = "converters.MessageToAnyBytesConverter"
				mapping.FromTargetConverterFunc = fmt.Sprintf("converters.AnyBytesToMessageConverter[%s]", sourceTypeName)
				mapping.ToTargetConversionType = ConvertByTransformerWithError
				mapping.FromTargetConversionType = ConvertByTransformerWithError
				return 1
			}

			// For single Any fields: Message → []byte
			mapping.ToTargetCode = fmt.Sprintf("converters.MessageToAnyBytes(src.%s)", params.FieldName)
			mapping.FromTargetCode = fmt.Sprintf("converters.AnyBytesToMessage[%s](src.%s)", sourceTypeName, params.FieldName)
			mapping.ToTargetConversionType = ConvertByTransformerWithError
			mapping.FromTargetConversionType = ConvertByTransformerWithError
			return 1
		}
		// Other well-known types handled by BuildConversionCode
		return 0
	}

	// Handle regular message→message conversions using converter registry
	if sourceMsg != nil && targetMsg != nil {
		sourceTypeName := string(sourceMsg.Desc.Name())
		targetTypeName := params.MsgRegistry.GetStructName(targetMsg)

		// Check if converter exists for this nested type
		if params.Registry.HasConverter(sourceTypeName, targetTypeName) {
			mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
			mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
			mapping.ToTargetConversionType = ConvertByTransformerWithError
			mapping.FromTargetConversionType = ConvertByTransformerWithError

			// For repeated/map fields, store element types (not needed for regular fields)
			if params.IsRepeated || params.IsMap {
				mapping.TargetElementType = targetTypeName
				mapping.SourceElementType = sourceTypeName
			}

			return 1
		}

		// No converter available - warn user and mark for skip
		if params.IsRepeated {
			log.Printf("WARNING: Field '%s' is []%s but no converter found for element type.\n",
				params.FieldName, sourceTypeName)
		} else if params.IsMap {
			log.Printf("WARNING: Field '%s' is map<K, %s> but no converter found for value type.\n",
				params.FieldName, sourceTypeName)
		} else {
			log.Printf("WARNING: Field '%s' has matching message types (%s → %s) but no converter found.\n",
				params.FieldName, sourceTypeName, targetTypeName)
		}
		log.Printf("         If you want automatic conversion, add 'source' annotation to %s message.\n",
			targetMsg.Desc.Name())
		log.Printf("         Skipping field - handle in decorator function.\n\n")

		return -1 // Skip field
	}

	return 0 // Not applicable
}

// BuildKnownTypeMapping handles known type conversions using BuildConversionCode.
// Returns true if a known conversion was found. Modifies mapping in place.
func BuildKnownTypeMapping(sourceField, targetField *protogen.Field, mapping *FieldMapping) bool {
	toCode, fromCode, convType, targetIsPtr := BuildConversionCode(sourceField, targetField)
	if toCode != "" && fromCode != "" {
		mapping.ToTargetCode = toCode
		mapping.FromTargetCode = fromCode
		mapping.ToTargetConversionType = convType
		mapping.FromTargetConversionType = convType
		if targetIsPtr != nil {
			mapping.TargetIsPointer = *targetIsPtr
		}
		return true
	}
	return false
}

// BuildSameTypeMapping handles same-type field conversions (direct assignment).
// Returns true if types match. Modifies mapping in place.
// Uses mapping.SourceIsOneofMember to determine correct field access (getter vs direct).
func BuildSameTypeMapping(sourceKind, targetKind, fieldName string, excludeMessages bool, mapping *FieldMapping) bool {
	if sourceKind != targetKind {
		return false
	}

	// Optionally exclude message types (GORM behavior)
	if excludeMessages && sourceKind == "message" {
		return false
	}

	// Use sourceFieldAccess to handle oneof member access correctly
	mapping.ToTargetCode = sourceFieldAccess(fieldName, mapping.SourceIsOneofMember)
	mapping.FromTargetCode = fmt.Sprintf("src.%s", fieldName)
	mapping.ToTargetConversionType = ConvertByAssignment
	mapping.FromTargetConversionType = ConvertByAssignment
	return true
}

// BuildNumericTypeMapping handles numeric type conversions using casting.
// Returns true if both types are numeric. Modifies mapping in place.
// Uses mapping.SourceIsOneofMember to determine correct field access (getter vs direct).
func BuildNumericTypeMapping(sourceKind, targetKind, fieldName string, mapping *FieldMapping) bool {
	if !common.IsNumericKind(sourceKind) || !common.IsNumericKind(targetKind) {
		return false
	}

	// Use sourceFieldAccess to handle oneof member access correctly
	srcAccess := sourceFieldAccess(fieldName, mapping.SourceIsOneofMember)
	mapping.ToTargetCode = fmt.Sprintf("%s(%s)", common.ProtoKindToGoType(targetKind), srcAccess)
	mapping.FromTargetCode = fmt.Sprintf("%s(src.%s)", common.ProtoKindToGoType(sourceKind), fieldName)
	mapping.ToTargetConversionType = ConvertByAssignment
	mapping.FromTargetConversionType = ConvertByAssignment
	return true
}

// RenderStrategyAdder is a function type for adding render strategies to a field mapping.
// This allows target-specific generators to customize the rendering logic.
type RenderStrategyAdder func(*FieldMapping)

// BuildFieldMapping builds a complete field mapping with type conversion logic.
// This is the unified field mapping function used by both GORM and Datastore generators.
//
// Parameters:
//   - sourceField: The protobuf field from the source/API message
//   - targetField: The protobuf field from the target (GORM/Datastore) message
//   - reg: Converter registry for looking up nested message converters
//   - msgRegistry: Message registry for resolving message type mappings
//   - sourcePkgName: Package name of the source message for template rendering
//   - addRenderStrategies: Function to add target-specific render strategies
//
// Returns nil if no conversion is possible (field should be skipped).
func BuildFieldMapping(
	sourceField, targetField *protogen.Field,
	reg *registry.ConverterRegistry,
	msgRegistry *common.MessageRegistry,
	sourcePkgName string,
	addRenderStrategies RenderStrategyAdder,
) *FieldMapping {
	sourceKind := sourceField.Desc.Kind().String()
	targetKind := targetField.Desc.Kind().String()
	fieldName := sourceField.GoName

	// Check if source field is part of a real oneof (not synthetic optional)
	// Real oneofs require getter method access (src.GetFieldName()) instead of direct field access
	sourceIsOneofMember := false
	if sourceField.Oneof != nil && !sourceField.Oneof.Desc.IsSynthetic() {
		sourceIsOneofMember = true
	}

	// Determine if fields are pointers
	// Source (proto): protoc-gen-go generates message fields as always pointers, optional scalars as pointers
	// For oneof members, pointer status depends on the actual type (messages are pointers, scalars are not)
	sourceIsPointer := sourceKind == "message" || sourceField.Desc.HasPresence()
	if sourceIsOneofMember {
		// Oneof scalar members (string, int, bool, etc.) are NOT pointers in proto-go
		// Only message types within oneofs are pointers
		sourceIsPointer = sourceKind == "message"
	}

	// Target: Only fields with explicit 'optional' keyword become pointers (message or scalar)
	// This gives us control over pointer vs value semantics in generated structs
	targetIsPointer := targetField.Desc.HasOptionalKeyword()

	mapping := &FieldMapping{
		SourceField:         sourceField.GoName,
		TargetField:         targetField.GoName,
		SourceIsPointer:     sourceIsPointer,
		TargetIsPointer:     targetIsPointer,
		SourceIsOneofMember: sourceIsOneofMember,
		SourcePkgName:       sourcePkgName,
	}

	// Mark if map or repeated (needed for later checks)
	mapping.IsMap = sourceField.Desc.IsMap()
	mapping.IsRepeated = sourceField.Desc.IsList()

	// Step 1: Check map fields (primitive maps return early)
	if BuildMapFieldMapping(MapFieldMappingParams{
		SourceField: sourceField,
		TargetField: targetField,
		Registry:    reg,
		MsgRegistry: msgRegistry,
		FieldName:   fieldName,
	}, mapping) {
		addRenderStrategies(mapping)
		return mapping
	}
	// If map with message values, set defaults and continue
	if mapping.IsMap {
		mapping.ToTargetConversionType = ConvertByTransformerWithError
		mapping.FromTargetConversionType = ConvertByTransformerWithError
	}

	// Step 2: Check repeated fields (primitive slices return early)
	if BuildRepeatedFieldMapping(RepeatedFieldMappingParams{
		SourceField: sourceField,
		TargetField: targetField,
		Registry:    reg,
		MsgRegistry: msgRegistry,
		FieldName:   fieldName,
	}, mapping) {
		addRenderStrategies(mapping)
		return mapping
	}
	// If repeated with message elements, set defaults and continue
	if mapping.IsRepeated {
		mapping.ToTargetConversionType = ConvertByTransformerWithError
		mapping.FromTargetConversionType = ConvertByTransformerWithError
	}

	// Set default conversion types for non-map, non-repeated fields
	if !mapping.IsMap && !mapping.IsRepeated {
		if sourceKind == "message" || targetKind == "message" {
			// Regular message fields default to transformer with error
			mapping.ToTargetConversionType = ConvertByTransformerWithError
			mapping.FromTargetConversionType = ConvertByTransformerWithError
		} else {
			// Scalar types default to assignment
			mapping.ToTargetConversionType = ConvertByAssignment
			mapping.FromTargetConversionType = ConvertByAssignment
		}
	}

	// Step 3: Check for custom converter functions (highest priority - overrides defaults)
	toTargetCode, fromTargetCode := common.ExtractCustomConverters(targetField, fieldName)
	if toTargetCode != "" {
		mapping.ToTargetCode = toTargetCode
		mapping.FromTargetCode = fromTargetCode
		mapping.ToTargetConversionType = ConvertByTransformer
		mapping.FromTargetConversionType = ConvertByTransformer
		addRenderStrategies(mapping)
		return mapping
	}

	// Step 4: Check for known type conversions using shared utility
	if BuildKnownTypeMapping(sourceField, targetField, mapping) {
		addRenderStrategies(mapping)
		return mapping
	}

	// Step 5: Check for same-type fields (excluding messages for consistency)
	if BuildSameTypeMapping(sourceKind, targetKind, fieldName, true, mapping) {
		addRenderStrategies(mapping)
		return mapping
	}

	// Step 6: Check for message→message conversions using shared utility
	msgStatus := BuildMessageToMessageMapping(MessageFieldMappingParams{
		SourceField: sourceField,
		TargetField: targetField,
		SourceKind:  sourceKind,
		TargetKind:  targetKind,
		Registry:    reg,
		MsgRegistry: msgRegistry,
		FieldName:   fieldName,
		IsRepeated:  mapping.IsRepeated,
		IsMap:       mapping.IsMap,
	}, mapping)
	if msgStatus == -1 {
		// No converter available - already logged warning - skip field
		return nil
	} else if msgStatus == 1 {
		// Conversion found
		addRenderStrategies(mapping)
		return mapping
	}

	// Step 7: Check for numeric type conversions using shared utility
	if BuildNumericTypeMapping(sourceKind, targetKind, fieldName, mapping) {
		addRenderStrategies(mapping)
		return mapping
	}

	// No built-in conversion available - log warning and skip
	log.Printf("WARNING: No type conversion found for field %q: %s (%s) → %s (%s).",
		fieldName,
		GetTypeName(sourceField), sourceKind,
		GetTypeName(targetField), targetKind)
	log.Printf("         Field will be skipped in converter - handle in decorator function.")
	return nil
}
