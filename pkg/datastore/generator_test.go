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
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

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
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto
			{
				name: "api/v1/user.proto",
				pkg:  "api.v1",
				messages: []testMessage{
					{
						name: "User",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "name", number: 2, typeName: "string"},
							{name: "email", number: 3, typeName: "string"},
						},
					},
				},
			},
			// Datastore DAL proto
			{
				name: "dal/v1/user_datastore.proto",
				pkg:  "dal.v1",
				messages: []testMessage{
					{
						name: "UserDatastore",
						datastoreOpts: &dalv1.DatastoreOptions{
							Source:    "api.v1.User",
							Kind:      "User",
							Namespace: "prod",
						},
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "name", number: 2, typeName: "string"},
							{name: "email", number: 3, typeName: "string"},
						},
					},
				},
			},
		},
	})

	// Collect Datastore messages
	messages := collector.CollectMessages(plugin, collector.TargetDatastore)
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
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto
			{
				name: "api/v1/user.proto",
				pkg:  "api.v1",
				messages: []testMessage{
					{
						name: "User",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "name", number: 2, typeName: "string"},
							{name: "email", number: 3, typeName: "string"},
						},
					},
				},
			},
			// Datastore DAL proto
			{
				name: "dal/v1/user_datastore.proto",
				pkg:  "dal.v1",
				messages: []testMessage{
					{
						name: "UserDatastore",
						datastoreOpts: &dalv1.DatastoreOptions{
							Source:    "api.v1.User",
							Kind:      "User",
							Namespace: "prod",
						},
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "name", number: 2, typeName: "string"},
							{name: "email", number: 3, typeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := collector.CollectMessages(plugin, collector.TargetDatastore)

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

// Test helpers (copied from collector_test.go pattern)

type testProtoSet struct {
	files []testFile
}

type testFile struct {
	name     string
	pkg      string
	messages []testMessage
}

type testMessage struct {
	name          string
	datastoreOpts *dalv1.DatastoreOptions
	fields        []testField
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
			fieldOpts := &descriptorpb.FieldOptions{}
			proto.SetExtension(fieldOpts, dalv1.E_Column, field.columnOpts)
			fieldDesc.Options = fieldOpts
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	// Add datastore options if present
	if msg.datastoreOpts != nil {
		opts := &descriptorpb.MessageOptions{}
		proto.SetExtension(opts, dalv1.E_DatastoreOptions, msg.datastoreOpts)
		msgDesc.Options = opts
	}

	return msgDesc
}
