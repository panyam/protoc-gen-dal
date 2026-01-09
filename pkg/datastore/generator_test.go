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

package datastore

import (
	"strings"
	"testing"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/testutil"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// TestGenerateDatastore_SimpleMessage tests that a simple message generates
// a correct Datastore entity struct with basic fields
//
// This is the first test that drives the Datastore generator implementation.
// It verifies that we can generate:
// - A properly named struct (UserDatastore)
// - Basic fields with correct datastore tags
// - A Kind() method
func TestGenerateDatastore_SimpleMessage(t *testing.T) {
	// Given: A simple User message with basic fields
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			// API proto
			{
				Name: "api/v1/user.proto",
				Pkg:  "api.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "User",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
						},
					},
				},
			},
			// Datastore DAL proto
			{
				Name: "dal/v1/user_datastore.proto",
				Pkg:  "dal.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source:    "api.v1.User",
							Kind:      "User",
							Namespace: "prod",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	// Collect Datastore messages
	messages, err := collector.CollectMessages(plugin, collector.TargetDatastore)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("Expected 1 Datastore message, got %d", len(messages))
	}

	// When: Generate Datastore code
	result, err := Generate(messages)

	// Then: Should succeed
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Then: Should generate at least one file
	if len(result.Files) == 0 {
		t.Fatal("Expected at least one generated file")
	}

	// Then: Check generated content
	content := result.Files[0].Content

	// Should have struct definition
	if !strings.Contains(content, "type UserDatastore struct") {
		t.Error("Expected struct definition for UserDatastore")
	}

	// Should have fields with datastore tags
	if !strings.Contains(content, "ID") && !strings.Contains(content, "Id") {
		t.Error("Expected ID field")
	}
	if !strings.Contains(content, "Name") {
		t.Error("Expected Name field")
	}
	if !strings.Contains(content, "Email") {
		t.Error("Expected Email field")
	}

	// Should have Kind() method
	if !strings.Contains(content, "func (*UserDatastore) Kind()") {
		t.Error("Expected Kind() method")
	}
	if !strings.Contains(content, `return "User"`) {
		t.Error("Expected Kind() to return 'User'")
	}

	// Should have Key field (excluded from datastore properties)
	if !strings.Contains(content, "Key") {
		t.Error("Expected Key field for datastore key management")
	}
	if !strings.Contains(content, "`datastore:\"-\"`") {
		t.Error("Expected Key field to be excluded from datastore properties")
	}
}

// TestGenerateConverters tests that converter functions are generated
// for converting between API messages and Datastore entities.
//
// This test verifies:
// - ToDatastore converter function with decorator parameter
// - FromDatastore converter function with decorator parameter
// - Reuses GORM converter infrastructure (built-in types, decorators)
func TestGenerateConverters(t *testing.T) {
	// Given: A User message with Datastore mapping
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			// API proto
			{
				Name: "api/v1/user.proto",
				Pkg:  "api.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "User",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "Name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
						},
					},
				},
			},
			// Datastore DAL proto
			{
				Name: "dal/v1/user_datastore.proto",
				Pkg:  "dal.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source:    "api.v1.User",
							Kind:      "User",
							Namespace: "prod",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages, err := collector.CollectMessages(plugin, collector.TargetDatastore)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}

	// When: Generate converters
	result, err := GenerateConverters(messages)

	// Then: Should succeed
	if err != nil {
		t.Fatalf("GenerateConverters failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("Expected at least one converter file")
	}

	converterCode := result.Files[0].Content

	// Then: Should generate ToDatastore function
	if !strings.Contains(converterCode, "func UserToUserDatastore(") {
		t.Error("Expected UserToUserDatastore function")
	}

	// Should accept API message pointer as src
	if !strings.Contains(converterCode, "src *") {
		t.Error("Expected src parameter in ToDatastore")
	}

	// Should accept dest parameter for in-place conversion
	if !strings.Contains(converterCode, "dest *UserDatastore") {
		t.Error("Expected dest *UserDatastore parameter")
	}

	// Should accept decorator function
	if !strings.Contains(converterCode, "decorator func(*") {
		t.Error("Expected decorator parameter in ToDatastore")
	}

	// Should have FromDatastore converter
	if !strings.Contains(converterCode, "func UserFromUserDatastore(") {
		t.Error("Expected UserFromUserDatastore function")
	}

	// Should handle nil src
	if !strings.Contains(converterCode, "if src == nil") {
		t.Error("Expected nil check for src")
	}
}

// TestGenerateDatastore_DatastoreTags tests that datastore_tags are correctly applied
// to generated struct field tags.
//
// This test verifies that:
// - datastore_tags: ["-"] generates `datastore:"-"` (ignore field)
// - datastore_tags: ["noindex"] generates `datastore:"field_name,noindex"`
// - datastore_tags: ["noindex", "omitempty"] generates `datastore:"field_name,noindex,omitempty"`
func TestGenerateDatastore_DatastoreTags(t *testing.T) {
	// Given: A message with various datastore_tags
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			// API proto
			{
				Name: "api/v1/user.proto",
				Pkg:  "api.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "User",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "name", Number: 2, TypeName: "string"},
							{Name: "email", Number: 3, TypeName: "string"},
							{Name: "bio", Number: 4, TypeName: "string"},
						},
					},
				},
			},
			// Datastore DAL proto with datastore_tags
			{
				Name: "dal/v1/user_datastore.proto",
				Pkg:  "dal.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "api.v1.User",
							Kind:   "User",
						},
						Fields: []testutil.TestField{
							{
								Name:     "id",
								Number:   1,
								TypeName: "string",
								ColumnOpts: &dalv1.ColumnOptions{
									DatastoreTags: []string{"-"}, // Ignore field
								},
							},
							{
								Name:     "name",
								Number:   2,
								TypeName: "string",
								// No tags - should get default tag
							},
							{
								Name:     "email",
								Number:   3,
								TypeName: "string",
								ColumnOpts: &dalv1.ColumnOptions{
									DatastoreTags: []string{"noindex"}, // Single tag
								},
							},
							{
								Name:     "bio",
								Number:   4,
								TypeName: "string",
								ColumnOpts: &dalv1.ColumnOptions{
									DatastoreTags: []string{"noindex", "omitempty"}, // Multiple tags
								},
							},
						},
					},
				},
			},
		},
	})

	// Collect Datastore messages
	messages, err := collector.CollectMessages(plugin, collector.TargetDatastore)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("Expected 1 Datastore message, got %d", len(messages))
	}

	// When: Generate Datastore code
	result, err := Generate(messages)

	// Then: Should succeed
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("Expected at least one generated file")
	}

	content := result.Files[0].Content

	// Then: ID field should have datastore:"-" tag (ignored)
	if !strings.Contains(content, "`datastore:\"-\"`") {
		t.Errorf("Expected ID field with datastore:\"-\" tag for ignored field.\nGenerated content:\n%s", content)
	}

	// Then: Name field should have default tag (just field name)
	if !strings.Contains(content, "`datastore:\"name\"`") {
		t.Errorf("Expected Name field with default datastore:\"name\" tag.\nGenerated content:\n%s", content)
	}

	// Then: Email field should have noindex tag
	if !strings.Contains(content, "`datastore:\"email,noindex\"`") {
		t.Errorf("Expected Email field with datastore:\"email,noindex\" tag.\nGenerated content:\n%s", content)
	}

	// Then: Bio field should have noindex and omitempty tags
	if !strings.Contains(content, "`datastore:\"bio,noindex,omitempty\"`") {
		t.Errorf("Expected Bio field with datastore:\"bio,noindex,omitempty\" tag.\nGenerated content:\n%s", content)
	}
}

// Test helpers (copied from collector_test.go pattern)

// Test helpers have been moved to pkg/generator/testutil
