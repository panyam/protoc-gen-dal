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

// TestGenerateDALHelpers_BasicGeneration verifies that DAL helpers are generated
// for Datastore messages with the dal option enabled.
func TestGenerateDALHelpers_BasicGeneration(t *testing.T) {
	// Create a test message with Datastore options
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "test.v1.User",
							Kind:   "User",
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

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   true,
		},
	}

	// Generate DAL helpers
	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	// Should generate at least one file
	if len(result.Files) == 0 {
		t.Fatal("Expected at least one generated file")
	}

	content := result.Files[0].Content

	// Verify basic DAL structure elements
	expectedElements := []string{
		"type UserDatastoreDAL struct",
		"Kind string",
		"Namespace string",
		"WillPut func(context.Context, *UserDatastore) error",
		"func NewUserDatastoreDAL(",
		"func (d *UserDatastoreDAL) Put(",
		"func (d *UserDatastoreDAL) Get(",
		"func (d *UserDatastoreDAL) Delete(",
		"func (d *UserDatastoreDAL) GetMulti(",
		"func (d *UserDatastoreDAL) PutMulti(",
		"func (d *UserDatastoreDAL) Query(",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected generated code to contain '%s'", expected)
		}
	}
}

// TestGenerateDALHelpers_IDBasedConvenienceMethods verifies that ID-based
// convenience methods are generated alongside key-based methods.
func TestGenerateDALHelpers_IDBasedConvenienceMethods(t *testing.T) {
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "test.v1.User",
							Kind:   "User",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "name", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   true,
		},
	}

	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	content := result.Files[0].Content

	// Verify ID-based convenience methods
	expectedMethods := []string{
		"func (d *UserDatastoreDAL) GetByID(",
		"func (d *UserDatastoreDAL) DeleteByID(",
		"func (d *UserDatastoreDAL) GetMultiByIDs(",
	}

	for _, expected := range expectedMethods {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected generated code to contain ID-based method '%s'", expected)
		}
	}
}

// TestGenerateDALHelpers_SkipsWhenDALFalse verifies that DAL generation is
// skipped when GenerateDAL is false.
func TestGenerateDALHelpers_SkipsWhenDALFalse(t *testing.T) {
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "test.v1.User",
							Kind:   "User",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   false, // DAL generation disabled
		},
	}

	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	// Should not generate any files when DAL is disabled
	if len(result.Files) != 0 {
		t.Errorf("Expected no files when GenerateDAL is false, got %d", len(result.Files))
	}
}

// TestGenerateDALHelpers_EmptyMessages verifies handling of empty message list.
func TestGenerateDALHelpers_EmptyMessages(t *testing.T) {
	result, err := GenerateDALHelpers([]*collector.MessageInfo{}, &DALOptions{})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	if len(result.Files) != 0 {
		t.Errorf("Expected no files for empty messages, got %d", len(result.Files))
	}
}

// TestGenerateDALFilename verifies correct filename generation for DAL files.
func TestGenerateDALFilename_DefaultSuffix(t *testing.T) {
	options := &DALOptions{
		FilenameSuffix: "_dal",
	}

	filename := generateDALFilename("datastore/user.proto", options)
	// Directory structure is preserved: datastore/user.proto -> datastore/user_datastore_dal.go
	expected := "datastore/user_datastore_dal.go"

	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

func TestGenerateDALFilename_WithOutputDir(t *testing.T) {
	options := &DALOptions{
		FilenameSuffix: "_dal",
		OutputDir:      "dal",
	}

	filename := generateDALFilename("datastore/user.proto", options)
	// Output dir prepended to full path with preserved directory
	expected := "dal/datastore/user_datastore_dal.go"

	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

// TestGenerateDALHelpers_NamespaceSupport verifies namespace field is included
// in the generated DAL struct.
func TestGenerateDALHelpers_NamespaceSupport(t *testing.T) {
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source:    "test.v1.User",
							Kind:      "User",
							Namespace: "prod",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   true,
		},
	}

	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	content := result.Files[0].Content

	// Verify namespace is used in key creation
	if !strings.Contains(content, "Namespace") {
		t.Error("Expected generated code to support Namespace field")
	}
}

// TestGenerateDALHelpers_WillPutHook verifies the WillPut hook is called
// before Put operations.
func TestGenerateDALHelpers_WillPutHook(t *testing.T) {
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "test.v1.User",
							Kind:   "User",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   true,
		},
	}

	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	content := result.Files[0].Content

	// Verify WillPut hook is called in Put method
	if !strings.Contains(content, "d.WillPut") {
		t.Error("Expected Put method to call WillPut hook")
	}
}

// TestGenerateDALHelpers_CountMethod verifies the Count method is generated.
func TestGenerateDALHelpers_CountMethod(t *testing.T) {
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "test.v1.User",
							Kind:   "User",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   true,
		},
	}

	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	content := result.Files[0].Content

	// Verify Count method exists
	if !strings.Contains(content, "func (d *UserDatastoreDAL) Count(") {
		t.Error("Expected generated code to contain Count method")
	}
}

// TestGenerateDALHelpers_DeleteMulti verifies DeleteMulti method is generated.
func TestGenerateDALHelpers_DeleteMulti(t *testing.T) {
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserDatastore",
						DatastoreOpts: &dalv1.DatastoreOptions{
							Source: "test.v1.User",
							Kind:   "User",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{
			TargetMessage: plugin.Files[0].Messages[0],
			GenerateDAL:   true,
		},
	}

	result, err := GenerateDALHelpers(messages, &DALOptions{
		FilenameSuffix: "_dal",
	})
	if err != nil {
		t.Fatalf("GenerateDALHelpers failed: %v", err)
	}

	content := result.Files[0].Content

	// Verify DeleteMulti method exists
	if !strings.Contains(content, "func (d *UserDatastoreDAL) DeleteMulti(") {
		t.Error("Expected generated code to contain DeleteMulti method")
	}
}
