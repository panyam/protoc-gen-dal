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
	"testing"

	"github.com/panyam/protoc-gen-dal/pkg/generator/testutil"
	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
	"google.golang.org/protobuf/compiler/protogen"
)

// TestMergeSourceFields_FieldOrdering tests that merged fields maintain source field order
// even when target fields override with different field numbers.
//
// This reproduces the issue where UserWithPermissions has created_at=4 and updated_at=5
// but they should appear in positions 8 and 9 (matching their source field numbers).
func TestMergeSourceFields_FieldOrdering(t *testing.T) {
	protoSet := &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "api/user.proto",
				Pkg:  "api",
				Messages: []testutil.TestMessage{
					{
						Name: "User",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "uint32"},
							{Name: "name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
							{Name: "age", Number: 4, TypeName: "uint32"},
							{Name: "birthday", Number: 5, TypeName: "string"},
							{Name: "member_number", Number: 6, TypeName: "string"},
							{Name: "activated_at", Number: 7, TypeName: "string"},
							{Name: "created_at", Number: 8, TypeName: "string"},
							{Name: "updated_at", Number: 9, TypeName: "string"},
						},
					},
				},
			},
			{
				Name: "gorm/user.proto",
				Pkg:  "gorm",
				Messages: []testutil.TestMessage{
					{
						Name: "UserWithPermissions",
						GormOpts: &dalv1.GormOptions{
							Source: "api.User",
							Table:  "users_with_perms",
						},
						Fields: []testutil.TestField{
							// Override first 3 fields with same numbers
							{Name: "id", Number: 1, TypeName: "uint32"},
							{Name: "name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
							// Override created_at and updated_at with DIFFERENT numbers
							// This is the key issue: these have lower numbers than their source positions
							{Name: "created_at", Number: 4, TypeName: "string"},
							{Name: "updated_at", Number: 5, TypeName: "string"},
						},
					},
				},
			},
		},
	}

	plugin := testutil.CreateTestPlugin(t, protoSet)

	// Find the messages
	var sourceMsg, targetMsg *protogen.Message
	for _, file := range plugin.Files {
		for _, msg := range file.Messages {
			if msg.Desc.Name() == "User" {
				sourceMsg = msg
			}
			if msg.Desc.Name() == "UserWithPermissions" {
				targetMsg = msg
			}
		}
	}

	if sourceMsg == nil || targetMsg == nil {
		t.Fatal("Could not find test messages")
	}

	// Merge fields
	merged, err := MergeSourceFields(sourceMsg, targetMsg)
	if err != nil {
		t.Fatalf("MergeSourceFields failed: %v", err)
	}

	// Expected order based on SOURCE field numbers:
	// 1. id (1)
	// 2. name (2)
	// 3. email (3)
	// 4. age (4) - from source
	// 5. birthday (5) - from source
	// 6. member_number (6) - from source
	// 7. activated_at (7) - from source
	// 8. created_at (8 from source, overridden by target field 4)
	// 9. updated_at (9 from source, overridden by target field 5)
	expectedOrder := []string{
		"id",            // 1
		"name",          // 2
		"email",         // 3
		"age",           // 4
		"birthday",      // 5
		"member_number", // 6
		"activated_at",  // 7
		"created_at",    // 8
		"updated_at",    // 9
	}

	if len(merged) != len(expectedOrder) {
		t.Fatalf("Expected %d fields, got %d", len(expectedOrder), len(merged))
	}

	for i, field := range merged {
		actualName := string(field.Desc.Name())
		expectedName := expectedOrder[i]
		if actualName != expectedName {
			t.Errorf("Field at position %d: expected %q, got %q", i, expectedName, actualName)
		}
	}
}
