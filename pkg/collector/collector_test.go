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
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// TestCollectMessages_FindsPostgresMessages tests that the collector finds
// all messages with postgres annotations across multiple proto files
func TestCollectMessages_FindsPostgresMessages(t *testing.T) {
	// Create a plugin with test proto files
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// File 1: API proto with Book message
			{
				name: "library/v1/book.proto",
				pkg:  "library.v1",
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

// testProtoSet represents a complete set of proto files for testing.
// This is a simplified structure that captures just what we need for tests.
type testProtoSet struct {
	files []testFile
}

// testFile represents a single proto file with messages.
type testFile struct {
	name     string // e.g., "library/v1/book.proto"
	pkg      string // e.g., "library.v1"
	messages []testMessage
}

// testMessage represents a proto message definition.
// It can have optional postgres annotation to mark it as a DAL schema.
type testMessage struct {
	name         string                 // e.g., "Book" or "BookPostgres"
	postgresOpts *dalv1.PostgresOptions // If present, this is a DAL schema message
	fields       []testField
}

// testField represents a simple proto field.
// For tests, we only support basic types (string, int32, int64).
type testField struct {
	name     string // Field name
	number   int32  // Field number
	typeName string // "string", "int32", or "int64"
}

// createTestPlugin creates a protogen.Plugin from test data.
//
// Why is this needed?
// protogen.Plugin normally comes from protoc, but in tests we need to
// construct one manually. This helper builds the necessary proto descriptors
// from our simplified test data structure.
//
// Process:
// 1. Convert testProtoSet -> CodeGeneratorRequest (what protoc would send)
// 2. Create protogen.Plugin from the request
// 3. Plugin now contains proto files that can be analyzed
func createTestPlugin(t *testing.T, protoSet *testProtoSet) *protogen.Plugin {
	t.Helper()

	// Build the CodeGeneratorRequest (what protoc would send)
	req := buildCodeGeneratorRequest(t, protoSet)

	// Create plugin from request
	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	return plugin
}

// buildCodeGeneratorRequest constructs a CodeGeneratorRequest from test data.
//
// This is what protoc normally sends to the plugin. We're building it manually
// for testing purposes. The request contains:
// - ProtoFile: All proto file descriptors
// - FileToGenerate: Which files should be generated (all of them in tests)
func buildCodeGeneratorRequest(t *testing.T, protoSet *testProtoSet) *pluginpb.CodeGeneratorRequest {
	t.Helper()

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{},
	}

	// Convert each test file to a FileDescriptorProto
	for _, file := range protoSet.files {
		fileDesc := buildFileDescriptor(t, file)
		req.ProtoFile = append(req.ProtoFile, fileDesc)
		req.FileToGenerate = append(req.FileToGenerate, file.name)
	}

	return req
}

// buildFileDescriptor creates a FileDescriptorProto from test file data.
//
// FileDescriptorProto is protobuf's self-description of a .proto file.
// It contains all the messages, enums, services, etc. defined in that file.
func buildFileDescriptor(t *testing.T, file testFile) *descriptorpb.FileDescriptorProto {
	t.Helper()

	// Convert package name to go_package path
	// e.g., "library.v1" -> "github.com/test/gen/go/library/v1"
	goPackage := "github.com/test/gen/go/" + strings.ReplaceAll(file.pkg, ".", "/")

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String(file.name),
		Package: proto.String(file.pkg),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String(goPackage),
		},
	}

	// Add all messages to the file descriptor
	for _, msg := range file.messages {
		msgDesc := buildMessageDescriptor(t, msg)
		fileDesc.MessageType = append(fileDesc.MessageType, msgDesc)
	}

	return fileDesc
}

// buildMessageDescriptor creates a DescriptorProto from test message data.
//
// DescriptorProto describes a single message type including:
// - Fields
// - Options (annotations like postgres)
// - Nested types (not used in our tests)
func buildMessageDescriptor(t *testing.T, msg testMessage) *descriptorpb.DescriptorProto {
	t.Helper()

	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String(msg.name),
	}

	// Add all fields
	for _, field := range msg.fields {
		fieldDesc := &descriptorpb.FieldDescriptorProto{
			Name:   proto.String(field.name),
			Number: proto.Int32(field.number),
		}

		// Map simple type names to protobuf field types
		switch field.typeName {
		case "string":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		case "int32":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		case "int64":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	// Add postgres options if present (this marks it as a DAL schema message)
	if msg.postgresOpts != nil {
		opts := &descriptorpb.MessageOptions{}
		proto.SetExtension(opts, dalv1.E_Postgres, msg.postgresOpts)
		msgDesc.Options = opts
	}

	return msgDesc
}
