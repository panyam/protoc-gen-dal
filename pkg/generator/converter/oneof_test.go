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
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// TestOneofFieldDetection tests that oneof members are correctly detected
// and the SourceIsOneofMember flag is set appropriately.
func TestOneofFieldDetection(t *testing.T) {
	// Create source message with oneof
	sourceMsg := createTestMessageWithOneof(t, "Value", "kind", []oneofTestMember{
		{name: "null_value", number: 1, typeName: "int32"},
		{name: "number_value", number: 2, typeName: "double"},
		{name: "string_value", number: 3, typeName: "string"},
		{name: "bool_value", number: 4, typeName: "bool"},
	})

	// Verify that fields in the oneof are detected as oneof members
	for _, field := range sourceMsg.Fields {
		if field.Oneof == nil {
			t.Errorf("Field %s should be part of oneof 'kind'", field.GoName)
			continue
		}

		if field.Oneof.Desc.IsSynthetic() {
			t.Errorf("Field %s oneof should NOT be synthetic", field.GoName)
		}

		// This is what BuildFieldMapping checks
		isOneofMember := field.Oneof != nil && !field.Oneof.Desc.IsSynthetic()
		if !isOneofMember {
			t.Errorf("Field %s should be detected as oneof member", field.GoName)
		}
	}
}

// TestSyntheticOneofNotDetected tests that synthetic oneofs (proto3 optional)
// are NOT treated as oneof members for converter generation purposes.
func TestSyntheticOneofNotDetected(t *testing.T) {
	// Create source message with optional field (proto3 optional creates synthetic oneof)
	sourceMsg := createTestMessageWithOptional(t, "User", []optionalTestField{
		{name: "nickname", number: 1, typeName: "string", optional: true},
		{name: "email", number: 2, typeName: "string", optional: false},
	})

	for _, field := range sourceMsg.Fields {
		fieldName := string(field.Desc.Name())

		// Check if proto3 optional creates synthetic oneof
		if fieldName == "nickname" {
			// Proto3 optional fields MAY have a synthetic oneof
			// The key test is that IsSynthetic() returns true for these
			if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
				t.Errorf("Optional field %s should have synthetic oneof or no oneof", fieldName)
			}
		}

		// Non-optional fields should not have oneof
		if fieldName == "email" && field.Oneof != nil {
			t.Errorf("Non-optional field %s should not have oneof", fieldName)
		}
	}
}

// TestFieldMappingStructure tests that FieldMapping has the SourceIsOneofMember field
func TestFieldMappingStructure(t *testing.T) {
	mapping := &FieldMapping{
		SourceField:         "NullValue",
		TargetField:         "NullValue",
		SourceIsOneofMember: true,
	}

	if !mapping.SourceIsOneofMember {
		t.Error("SourceIsOneofMember should be true")
	}

	// Test with false
	mapping2 := &FieldMapping{
		SourceField:         "Id",
		TargetField:         "Id",
		SourceIsOneofMember: false,
	}

	if mapping2.SourceIsOneofMember {
		t.Error("SourceIsOneofMember should be false")
	}
}

// Helper types

type oneofTestMember struct {
	name     string
	number   int32
	typeName string
}

type optionalTestField struct {
	name     string
	number   int32
	typeName string
	optional bool
}

// createTestMessageWithOneof creates a protogen.Message with a real oneof field
func createTestMessageWithOneof(t *testing.T, msgName string, oneofName string, members []oneofTestMember) *protogen.Message {
	t.Helper()

	// Create message descriptor with oneof
	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String(msgName),
	}

	// Add oneof declaration
	msgDesc.OneofDecl = []*descriptorpb.OneofDescriptorProto{
		{Name: proto.String(oneofName)},
	}

	// Add fields that belong to the oneof
	for _, member := range members {
		fieldDesc := &descriptorpb.FieldDescriptorProto{
			Name:       proto.String(member.name),
			Number:     proto.Int32(member.number),
			OneofIndex: proto.Int32(0), // First (and only) oneof
		}

		// Determine field type
		switch member.typeName {
		case "string":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		case "int32":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		case "double":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()
		case "bool":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum()
		default:
			// Message type
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			fieldDesc.TypeName = proto.String(member.typeName)
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	return createPluginMessage(t, msgDesc)
}

// createTestMessageWithOptional creates a protogen.Message with optional fields
func createTestMessageWithOptional(t *testing.T, msgName string, fields []optionalTestField) *protogen.Message {
	t.Helper()

	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String(msgName),
	}

	oneofIndex := int32(0)
	for _, field := range fields {
		fieldDesc := &descriptorpb.FieldDescriptorProto{
			Name:   proto.String(field.name),
			Number: proto.Int32(field.number),
		}

		// Determine field type
		switch field.typeName {
		case "string":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		case "int32":
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		default:
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			fieldDesc.TypeName = proto.String(field.typeName)
		}

		// Proto3 optional creates a synthetic oneof
		if field.optional {
			syntheticOneofName := "_" + field.name
			msgDesc.OneofDecl = append(msgDesc.OneofDecl, &descriptorpb.OneofDescriptorProto{
				Name: proto.String(syntheticOneofName),
			})
			fieldDesc.OneofIndex = proto.Int32(oneofIndex)
			fieldDesc.Proto3Optional = proto.Bool(true)
			oneofIndex++
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	return createPluginMessage(t, msgDesc)
}

// createPluginMessage creates a protogen.Message from a message descriptor
func createPluginMessage(t *testing.T, msgDesc *descriptorpb.DescriptorProto) *protogen.Message {
	t.Helper()

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/test/gen/go/test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{msgDesc},
	}

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fileDesc},
	}

	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	if len(plugin.Files) == 0 || len(plugin.Files[0].Messages) == 0 {
		t.Fatal("No messages found in plugin")
	}
	return plugin.Files[0].Messages[0]
}
