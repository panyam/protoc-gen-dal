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
	"strings"
	"testing"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// TestGenerateGORM_SimpleMessage tests that a simple message generates
// a correct GORM struct with basic fields
//
// This is the first test that drives the GORM generator implementation.
// It verifies that we can generate:
// - A properly named struct (BookGORM)
// - Basic fields with correct GORM tags
// - A TableName() method
func TestGenerateGORM_SimpleMessage(t *testing.T) {
	// Given: A simple Book message with basic fields
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto
			{
				name: "library/v1/book.proto",
				pkg:  "library.v1",
				messages: []testMessage{
					{
						name: "Book",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "title", number: 2, typeName: "string"},
						},
					},
				},
			},
			// GORM DAL proto
			{
				name: "library/v1/dal/book_gorm.proto",
				pkg:  "library.v1.dal",
				messages: []testMessage{
					{
						name: "BookGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "library.v1.Book",
							Table:  "books",
						},
						fields: []testField{
							{
								name:     "id",
								number:   1,
								typeName: "string",
								columnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{name: "title", number: 2, typeName: "string"},
						},
					},
				},
			},
		},
	})

	// Collect GORM messages
	messages := collector.CollectMessages(plugin, collector.TargetGorm)
	if len(messages) != 1 {
		t.Fatalf("Expected 1 GORM message, got %d", len(messages))
	}

	// When: Generate GORM code
	result, err := Generate(messages)

	// Then: Should succeed
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Then: Should generate at least one file
	if len(result.Files) == 0 {
		t.Fatal("Expected at least one generated file")
	}

	generatedCode := result.Files[0].Content

	// Then: Should generate GORM struct with correct name
	if !strings.Contains(generatedCode, "type BookGORM struct") {
		t.Error("Expected 'type BookGORM struct' in generated code")
	}

	// Then: Should generate fields with GORM tags
	if !strings.Contains(generatedCode, "ID") && !strings.Contains(generatedCode, "Id") {
		t.Error("Expected ID field in generated struct")
	}
	if !strings.Contains(generatedCode, "Title") {
		t.Error("Expected Title field in generated struct")
	}

	// Then: Should generate primary key tag
	if !strings.Contains(generatedCode, "primaryKey") {
		t.Error("Expected primaryKey tag for ID field")
	}

	// Then: Should generate TableName method with pointer receiver
	if !strings.Contains(generatedCode, "func (*BookGORM) TableName() string") {
		t.Error("Expected TableName() method with pointer receiver in generated code")
	}

	if !strings.Contains(generatedCode, `return "books"`) {
		t.Error("Expected TableName() to return 'books'")
	}
}

// TestGenerateConverters tests that converter functions are generated
// for converting between API messages and GORM structs.
//
// This test verifies:
// - ToGORM converter function
// - FromGORM converter function
// - Pointer receivers to avoid struct copying
func TestGenerateConverters(t *testing.T) {
	// Given: A Book message with GORM mapping
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto
			{
				name: "library/v1/book.proto",
				pkg:  "library.v1",
				messages: []testMessage{
					{
						name: "Book",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "title", number: 2, typeName: "string"},
						},
					},
				},
			},
			// GORM DAL proto
			{
				name: "library/v1/dal/book_gorm.proto",
				pkg:  "library.v1.dal",
				messages: []testMessage{
					{
						name: "BookGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "library.v1.Book",
							Table:  "books",
						},
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "title", number: 2, typeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := collector.CollectMessages(plugin, collector.TargetGorm)

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

	// Then: Should generate ToGORM function with full name (SourceToGormType)
	if !strings.Contains(converterCode, "func BookToBookGORM(") {
		t.Error("Expected BookToBookGORM function")
	}

	// Should accept API message pointer as src
	if !strings.Contains(converterCode, "src *") {
		t.Error("Expected src parameter in ToGORM")
	}

	// Should accept dest parameter for in-place conversion
	if !strings.Contains(converterCode, "dest *BookGORM") {
		t.Error("Expected dest *BookGORM parameter in ToGORM")
	}

	// Should return named GORM struct pointer and error
	if !strings.Contains(converterCode, "(out *BookGORM, err error)") {
		t.Error("Expected (out *BookGORM, err error) return in ToGORM")
	}

	// Then: Should generate FromGORM function with full name (SourceFromTargetType)
	if !strings.Contains(converterCode, "func BookFromBookGORM(") {
		t.Error("Expected BookFromBookGORM function")
	}

	// Should accept dest parameter (for proto message)
	if !strings.Contains(converterCode, "dest *") {
		t.Error("Expected dest parameter in FromGORM")
	}

	// Should accept GORM struct pointer as src
	if !strings.Contains(converterCode, "src *BookGORM") {
		t.Error("Expected src *BookGORM parameter in FromGORM")
	}

	// Should return named API message pointer and error (with package qualifier)
	if !strings.Contains(converterCode, ".Book") || !strings.Contains(converterCode, "(out *") {
		t.Error("Expected qualified Book type and named return in FromGORM")
	}
}

// Test helpers - similar to collector tests but extended for GORM

type testProtoSet struct {
	files []testFile
}

type testFile struct {
	name     string
	pkg      string
	messages []testMessage
}

type testMessage struct {
	name     string
	gormOpts *dalv1.GormOptions
	fields   []testField
}

type testField struct {
	name       string
	number     int32
	typeName   string
	columnOpts *dalv1.ColumnOptions
}

func createTestPlugin(t *testing.T, protoSet *testProtoSet) *protogen.Plugin {
	t.Helper()

	req := buildCodeGeneratorRequest(t, protoSet)
	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	return plugin
}

func buildCodeGeneratorRequest(t *testing.T, protoSet *testProtoSet) *pluginpb.CodeGeneratorRequest {
	t.Helper()

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{},
	}

	for _, file := range protoSet.files {
		fileDesc := buildFileDescriptor(t, file)
		req.ProtoFile = append(req.ProtoFile, fileDesc)
		req.FileToGenerate = append(req.FileToGenerate, file.name)
	}

	return req
}

func buildFileDescriptor(t *testing.T, file testFile) *descriptorpb.FileDescriptorProto {
	t.Helper()

	goPackage := "github.com/test/gen/go/" + strings.ReplaceAll(file.pkg, ".", "/")

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String(file.name),
		Package: proto.String(file.pkg),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String(goPackage),
		},
	}

	for _, msg := range file.messages {
		msgDesc := buildMessageDescriptor(t, msg)
		fileDesc.MessageType = append(fileDesc.MessageType, msgDesc)
	}

	return fileDesc
}

func buildMessageDescriptor(t *testing.T, msg testMessage) *descriptorpb.DescriptorProto {
	t.Helper()

	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String(msg.name),
	}

	// Add fields
	for _, field := range msg.fields {
		fieldDesc := &descriptorpb.FieldDescriptorProto{
			Name:   proto.String(field.name),
			Number: proto.Int32(field.number),
		}

		switch field.typeName {
		case "string":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		case "int32":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		case "int64":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()
		}

		// Add column options if present
		if field.columnOpts != nil {
			opts := &descriptorpb.FieldOptions{}
			proto.SetExtension(opts, dalv1.E_Column, field.columnOpts)
			fieldDesc.Options = opts
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	// Add GORM options if present
	if msg.gormOpts != nil {
		opts := &descriptorpb.MessageOptions{}
		proto.SetExtension(opts, dalv1.E_Gorm, msg.gormOpts)
		msgDesc.Options = opts
	}

	return msgDesc
}
