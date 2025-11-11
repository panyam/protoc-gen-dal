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

	"google.golang.org/protobuf/compiler/protogen"
)

// CheckMapValueType determines if a map field contains primitive or message values.
//
// This implements the "applicative" pattern for map fields: examine the contained
// type to decide conversion strategy rather than treating all maps the same.
//
// Parameters:
//   - field: Proto field descriptor (must be a map field)
//
// Returns:
//   - isPrimitive: true if map value is a scalar type, false if message type
//   - valueMsg: message descriptor if value is a message, nil if primitive
func CheckMapValueType(field *protogen.Field) (isPrimitive bool, valueMsg *protogen.Message) {
	if !field.Desc.IsMap() {
		return false, nil
	}

	// Map fields have an entry message with key at index 0, value at index 1
	mapEntry := field.Message
	if mapEntry == nil || len(mapEntry.Fields) < 2 {
		return false, nil
	}

	valueField := mapEntry.Fields[1]
	valueKind := valueField.Desc.Kind().String()

	if valueKind == "message" {
		return false, valueField.Message
	}

	return true, nil
}

// CheckRepeatedElementType determines if a repeated field contains primitive or message elements.
//
// This implements the "applicative" pattern for repeated fields: examine the contained
// type to decide conversion strategy rather than treating all slices the same.
//
// Parameters:
//   - field: Proto field descriptor (must be a repeated field)
//
// Returns:
//   - isPrimitive: true if element is a scalar type, false if message type
//   - elemMsg: message descriptor if element is a message, nil if primitive
func CheckRepeatedElementType(field *protogen.Field) (isPrimitive bool, elemMsg *protogen.Message) {
	if !field.Desc.IsList() {
		return false, nil
	}

	elementKind := field.Desc.Kind().String()

	if elementKind == "message" {
		return false, field.Message
	}

	return true, nil
}

// BuildNestedConverterName generates To/From converter function names for nested types.
//
// Uses the standard naming convention: {SourceType}To{TargetType} and {SourceType}From{TargetType}.
// This prevents naming collisions when source and target have the same base name.
//
// Parameters:
//   - sourceType: source message name (e.g., "Author")
//   - targetType: target message name (e.g., "AuthorGORM" or "AuthorDatastore")
//
// Returns:
//   - toFunc: function name for source→target conversion (e.g., "AuthorToAuthorGORM")
//   - fromFunc: function name for target→source conversion (e.g., "AuthorFromAuthorGORM")
func BuildNestedConverterName(sourceType, targetType string) (toFunc, fromFunc string) {
	toFunc = fmt.Sprintf("%sTo%s", sourceType, targetType)
	fromFunc = fmt.Sprintf("%sFrom%s", sourceType, targetType)
	return toFunc, fromFunc
}

// ExtractMapMessages extracts source and target message descriptors from map fields.
//
// For map<K, MessageType> fields, this returns the message descriptors for the
// value types in both source and target maps.
//
// Parameters:
//   - sourceField: source map field descriptor
//   - targetField: target map field descriptor
//
// Returns:
//   - sourceMsg: message descriptor for source map value type, nil if not a message
//   - targetMsg: message descriptor for target map value type, nil if not a message
func ExtractMapMessages(sourceField, targetField *protogen.Field) (sourceMsg, targetMsg *protogen.Message) {
	if !sourceField.Desc.IsMap() || !targetField.Desc.IsMap() {
		return nil, nil
	}

	sourceMapEntry := sourceField.Message
	targetMapEntry := targetField.Message

	if sourceMapEntry == nil || targetMapEntry == nil {
		return nil, nil
	}

	if len(sourceMapEntry.Fields) < 2 || len(targetMapEntry.Fields) < 2 {
		return nil, nil
	}

	sourceValueField := sourceMapEntry.Fields[1]
	targetValueField := targetMapEntry.Fields[1]

	return sourceValueField.Message, targetValueField.Message
}

// ExtractRepeatedMessages extracts source and target message descriptors from repeated fields.
//
// For []MessageType fields, this returns the message descriptors for the
// element types in both source and target slices.
//
// Parameters:
//   - sourceField: source repeated field descriptor
//   - targetField: target repeated field descriptor
//
// Returns:
//   - sourceMsg: message descriptor for source element type, nil if not a message
//   - targetMsg: message descriptor for target element type, nil if not a message
func ExtractRepeatedMessages(sourceField, targetField *protogen.Field) (sourceMsg, targetMsg *protogen.Message) {
	if !sourceField.Desc.IsList() || !targetField.Desc.IsList() {
		return nil, nil
	}

	return sourceField.Message, targetField.Message
}
