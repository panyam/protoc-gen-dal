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

package testutil

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// TestProtoSet represents a collection of proto files for testing.
type TestProtoSet struct {
	Files []TestFile
}

// TestFile represents a single proto file with messages.
type TestFile struct {
	Name     string
	Pkg      string
	Messages []TestMessage
}

// TestMessage represents a proto message with optional DAL options.
type TestMessage struct {
	Name          string
	Fields        []TestField
	GormOpts      *dalv1.GormOptions
	DatastoreOpts *dalv1.DatastoreOptions
}

// TestField represents a proto field.
type TestField struct {
	Name       string
	Number     int32
	TypeName   string
	ColumnOpts *dalv1.ColumnOptions
	Repeated   bool
	IsMap      bool
	MapKeyType string // For map fields: "int32", "string", etc.
}

// CreateTestPlugin creates a protogen.Plugin from a test proto set.
func CreateTestPlugin(t *testing.T, protoSet *TestProtoSet) *protogen.Plugin {
	t.Helper()

	req := BuildCodeGeneratorRequest(t, protoSet)
	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	return plugin
}

// BuildCodeGeneratorRequest creates a CodeGeneratorRequest from a test proto set.
func BuildCodeGeneratorRequest(t *testing.T, protoSet *TestProtoSet) *pluginpb.CodeGeneratorRequest {
	t.Helper()

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{},
	}

	for _, file := range protoSet.Files {
		fileDesc := BuildFileDescriptor(t, file)
		req.ProtoFile = append(req.ProtoFile, fileDesc)
		req.FileToGenerate = append(req.FileToGenerate, file.Name)
	}

	return req
}

// BuildFileDescriptor creates a FileDescriptorProto from a test file.
func BuildFileDescriptor(t *testing.T, file TestFile) *descriptorpb.FileDescriptorProto {
	t.Helper()

	goPackage := "github.com/test/gen/go/" + strings.ReplaceAll(file.Pkg, ".", "/")

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String(file.Name),
		Package: proto.String(file.Pkg),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String(goPackage),
		},
	}

	for _, msg := range file.Messages {
		msgDesc := BuildMessageDescriptorWithPackage(t, msg, file.Pkg)
		fileDesc.MessageType = append(fileDesc.MessageType, msgDesc)
	}

	return fileDesc
}

// BuildMessageDescriptor creates a DescriptorProto from a test message (no package context).
func BuildMessageDescriptor(t *testing.T, msg TestMessage) *descriptorpb.DescriptorProto {
	return BuildMessageDescriptorWithPackage(t, msg, "")
}

// BuildMessageDescriptorWithPackage creates a DescriptorProto from a test message with package context.
func BuildMessageDescriptorWithPackage(t *testing.T, msg TestMessage, pkg string) *descriptorpb.DescriptorProto {
	t.Helper()

	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String(msg.Name),
	}

	// Add fields
	for _, field := range msg.Fields {
		if field.IsMap {
			// Map fields require a nested entry message
			// Capitalize first letter of field name
			fieldName := field.Name
			if len(fieldName) > 0 {
				fieldName = strings.ToUpper(fieldName[:1]) + fieldName[1:]
			}
			entryMsgName := fieldName + "Entry"
			entryMsg := &descriptorpb.DescriptorProto{
				Name: proto.String(entryMsgName),
				Options: &descriptorpb.MessageOptions{
					MapEntry: proto.Bool(true),
				},
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("key"),
						Number: proto.Int32(1),
						Type:   GetFieldType(field.MapKeyType),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:     proto.String("value"),
						Number:   proto.Int32(2),
						Type:     GetFieldType(field.TypeName),
						TypeName: GetTypeName(field.TypeName),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			}
			msgDesc.NestedType = append(msgDesc.NestedType, entryMsg)

			// Add the map field itself
			fullEntryName := "." + pkg + "." + msg.Name + "." + entryMsgName
			if pkg == "" {
				fullEntryName = "." + msg.Name + "." + entryMsgName
			}
			fieldDesc := &descriptorpb.FieldDescriptorProto{
				Name:     proto.String(field.Name),
				Number:   proto.Int32(field.Number),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String(fullEntryName),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			}
			msgDesc.Field = append(msgDesc.Field, fieldDesc)
		} else {
			fieldDesc := &descriptorpb.FieldDescriptorProto{
				Name:   proto.String(field.Name),
				Number: proto.Int32(field.Number),
			}

			switch field.TypeName {
			case "string":
				fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
			case "int32":
				fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
			case "int64":
				fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()
			default:
				// If not a scalar type, assume it's a message type
				fieldDesc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
				fieldDesc.TypeName = proto.String("." + field.TypeName)
			}

			// Set label for repeated fields
			if field.Repeated {
				fieldDesc.Label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()
			}

			// Add column options if present
			if field.ColumnOpts != nil {
				opts := &descriptorpb.FieldOptions{}
				proto.SetExtension(opts, dalv1.E_Column, field.ColumnOpts)
				fieldDesc.Options = opts
			}

			msgDesc.Field = append(msgDesc.Field, fieldDesc)
		}
	}

	// Add DAL options if present
	if msg.GormOpts != nil {
		opts := &descriptorpb.MessageOptions{}
		proto.SetExtension(opts, dalv1.E_Gorm, msg.GormOpts)
		msgDesc.Options = opts
	}
	if msg.DatastoreOpts != nil {
		opts := &descriptorpb.MessageOptions{}
		proto.SetExtension(opts, dalv1.E_DatastoreOptions, msg.DatastoreOpts)
		msgDesc.Options = opts
	}

	return msgDesc
}

// GetFieldType returns the proto field type enum for a type name.
func GetFieldType(typeName string) *descriptorpb.FieldDescriptorProto_Type {
	switch typeName {
	case "string":
		return descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
	case "int32":
		return descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
	case "int64":
		return descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()
	default:
		return descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
	}
}

// GetTypeName returns the full type name for message types, nil for scalars.
func GetTypeName(typeName string) *string {
	switch typeName {
	case "string", "int32", "int64":
		return nil
	default:
		return proto.String("." + typeName)
	}
}
