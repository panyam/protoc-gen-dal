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

	// Collection characteristics
	IsRepeated bool // Whether this is a repeated field (needs loop-based conversion)
	IsMap      bool // Whether this is a map field (needs loop-based conversion)

	// Element types for repeated/map fields
	TargetElementType string // For repeated/map: Go type of target element/value (e.g., "AuthorGORM")
	SourceElementType string // For repeated/map: Go type of source element/value (e.g., "Author")

	// Package information
	SourcePkgName string // Source package name (e.g., "api" or "testapi") - needed for type references
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
	valueField := mapEntry.Fields[1] // value field is always index 1
	valueKind := valueField.Desc.Kind().String()

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
	mapping.ToTargetCode = fmt.Sprintf("src.%s", params.FieldName)
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
	mapping.ToTargetCode = fmt.Sprintf("src.%s", params.FieldName)
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
func BuildSameTypeMapping(sourceKind, targetKind, fieldName string, excludeMessages bool, mapping *FieldMapping) bool {
	if sourceKind != targetKind {
		return false
	}

	// Optionally exclude message types (GORM behavior)
	if excludeMessages && sourceKind == "message" {
		return false
	}

	mapping.ToTargetCode = fmt.Sprintf("src.%s", fieldName)
	mapping.FromTargetCode = fmt.Sprintf("src.%s", fieldName)
	mapping.ToTargetConversionType = ConvertByAssignment
	mapping.FromTargetConversionType = ConvertByAssignment
	return true
}

// BuildNumericTypeMapping handles numeric type conversions using casting.
// Returns true if both types are numeric. Modifies mapping in place.
func BuildNumericTypeMapping(sourceKind, targetKind, fieldName string, mapping *FieldMapping) bool {
	if !common.IsNumericKind(sourceKind) || !common.IsNumericKind(targetKind) {
		return false
	}

	mapping.ToTargetCode = fmt.Sprintf("%s(src.%s)", common.ProtoKindToGoType(targetKind), fieldName)
	mapping.FromTargetCode = fmt.Sprintf("%s(src.%s)", common.ProtoKindToGoType(sourceKind), fieldName)
	mapping.ToTargetConversionType = ConvertByAssignment
	mapping.FromTargetConversionType = ConvertByAssignment
	return true
}
