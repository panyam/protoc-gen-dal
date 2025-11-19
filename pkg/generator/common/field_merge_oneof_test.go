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
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// TestMergeSourceFields_OneofReplacement tests that when target has a field
// matching a source oneof name, all oneof members are automatically skipped
func TestMergeSourceFields_OneofReplacement(t *testing.T) {
	// Create source message with oneof
	// Use string type for simplicity (no message type references needed)
	sourceMsg := createMessageWithOneof(t, "GameMove", "move_type", []oneofMember{
		{name: "move_unit", number: 4, typeName: "string"},
		{name: "attack_unit", number: 5, typeName: "string"},
	})

	// Create target message with field matching oneof name
	targetMsg := createMessage(t, "GameMoveGORM", []testField{
		{name: "move_type", number: 1, typeName: "string"},
	})

	// Act: Merge fields
	merged, err := MergeSourceFields(sourceMsg, targetMsg)

	// Assert: Should succeed (once implemented)
	if err != nil {
		t.Fatalf("MergeSourceFields failed: %v", err)
	}

	// Check that result has move_type but NOT move_unit or attack_unit
	fieldNames := make(map[string]bool)
	for _, field := range merged {
		fieldNames[string(field.Desc.Name())] = true
	}

	if !fieldNames["move_type"] {
		t.Error("Expected 'move_type' field in merged result")
	}

	// These should be auto-skipped because move_type matches the oneof name
	if fieldNames["move_unit"] {
		t.Error("Expected 'move_unit' to be auto-skipped (oneof member)")
	}
	if fieldNames["attack_unit"] {
		t.Error("Expected 'attack_unit' to be auto-skipped (oneof member)")
	}
}

// TestMergeSourceFields_OneofNoReplacement tests that oneof members are included
// when there's NO field matching the oneof name in target
func TestMergeSourceFields_OneofNoReplacement(t *testing.T) {
	// Create source message with oneof (using string type for simplicity)
	sourceMsg := createMessageWithOneof(t, "GameMove", "move_type", []oneofMember{
		{name: "move_unit", number: 4, typeName: "string"},
		{name: "attack_unit", number: 5, typeName: "string"},
	})

	// Create target message with NO field matching oneof name
	targetMsg := createMessage(t, "GameMoveGORM", []testField{
		{name: "player", number: 1, typeName: "int32"},
	})

	// Act: Merge fields
	merged, err := MergeSourceFields(sourceMsg, targetMsg)

	// For now, should succeed and include oneof members
	// TODO: Later we might want to error if oneof is not handled
	if err != nil {
		t.Fatalf("MergeSourceFields failed: %v", err)
	}

	// Check that oneof members ARE included (no replacement field exists)
	fieldNames := make(map[string]bool)
	for _, field := range merged {
		fieldNames[string(field.Desc.Name())] = true
	}

	if !fieldNames["move_unit"] {
		t.Error("Expected 'move_unit' to be included (no oneof replacement)")
	}
	if !fieldNames["attack_unit"] {
		t.Error("Expected 'attack_unit' to be included (no oneof replacement)")
	}
}

// Helper types and functions

type oneofMember struct {
	name     string
	number   int32
	typeName string
}

type testField struct {
	name     string
	number   int32
	typeName string
}

// createMessageWithOneof creates a protogen.Message with a oneof field
func createMessageWithOneof(t *testing.T, msgName string, oneofName string, members []oneofMember) *protogen.Message {
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
		if member.typeName == "string" {
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		} else if member.typeName == "int32" {
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		} else {
			// Message type
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			fieldDesc.TypeName = proto.String(member.typeName)
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	// Create file descriptor
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/test/gen/go/test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{msgDesc},
	}

	// Create plugin
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fileDesc},
	}

	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Return the first message
	if len(plugin.Files) == 0 || len(plugin.Files[0].Messages) == 0 {
		t.Fatal("No messages found in plugin")
	}
	return plugin.Files[0].Messages[0]
}

// createMessage creates a simple protogen.Message without oneofs
func createMessage(t *testing.T, msgName string, fields []testField) *protogen.Message {
	t.Helper()

	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String(msgName),
	}

	// Add fields
	for _, field := range fields {
		fieldDesc := &descriptorpb.FieldDescriptorProto{
			Name:   proto.String(field.name),
			Number: proto.Int32(field.number),
		}

		// Determine field type
		if field.typeName == "string" {
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		} else if field.typeName == "int32" {
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		} else if strings.Contains(field.typeName, ".") {
			// Message type
			fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			fieldDesc.TypeName = proto.String(field.typeName)
		}

		msgDesc.Field = append(msgDesc.Field, fieldDesc)
	}

	// Create file descriptor
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/test/gen/go/test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{msgDesc},
	}

	// Create plugin
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fileDesc},
	}

	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Return the first message
	if len(plugin.Files) == 0 || len(plugin.Files[0].Messages) == 0 {
		t.Fatal("No messages found in plugin")
	}
	return plugin.Files[0].Messages[0]
}
