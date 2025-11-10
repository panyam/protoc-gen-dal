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

package common

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

// ProtoScalarToGo maps proto scalar types to their Go equivalents.
//
// This function handles the conversion of proto primitive types to Go types.
// It's used for scalar fields, map keys, and array elements.
//
// Parameters:
//   - protoType: The proto kind as a string (e.g., "string", "int32", "bool")
//
// Returns:
//   - the corresponding Go type (e.g., "string", "int32", "bool")
//   - "interface{}" for unknown types (fallback)
func ProtoScalarToGo(protoType string) string {
	switch protoType {
	case "string":
		return "string"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "bool":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "bytes":
		return "[]byte"
	default:
		return "interface{}" // Fallback for unknown types
	}
}

// StructNameFunc is a function that converts a proto message to a struct name.
// Different targets may use different naming conventions:
//   - GORM: "UserGorm" -> "UserGORM"
//   - Datastore: "UserDatastore" -> "UserDatastore" (no change)
type StructNameFunc func(*protogen.Message) string

// ProtoFieldToGoType converts a proto field to its Go type representation.
//
// This function handles all proto field types including scalars, messages,
// repeated fields, and maps. It correctly generates Go map types for proto maps
// instead of using entry struct types.
//
// How it handles different field types:
//   - Scalars: "string", "int32", etc.
//   - Messages: Uses structNameFunc to get the target struct name
//   - Repeated scalars: "[]string", "[]int32", etc.
//   - Repeated messages: "[]BookGORM", "[]AuthorDatastore", etc.
//   - Maps with scalar values: "map[string]int32", "map[string]string", etc.
//   - Maps with message values: "map[string]BookGORM", "map[uint32]AuthorDatastore", etc.
//
// Parameters:
//   - field: The proto field to convert
//   - structNameFunc: Function to convert message names to struct names
//
// Returns:
//   - the Go type string for the field
func ProtoFieldToGoType(field *protogen.Field, structNameFunc StructNameFunc) string {
	kind := field.Desc.Kind().String()

	// Handle map fields - proto represents maps as special message types
	// We want to generate native Go maps: map[K]V
	// NOT entry struct types like []BookEntry{Key, Value}
	if field.Desc.IsMap() {
		// Extract key and value types from the map entry message
		mapEntry := field.Message
		keyField := mapEntry.Fields[0]   // maps always have key at index 0
		valueField := mapEntry.Fields[1] // maps always have value at index 1

		keyType := ProtoScalarToGo(keyField.Desc.Kind().String())

		// Check if value is a message type or scalar
		var valueType string
		if valueField.Desc.Kind().String() == "message" {
			// Map value is a message type - use the target struct name
			valueType = structNameFunc(valueField.Message)
		} else {
			// Map value is a scalar type
			valueType = ProtoScalarToGo(valueField.Desc.Kind().String())
		}

		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	}

	// Handle message types (embedded structs, nested objects, etc.)
	if kind == "message" {
		// For repeated message fields: []BookGORM, []AuthorDatastore
		if field.Desc.Cardinality().String() == "repeated" {
			return "[]" + structNameFunc(field.Message)
		}
		// For singular message fields: BookGORM, AuthorDatastore
		return structNameFunc(field.Message)
	}

	// Handle repeated scalar fields: []string, []int32, etc.
	if field.Desc.Cardinality().String() == "repeated" {
		elemType := ProtoScalarToGo(kind)
		return "[]" + elemType
	}

	// Regular scalar field: string, int32, bool, etc.
	return ProtoScalarToGo(kind)
}

// IsNumericKind checks if a proto kind represents a numeric type.
//
// Numeric types can be safely cast between each other in generated code,
// though precision may be lost in some conversions.
//
// Parameters:
//   - kind: The proto kind as a string
//
// Returns:
//   - true if the kind is numeric, false otherwise
func IsNumericKind(kind string) bool {
	numericKinds := map[string]bool{
		"int32":    true,
		"int64":    true,
		"uint32":   true,
		"uint64":   true,
		"sint32":   true,
		"sint64":   true,
		"fixed32":  true,
		"fixed64":  true,
		"sfixed32": true,
		"sfixed64": true,
		"float":    true,
		"double":   true,
	}
	return numericKinds[kind]
}

// ProtoKindToGoType converts a proto kind to its Go type for casting.
//
// This is useful when generating type conversion code where explicit casts
// are needed (e.g., converting between different numeric types).
//
// Parameters:
//   - kind: The proto kind as a string
//
// Returns:
//   - the Go type suitable for type casting
func ProtoKindToGoType(kind string) string {
	switch kind {
	case "int32", "sint32", "sfixed32":
		return "int32"
	case "int64", "sint64", "sfixed64":
		return "int64"
	case "uint32", "fixed32":
		return "uint32"
	case "uint64", "fixed64":
		return "uint64"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return kind
	}
}
