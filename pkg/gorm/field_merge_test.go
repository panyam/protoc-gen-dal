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

package gorm

import (
	"testing"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"google.golang.org/protobuf/compiler/protogen"
)

// TestFieldMerging tests that fields from source message are automatically merged
// into target message when target has a source annotation.
//
// Test cases:
// 1. Empty target → should get ALL source fields
// 2. Partial target → should get ALL source fields with some overridden
// 3. Skip field → source field with skip_field should be excluded
// 4. Extra fields → target can add fields not in source
func TestFieldMerging(t *testing.T) {
	// This test will fail until we implement field merging
	// Expected behavior after implementation:
	//
	// DocumentGormEmpty:
	//   - Should have 9 fields from api.Document (all auto-merged)
	//
	// DocumentGormPartial:
	//   - Should have 9 fields total
	//   - id and title customized with GORM tags
	//   - Other 7 fields auto-merged from source
	//
	// DocumentGormSkip:
	//   - Should have 8 fields (9 from source - 1 skipped)
	//   - content field should be excluded (has skip_field = true)
	//
	// DocumentGormExtra:
	//   - Should have 11 fields (9 from source + 2 extra)
	//   - deleted_at and version are database-specific additions

	t.Skip("Field merging not yet implemented - this test documents expected behavior")

	// TODO: After implementing field merging, this test should:
	// 1. Load the generated document_gorm.go file
	// 2. Parse the struct definitions
	// 3. Assert field counts and field names match expectations above
}

// TestInvalidSourceReference tests that we error when source message doesn't exist
func TestInvalidSourceReference(t *testing.T) {
	t.Skip("Validation not yet implemented")

	// TODO: Test that referencing non-existent source message produces error
	// Example: source: "api.NonExistentMessage"
}

// TestInvalidSkipField tests that we error when skip_field references invalid field
func TestInvalidSkipField(t *testing.T) {
	t.Skip("Validation not yet implemented")

	// TODO: Test that skip_field on field not in source produces error
	// Example: source message has fields [a, b, c] but skip_field applied to field 'd'
}

// mergeSourceFields merges source message fields into target message fields.
//
// Algorithm:
// 1. Start with all fields from source message
// 2. For each target field:
//    - If field number matches source → override (allows customization)
//    - If field number is new → add (allows extra fields)
// 3. Filter out fields marked with skip_field
//
// This will be the core implementation function.
func mergeSourceFields(sourceMsg, targetMsg *protogen.Message) []*protogen.Field {
	// TODO: Implement this function
	// For now, just return target fields (current behavior)
	return targetMsg.Fields
}

// hasSkipField checks if a field has the skip_field annotation set to true
func hasSkipField(field *protogen.Field) bool {
	// TODO: Implement this by checking field options for dal.v1.skip_field
	return false
}

// validateSourceMessage checks that the source message reference is valid
func validateSourceMessage(info *collector.MessageInfo) error {
	// TODO: Implement validation
	// - Check that SourceMessage is not nil
	// - Check that SourceName matches an actual message
	return nil
}

// validateSkipFields checks that all skip_field annotations reference valid source fields
func validateSkipFields(sourceMsg, targetMsg *protogen.Message) error {
	// TODO: Implement validation
	// - For each target field with skip_field = true
	// - Verify that field exists in source message
	return nil
}
