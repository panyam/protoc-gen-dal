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

package collector

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	dalv1 "github.com/panyam/protoc-gen-go-dal/proto/gen/go/dal/v1"
)

// TestCollectMessages_FindsPostgresMessages tests that the collector finds
// all messages with postgres annotations across multiple proto files
func TestCollectMessages_FindsPostgresMessages(t *testing.T) {
	// Create a plugin with test proto files
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// File 1: API proto with Book message
			{
				name:    "library/v1/book.proto",
				pkg:     "library.v1",
				messages: []testMessage{
					{
						name: "Book",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "title", number: 2, typeName: "string"},
							{name: "author", number: 3, typeName: "string"},
						},
					},
				},
			},
			// File 2: DAL proto with BookPostgres message
			{
				name: "library/v1/dal/book_postgres.proto",
				pkg:  "library.v1.dal",
				messages: []testMessage{
					{
						name: "BookPostgres",
						postgresOpts: &dalv1.PostgresOptions{
							Source: "library.v1.Book",
							Table:  "books",
							Schema: "library",
						},
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "title", number: 2, typeName: "string"},
							{name: "author", number: 3, typeName: "string"},
						},
					},
				},
			},
		},
	})

	// Act: Collect postgres messages
	messages := CollectMessages(plugin, TargetPostgres)

	// Assert: Should find exactly one postgres message
	if len(messages) != 1 {
		t.Fatalf("Expected 1 postgres message, got %d", len(messages))
	}

	msg := messages[0]

	// Assert: Source message should be Book
	if msg.SourceMessage == nil {
		t.Fatal("SourceMessage should not be nil")
	}
	if string(msg.SourceMessage.Desc.Name()) != "Book" {
		t.Errorf("Expected source message 'Book', got '%s'", msg.SourceMessage.Desc.Name())
	}

	// Assert: Target message should be BookPostgres
	if msg.TargetMessage == nil {
		t.Fatal("TargetMessage should not be nil")
	}
	if string(msg.TargetMessage.Desc.Name()) != "BookPostgres" {
		t.Errorf("Expected target message 'BookPostgres', got '%s'", msg.TargetMessage.Desc.Name())
	}

	// Assert: Metadata should be extracted correctly
	if msg.SourceName != "library.v1.Book" {
		t.Errorf("Expected SourceName 'library.v1.Book', got '%s'", msg.SourceName)
	}
	if msg.TableName != "books" {
		t.Errorf("Expected TableName 'books', got '%s'", msg.TableName)
	}
	if msg.SchemaName != "library" {
		t.Errorf("Expected SchemaName 'library', got '%s'", msg.SchemaName)
	}
}

// TestCollectMessages_SkipsNonPostgresMessages tests that messages without
// postgres annotations are skipped
func TestCollectMessages_SkipsNonPostgresMessages(t *testing.T) {
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			{
				name: "library/v1/book.proto",
				pkg:  "library.v1",
				messages: []testMessage{
					{
						name: "Book", // No postgres annotation
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := CollectMessages(plugin, TargetPostgres)

	if len(messages) != 0 {
		t.Fatalf("Expected 0 postgres messages, got %d", len(messages))
	}
}

// TestCollectMessages_HandlesMultipleMessages tests collecting multiple
// postgres messages from different files
func TestCollectMessages_HandlesMultipleMessages(t *testing.T) {
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API protos
			{
				name: "library/v1/book.proto",
				pkg:  "library.v1",
				messages: []testMessage{
					{name: "Book", fields: []testField{{name: "id", number: 1, typeName: "string"}}},
				},
			},
			{
				name: "library/v1/author.proto",
				pkg:  "library.v1",
				messages: []testMessage{
					{name: "Author", fields: []testField{{name: "id", number: 1, typeName: "string"}}},
				},
			},
			// DAL protos
			{
				name: "library/v1/dal/book_postgres.proto",
				pkg:  "library.v1.dal",
				messages: []testMessage{
					{
						name: "BookPostgres",
						postgresOpts: &dalv1.PostgresOptions{
							Source: "library.v1.Book",
							Table:  "books",
						},
						fields: []testField{{name: "id", number: 1, typeName: "string"}},
					},
				},
			},
			{
				name: "library/v1/dal/author_postgres.proto",
				pkg:  "library.v1.dal",
				messages: []testMessage{
					{
						name: "AuthorPostgres",
						postgresOpts: &dalv1.PostgresOptions{
							Source: "library.v1.Author",
							Table:  "authors",
						},
						fields: []testField{{name: "id", number: 1, typeName: "string"}},
					},
				},
			},
		},
	})

	messages := CollectMessages(plugin, TargetPostgres)

	if len(messages) != 2 {
		t.Fatalf("Expected 2 postgres messages, got %d", len(messages))
	}

	// Verify both messages are found
	tables := make(map[string]bool)
	for _, msg := range messages {
		tables[msg.TableName] = true
	}

	if !tables["books"] {
		t.Error("Expected to find 'books' table")
	}
	if !tables["authors"] {
		t.Error("Expected to find 'authors' table")
	}
}

// Test helpers

type testProtoSet struct {
	files []testFile
}

type testFile struct {
	name     string
	pkg      string
	messages []testMessage
}

type testMessage struct {
	name         string
	postgresOpts *dalv1.PostgresOptions
	fields       []testField
}

type testField struct {
	name     string
	number   int32
	typeName string
}

// createTestPlugin creates a protogen.Plugin from test data
// This is a placeholder - will need proper implementation
func createTestPlugin(t *testing.T, protoSet *testProtoSet) *protogen.Plugin {
	t.Helper()

	// TODO: Implement proper test plugin creation
	// For now, this will make the test compile but fail when run
	req := &pluginpb.CodeGeneratorRequest{}

	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	return plugin
}
