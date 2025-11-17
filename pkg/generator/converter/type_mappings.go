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

	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"google.golang.org/protobuf/compiler/protogen"
)

// TypeMapping defines a conversion between two types.
type TypeMapping struct {
	// ToTargetTemplate is the Go expression template for source→target conversion
	// Use {{.SourceField}} placeholder for field name
	ToTargetTemplate string

	// FromTargetTemplate is the Go expression template for target→source conversion
	// Use {{.TargetField}} placeholder for field name
	FromTargetTemplate string

	// ConversionType indicates how to apply the conversion
	ConversionType ConversionType

	// TargetIsPointer overrides the pointer detection for target field
	// Set to false for value types like time.Time
	TargetIsPointer *bool
}

// TypePair uniquely identifies a source→target type conversion.
type TypePair struct {
	SourceType string // Fully qualified proto type or proto kind (e.g., "google.protobuf.Timestamp", "int64")
	TargetType string
}

// globalTypeMappings defines all known type conversions.
var globalTypeMappings = map[TypePair]TypeMapping{
	// google.protobuf.Timestamp → int64 (Unix epoch seconds)
	{SourceType: "google.protobuf.Timestamp", TargetType: "int64"}: {
		ToTargetTemplate:   "converters.TimestampToInt64(src.{{.SourceField}})",
		FromTargetTemplate: "converters.Int64ToTimestamp(src.{{.TargetField}})",
		ConversionType:     ConvertByTransformer,
	},

	// google.protobuf.Timestamp → google.protobuf.Timestamp (maps to time.Time in Go)
	{SourceType: "google.protobuf.Timestamp", TargetType: "google.protobuf.Timestamp"}: {
		ToTargetTemplate:   "converters.TimestampToTime(src.{{.SourceField}})",
		FromTargetTemplate: "converters.TimeToTimestamp(src.{{.TargetField}})",
		ConversionType:     ConvertByTransformer,
		TargetIsPointer:    boolPtr(false), // time.Time is a value type
	},

	// google.protobuf.Any → bytes (binary serialization)
	{SourceType: "google.protobuf.Any", TargetType: "bytes"}: {
		ToTargetTemplate:   "converters.AnyToBytes(src.{{.SourceField}})",
		FromTargetTemplate: "converters.BytesToAny(src.{{.TargetField}})",
		ConversionType:     ConvertByTransformerWithError,
	},

	// uint32 → string (for ID conversions)
	{SourceType: "uint32", TargetType: "string"}: {
		ToTargetTemplate:   "strconv.FormatUint(uint64(src.{{.SourceField}}), 10)",
		FromTargetTemplate: "uint32(converters.MustParseUint(src.{{.TargetField}}))",
		ConversionType:     ConvertByTransformer,
	},
}

// GetTypeMapping finds a type mapping for the given source and target fields.
// Returns nil if no mapping exists.
func GetTypeMapping(sourceField, targetField *protogen.Field) *TypeMapping {
	sourceType := getSourceTypeKey(sourceField)
	targetType := getTargetTypeKey(targetField)

	pair := TypePair{SourceType: sourceType, TargetType: targetType}
	if mapping, exists := globalTypeMappings[pair]; exists {
		return &mapping
	}

	return nil
}

// getSourceTypeKey returns the type key for a source field (from proto-generated code).
// For messages, returns the fully qualified name (e.g., "google.protobuf.Timestamp").
// For primitives, returns the proto kind (e.g., "int64", "string").
// Source fields use the proto-generated Go types (e.g., *anypb.Any), not our custom mappings.
func getSourceTypeKey(field *protogen.Field) string {
	kind := field.Desc.Kind().String()

	if kind == "message" && field.Message != nil {
		return string(field.Message.Desc.FullName())
	}

	return kind
}

// getTargetTypeKey returns the type key for a target field (in our generated code).
// For messages, returns the fully qualified name or mapped type.
// For primitives, returns the proto kind.
// Target fields may use well-known type mappings (e.g., google.protobuf.Any → []byte).
func getTargetTypeKey(field *protogen.Field) string {
	kind := field.Desc.Kind().String()

	if kind == "message" && field.Message != nil {
		protoFullName := string(field.Message.Desc.FullName())

		// Check if this is a well-known type that gets mapped to a different Go type
		// If so, return the type key for the Go type instead of the proto type
		// For example: google.protobuf.Any → []byte in our generated code → "bytes" type key
		if wellKnownMapping, exists := common.GetWellKnownTypeMapping(field.Message); exists {
			// Map Go type back to proto kind for type matching
			// []byte → "bytes", time.Time → special case (keep as proto name for compatibility)
			if wellKnownMapping.GoType == "[]byte" {
				return "bytes"
			}
			// For types like time.Time, keep using the proto full name
			// This allows Timestamp→Timestamp mappings to work
			return protoFullName
		}

		return protoFullName
	}

	return kind
}

// ApplyTypeMapping applies a type mapping to generate conversion code.
func ApplyTypeMapping(mapping *TypeMapping, sourceFieldName, targetFieldName string) (toCode, fromCode string) {
	// Simple template replacement
	toCode = replaceTemplate(mapping.ToTargetTemplate, sourceFieldName, targetFieldName)
	fromCode = replaceTemplate(mapping.FromTargetTemplate, sourceFieldName, targetFieldName)
	return toCode, fromCode
}

// replaceTemplate performs simple {{.SourceField}} and {{.TargetField}} replacement.
func replaceTemplate(template, sourceField, targetField string) string {
	// Simple string replacement for now
	result := template
	result = replaceAll(result, "{{.SourceField}}", sourceField)
	result = replaceAll(result, "{{.TargetField}}", targetField)
	return result
}

// replaceAll is a helper for string replacement.
func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

// BuildConversionCode builds conversion code using type mappings.
// Returns empty strings if no mapping found.
func BuildConversionCode(sourceField, targetField *protogen.Field) (toCode, fromCode string, convType ConversionType, targetIsPtr *bool) {
	mapping := GetTypeMapping(sourceField, targetField)
	if mapping == nil {
		return "", "", ConvertIgnore, nil
	}

	sourceFieldName := sourceField.GoName
	targetFieldName := targetField.GoName

	toCode, fromCode = ApplyTypeMapping(mapping, sourceFieldName, targetFieldName)
	return toCode, fromCode, mapping.ConversionType, mapping.TargetIsPointer
}

// RegisterTypeMapping allows registration of custom type mappings.
// This is useful for plugins or extensions that need custom conversions.
func RegisterTypeMapping(sourceType, targetType string, mapping TypeMapping) {
	pair := TypePair{SourceType: sourceType, TargetType: targetType}
	globalTypeMappings[pair] = mapping
}

// GetTypeName returns a human-readable type name for error messages.
// Uses source type key (proto full name).
func GetTypeName(field *protogen.Field) string {
	return getSourceTypeKey(field)
}

// Debug helper to show what mapping was found
func DebugTypeMapping(sourceField, targetField *protogen.Field) string {
	sourceType := getSourceTypeKey(sourceField)
	targetType := getTargetTypeKey(targetField)
	pair := TypePair{SourceType: sourceType, TargetType: targetType}

	if mapping, exists := globalTypeMappings[pair]; exists {
		return fmt.Sprintf("Found mapping: %s → %s (to: %s, from: %s)",
			sourceType, targetType, mapping.ToTargetTemplate, mapping.FromTargetTemplate)
	}

	return fmt.Sprintf("No mapping found for: %s → %s", sourceType, targetType)
}
