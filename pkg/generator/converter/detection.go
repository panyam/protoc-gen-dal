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

// IsTimestampToInt64 checks if this is a google.protobuf.Timestamp → int64 conversion.
//
// This is a common built-in conversion where protobuf Timestamp messages are stored
// as Unix epoch seconds (int64) in the database.
//
// Parameters:
//   - sourceField: source field descriptor
//   - targetField: target field descriptor
//
// Returns:
//   - true if source is google.protobuf.Timestamp and target is int64
func IsTimestampToInt64(sourceField, targetField *protogen.Field) bool {
	sourceKind := sourceField.Desc.Kind().String()
	targetKind := targetField.Desc.Kind().String()

	if sourceKind != "message" || targetKind != "int64" {
		return false
	}

	if sourceField.Message == nil {
		return false
	}

	return string(sourceField.Message.Desc.FullName()) == "google.protobuf.Timestamp"
}

// IsNumericConversion checks if both source and target are numeric types.
//
// Numeric conversions require explicit type casting in Go (e.g., int32(value)).
// This delegates to common.IsNumericKind for the actual numeric type checking.
//
// Parameters:
//   - sourceKind: proto kind string (e.g., "int32", "uint64")
//   - targetKind: proto kind string
//
// Returns:
//   - true if both are numeric types
func IsNumericConversion(sourceKind, targetKind string) bool {
	return common.IsNumericKind(sourceKind) && common.IsNumericKind(targetKind)
}

// IsSameScalarType checks if source and target have the same scalar type.
//
// When types match and are not messages, conversion is a simple assignment.
//
// Parameters:
//   - sourceKind: proto kind string
//   - targetKind: proto kind string
//
// Returns:
//   - true if both are the same scalar (non-message) type
func IsSameScalarType(sourceKind, targetKind string) bool {
	return sourceKind == targetKind && sourceKind != "message"
}

// BuildNumericCast generates a Go type cast expression for numeric conversions.
//
// Generates code like: "int64(src.FieldName)" or "uint32(src.FieldName)".
//
// Parameters:
//   - sourceFieldName: name of the source field (e.g., "Count")
//   - targetKind: target proto kind (e.g., "int64")
//
// Returns:
//   - Go code expression for the type cast
func BuildNumericCast(sourceFieldName, targetKind string) string {
	targetGoType := common.ProtoKindToGoType(targetKind)
	return fmt.Sprintf("%s(src.%s)", targetGoType, sourceFieldName)
}

// BuildTimestampConversion generates conversion code for Timestamp ↔ int64.
//
// Parameters:
//   - sourceFieldName: name of the field to convert
//   - toInt64: true for Timestamp→int64, false for int64→Timestamp
//
// Returns:
//   - toCode: conversion expression for source→target
//   - fromCode: conversion expression for target→source
func BuildTimestampConversion(sourceFieldName string) (toCode, fromCode string) {
	toCode = fmt.Sprintf("timestampToInt64(src.%s)", sourceFieldName)
	fromCode = fmt.Sprintf("int64ToTimestamp(src.%s)", sourceFieldName)
	return toCode, fromCode
}
