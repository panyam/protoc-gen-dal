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
	"sort"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// MergeSourceFields merges source message fields into target message fields.
//
// This implements the opt-out field model with oneof-aware handling:
// 1. Start with all fields from source message (if source exists)
// 2. Detect oneof replacement: if target has field matching source oneof name,
//    automatically skip all oneof members
// 3. For each target field:
//    - If field name matches source field → override (allows customization)
//    - If field name is new → add (allows extra fields)
// 4. Filter out fields marked with skip_field = true
//
// Oneof Handling:
// - If target field name matches source oneof name: auto-skip ALL oneof members
// - Otherwise: oneof members are merged normally (can be overridden or skipped individually)
//
// Parameters:
//   - sourceMsg: The source API message (can be nil if no source annotation)
//   - targetMsg: The target GORM/Datastore message
//
// Returns:
//   - Merged list of fields to generate
//   - Error if oneof handling is invalid (future enhancement)
func MergeSourceFields(sourceMsg, targetMsg *protogen.Message) ([]*protogen.Field, error) {
	// If no source message, just use target fields as-is
	if sourceMsg == nil {
		return targetMsg.Fields, nil
	}

	// Build a map of target field names that replace oneofs
	targetReplacesOneof := make(map[string]bool)
	for _, targetField := range targetMsg.Fields {
		targetFieldName := string(targetField.Desc.Name())

		// Check if this target field name matches any source oneof name
		for _, oneof := range sourceMsg.Oneofs {
			if string(oneof.Desc.Name()) == targetFieldName {
				targetReplacesOneof[targetFieldName] = true
				break
			}
		}
	}

	// Build a map of source fields by field name
	sourceFieldsByName := make(map[string]*protogen.Field)
	for _, field := range sourceMsg.Fields {
		fieldName := string(field.Desc.Name())

		// Skip oneof members if target has a replacement field
		if field.Oneof != nil {
			oneofName := string(field.Oneof.Desc.Name())
			if targetReplacesOneof[oneofName] {
				// Auto-skip this oneof member
				continue
			}
		}

		sourceFieldsByName[fieldName] = field
	}

	// Build result: start with source fields, then apply overrides/additions
	resultByName := make(map[string]*protogen.Field)
	// Track source field numbers for fields that get overridden
	// This ensures we maintain source ordering even when target uses different field numbers
	sourceFieldNumbers := make(map[string]int32)

	// Copy all source fields to result and track their numbers
	for name, field := range sourceFieldsByName {
		resultByName[name] = field
		sourceFieldNumbers[name] = int32(field.Desc.Number())
	}

	// Process target fields: override source or add new
	for _, targetField := range targetMsg.Fields {
		fieldName := string(targetField.Desc.Name())

		// Check if this field has skip_field = true
		if HasSkipField(targetField) {
			// Remove from result if it exists
			delete(resultByName, fieldName)
			delete(sourceFieldNumbers, fieldName)
			continue
		}

		// Override source field or add new field
		resultByName[fieldName] = targetField
		// If this is a new field (not in source), use target's field number
		if _, existsInSource := sourceFieldsByName[fieldName]; !existsInSource {
			sourceFieldNumbers[fieldName] = int32(targetField.Desc.Number())
		}
		// Otherwise keep the source field number for sorting (already set above)
	}

	// Convert map back to slice
	var result []*protogen.Field
	for _, field := range resultByName {
		result = append(result, field)
	}

	// Sort by SOURCE field numbers to maintain source ordering
	// This ensures fields appear in the same order as the source proto,
	// even when target overrides them with different field numbers
	sort.Slice(result, func(i, j int) bool {
		iName := string(result[i].Desc.Name())
		jName := string(result[j].Desc.Name())
		return sourceFieldNumbers[iName] < sourceFieldNumbers[jName]
	})

	return result, nil
}

// HasSkipField checks if a field has the skip_field annotation set to true.
//
// Parameters:
//   - field: The proto field to check
//
// Returns:
//   - true if field has skip_field = true annotation
func HasSkipField(field *protogen.Field) bool {
	opts := field.Desc.Options()
	if opts == nil {
		return false
	}

	// Get skip_field extension
	v := proto.GetExtension(opts, dalv1.E_SkipField)
	if v == nil {
		return false
	}

	skipField, ok := v.(bool)
	return ok && skipField
}

// ValidateFieldMerge validates that field merging is correct.
//
// Checks:
// 1. If target field has skip_field = true, it must exist in source
// 2. Source message reference must be valid
//
// Parameters:
//   - sourceMsg: The source message (can be nil)
//   - targetMsg: The target message
//   - sourceName: The source message name from annotation (for error messages)
//
// Returns:
//   - error if validation fails
func ValidateFieldMerge(sourceMsg *protogen.Message, targetMsg *protogen.Message, sourceName string) error {
	// If no source specified, nothing to validate
	if sourceMsg == nil {
		// But if sourceName is specified, that's an error - source not found
		if sourceName != "" {
			return fmt.Errorf("source message %q not found for target %q", sourceName, targetMsg.Desc.Name())
		}
		return nil
	}

	// Build map of source field names
	sourceFieldNames := make(map[string]bool)
	for _, field := range sourceMsg.Fields {
		sourceFieldNames[string(field.Desc.Name())] = true
	}

	// Validate skip_field annotations
	for _, targetField := range targetMsg.Fields {
		if HasSkipField(targetField) {
			fieldName := string(targetField.Desc.Name())
			if !sourceFieldNames[fieldName] {
				return fmt.Errorf(
					"field %q in target %q has skip_field=true but does not exist in source %q",
					targetField.Desc.Name(),
					targetMsg.Desc.Name(),
					sourceName,
				)
			}
		}
	}

	return nil
}
