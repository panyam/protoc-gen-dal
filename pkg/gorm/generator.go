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

	dalv1 "github.com/panyam/protoc-gen-dal/proto/gen/dal/v1"
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

	var files []*GeneratedFile

	// Generate code for each message
	for _, msg := range messages {
		content, err := generateMessageCode(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code for %s: %w", msg.TargetMessage.Desc.Name(), err)
		}

		files = append(files, &GeneratedFile{
			Path:    "generated.go", // TODO: Proper path generation
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// generateMessageCode generates the complete Go code for a single message
// by building template data and rendering the file template.
func generateMessageCode(msg *collector.MessageInfo) (string, error) {
	// Build struct data from the message
	structData, err := buildStructData(msg)
	if err != nil {
		return "", err
	}

	// Build template data
	data := TemplateData{
		PackageName: "dal", // TODO: Extract from proto package
		Imports:     []string{}, // TODO: Determine required imports
		Struct:      structData,
	}

	// Render the file template
	return renderTemplate("file.go.tmpl", data)
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
	// Handle repeated fields
	if field.Desc.Cardinality().String() == "repeated" {
		elemType := protoScalarToGo(field.Desc.Kind().String())
		return "[]" + elemType
	}

	return protoScalarToGo(field.Desc.Kind().String())
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
