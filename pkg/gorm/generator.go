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
	"fmt"
	"strings"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// GeneratedFile represents a single generated file.
type GeneratedFile struct {
	// Path is the output file path (e.g., "book_gorm.go")
	Path string

	// Content is the generated Go code
	Content string
}

// GenerateResult contains all generated files.
type GenerateResult struct {
	// Files is the list of generated files
	Files []*GeneratedFile
}

// Generate generates GORM code for the given messages.
//
// This is the main entry point for GORM code generation. It receives all
// messages collected for the GORM target and generates:
// - GORM struct definitions with tags
// - TableName() methods
// - Converter functions (ToGORM/FromGORM)
// - Optional repository patterns
//
// Parameters:
//   - messages: Collected GORM messages from the collector
//
// Returns:
//   - GenerateResult containing all generated files
//   - error if generation fails
func Generate(messages []*collector.MessageInfo) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Group messages by their source proto file
	fileGroups := groupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one file per proto file
	for protoFile, msgs := range fileGroups {
		content, err := generateFileCode(msgs)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code for %s: %w", protoFile, err)
		}

		// Generate filename based on the proto file
		// e.g., gorm/user.proto -> user_gorm.go
		filename := generateFilenameFromProto(protoFile)

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// groupMessagesByFile groups messages by their source proto file path.
func groupMessagesByFile(messages []*collector.MessageInfo) map[string][]*collector.MessageInfo {
	groups := make(map[string][]*collector.MessageInfo)
	for _, msg := range messages {
		// Get the proto file path from the target message
		protoFile := msg.TargetMessage.Desc.ParentFile().Path()
		groups[protoFile] = append(groups[protoFile], msg)
	}
	return groups
}

// generateFilenameFromProto creates the output filename from a proto file path.
// e.g., "gorm/user.proto" -> "user_gorm.go"
func generateFilenameFromProto(protoPath string) string {
	// Extract base name without extension
	// e.g., "gorm/user.proto" -> "user"
	baseName := protoPath
	if idx := strings.LastIndex(baseName, "/"); idx != -1 {
		baseName = baseName[idx+1:]
	}
	if idx := strings.LastIndex(baseName, ".proto"); idx != -1 {
		baseName = baseName[:idx]
	}
	return baseName + "_gorm.go"
}

// generateFileCode generates the complete Go code for all messages in a proto file.
func generateFileCode(messages []*collector.MessageInfo) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate")
	}

	// Extract package name from the first message's target
	packageName := extractPackageName(messages[0].TargetMessage)

	// Build struct data for all messages with GORM annotations
	var structs []StructData
	embeddedTypes := make(map[string]*protogen.Message)

	for _, msg := range messages {
		structData, err := buildStructData(msg)
		if err != nil {
			return "", err
		}
		structs = append(structs, structData)

		// Collect embedded message types
		collectEmbeddedTypes(msg.TargetMessage, embeddedTypes)
	}

	// Add structs for embedded types that aren't already GORM messages
	for _, embMsg := range embeddedTypes {
		// Check if this message is already in our GORM messages
		isGormMsg := false
		for _, msg := range messages {
			if msg.TargetMessage == embMsg {
				isGormMsg = true
				break
			}
		}

		if !isGormMsg {
			// Build a simple struct for this embedded type (no table name)
			fields, err := buildFields(embMsg)
			if err != nil {
				return "", err
			}
			structs = append(structs, StructData{
				Name:      buildStructName(embMsg),
				TableName: "", // No table for embedded types
				Fields:    fields,
			})
		}
	}

	// Build template data
	data := TemplateData{
		PackageName: packageName,
		Imports:     []string{}, // TODO: Determine required imports
		Structs:     structs,
	}

	// Render the file template
	return renderTemplate("file.go.tmpl", data)
}

// collectEmbeddedTypes collects all message-type fields from a message.
func collectEmbeddedTypes(msg *protogen.Message, types map[string]*protogen.Message) {
	for _, field := range msg.Fields {
		if field.Desc.Kind().String() == "message" && field.Message != nil {
			// Add to map (using full name as key to avoid duplicates)
			fullName := string(field.Message.Desc.FullName())
			if _, exists := types[fullName]; !exists {
				types[fullName] = field.Message
			}
		}
	}
}

// extractPackageName extracts the Go package name from a protogen message.
// For example, "github.com/panyam/protoc-gen-dal/test/gen/dal;testdal" -> "testdal"
func extractPackageName(msg *protogen.Message) string {
	// GoImportPath format is "path/to/package" or "path/to/package;packagename"
	importPath := string(msg.GoIdent.GoImportPath)

	// Check if there's a package override (after semicolon)
	if idx := strings.LastIndex(importPath, ";"); idx != -1 {
		return importPath[idx+1:]
	}

	// Otherwise use the last part of the path
	if idx := strings.LastIndex(importPath, "/"); idx != -1 {
		return importPath[idx+1:]
	}

	return importPath
}

// buildStructData extracts struct information from a MessageInfo.
func buildStructData(msg *collector.MessageInfo) (StructData, error) {
	targetMsg := msg.TargetMessage

	// Build struct name: e.g., "BookGorm" -> "BookGORM"
	structName := buildStructName(targetMsg)

	// Build fields
	fields, err := buildFields(targetMsg)
	if err != nil {
		return StructData{}, err
	}

	return StructData{
		Name:       structName,
		SourceName: msg.SourceName,
		TableName:  msg.TableName,
		Fields:     fields,
	}, nil
}

// buildStructName generates the GORM struct name from the target message name.
// E.g., "BookGorm" -> "BookGORM"
func buildStructName(msg *protogen.Message) string {
	name := string(msg.Desc.Name())
	// Replace "Gorm" suffix with "GORM"
	if strings.HasSuffix(name, "Gorm") {
		name = strings.TrimSuffix(name, "Gorm") + "GORM"
	}
	return name
}

// buildFields extracts field information from a proto message.
func buildFields(msg *protogen.Message) ([]FieldData, error) {
	var fields []FieldData

	for _, field := range msg.Fields {
		fieldData, err := buildField(field)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldData)
	}

	return fields, nil
}

// buildField converts a proto field to FieldData.
func buildField(field *protogen.Field) (FieldData, error) {
	// Convert field name to Go naming: id -> ID, title -> Title
	goName := field.GoName

	// Convert proto type to Go type
	goType := protoToGoType(field)

	// Extract GORM tags from column options
	gormTag := extractGormTags(field)

	return FieldData{
		Name:    goName,
		Type:    goType,
		GormTag: gormTag,
	}, nil
}

// protoToGoType converts a proto field type to a Go type string.
func protoToGoType(field *protogen.Field) string {
	kind := field.Desc.Kind().String()

	// Handle message types (embedded structs, etc.)
	if kind == "message" {
		// For repeated message fields
		if field.Desc.Cardinality().String() == "repeated" {
			return "[]" + buildStructName(field.Message)
		}
		// For singular message fields, use the GORM struct name
		return buildStructName(field.Message)
	}

	// Handle repeated scalar fields
	if field.Desc.Cardinality().String() == "repeated" {
		elemType := protoScalarToGo(kind)
		return "[]" + elemType
	}

	return protoScalarToGo(kind)
}

// protoScalarToGo maps proto scalar types to Go types.
func protoScalarToGo(protoType string) string {
	switch protoType {
	case "string":
		return "string"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "bool":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "bytes":
		return "[]byte"
	default:
		return "interface{}" // Fallback for unknown types
	}
}

// extractGormTags extracts GORM tags from field column options.
func extractGormTags(field *protogen.Field) string {
	opts := field.Desc.Options()
	if opts == nil {
		return ""
	}

	// Get column options
	if v := proto.GetExtension(opts, dalv1.E_Column); v != nil {
		if colOpts, ok := v.(*dalv1.ColumnOptions); ok && colOpts != nil {
			// Join gorm_tags with semicolons
			if len(colOpts.GormTags) > 0 {
				return strings.Join(colOpts.GormTags, ";")
			}
		}
	}

	return ""
}
