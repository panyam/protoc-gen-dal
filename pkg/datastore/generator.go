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
	"fmt"
	"strings"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"google.golang.org/protobuf/compiler/protogen"
)

// GeneratedFile represents a single generated file.
type GeneratedFile struct {
	// Path is the output file path (e.g., "user_datastore.go")
	Path string

	// Content is the generated Go code
	Content string
}

// GenerateResult contains all generated files.
type GenerateResult struct {
	// Files is the list of generated files
	Files []*GeneratedFile
}

// Generate generates Datastore code for the given messages.
//
// This is the main entry point for Datastore code generation. It receives all
// messages collected for the Datastore target and generates:
// - Datastore entity struct definitions with tags
// - Kind() methods
// - LoadKey()/SaveKey() methods for key management
//
// Parameters:
//   - messages: Collected Datastore messages from the collector
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
		// e.g., datastore/user.proto -> user_datastore.go
		filename := generateFilenameFromProto(protoFile)

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// groupMessagesByFile groups messages by their proto file path.
//
// Why group by file?
// - Users organize their proto files logically
// - Generate one Go file per proto file for better organization
// - e.g., all messages in "dal/user.proto" go to "user_datastore.go"
func groupMessagesByFile(messages []*collector.MessageInfo) map[string][]*collector.MessageInfo {
	groups := make(map[string][]*collector.MessageInfo)
	for _, msg := range messages {
		// Get the proto file path from the target message
		filePath := msg.TargetMessage.Desc.ParentFile().Path()
		groups[filePath] = append(groups[filePath], msg)
	}
	return groups
}

// generateFilenameFromProto creates the output filename from a proto file path.
// e.g., "dal/v1/user_datastore.proto" -> "user_datastore.go"
func generateFilenameFromProto(protoPath string) string {
	// Extract base name without extension
	baseName := protoPath
	if idx := strings.LastIndex(baseName, "/"); idx != -1 {
		baseName = baseName[idx+1:]
	}
	if idx := strings.LastIndex(baseName, ".proto"); idx != -1 {
		baseName = baseName[:idx]
	}
	return baseName + ".go"
}

// generateFileCode generates the complete Go code for all messages in a proto file.
func generateFileCode(messages []*collector.MessageInfo) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate")
	}

	// Build template data
	data, err := buildTemplateData(messages)
	if err != nil {
		return "", fmt.Errorf("failed to build template data: %w", err)
	}

	// Execute template
	content, err := executeTemplate(data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return content, nil
}

// buildTemplateData extracts metadata from messages for template rendering.
func buildTemplateData(messages []*collector.MessageInfo) (*TemplateData, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	// Extract package name from first message
	// All messages in same file should have same package
	firstMsg := messages[0].TargetMessage
	pkgName := extractPackageName(firstMsg)

	var structs []*StructData

	// Build struct data for each message
	for _, msgInfo := range messages {
		structData, err := buildStructData(msgInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to build struct data for %s: %w",
				msgInfo.TargetMessage.Desc.Name(), err)
		}
		structs = append(structs, structData)
	}

	return &TemplateData{
		Package: pkgName,
		Imports: []string{
			"cloud.google.com/go/datastore",
		},
		Structs: structs,
	}, nil
}

// buildStructData extracts metadata for a single struct/entity.
func buildStructData(msgInfo *collector.MessageInfo) (*StructData, error) {
	targetMsg := msgInfo.TargetMessage
	structName := string(targetMsg.Desc.Name())

	// Extract fields
	var fields []*FieldData
	for _, field := range targetMsg.Fields {
		fieldData := &FieldData{
			Name:  fieldName(field),
			Type:  fieldType(field),
			Tags:  buildFieldTags(field),
		}
		fields = append(fields, fieldData)
	}

	// Add Key field at the beginning (excluded from datastore properties)
	keyField := &FieldData{
		Name: "Key",
		Type: "*datastore.Key",
		Tags: "`datastore:\"-\"`",
	}
	fields = append([]*FieldData{keyField}, fields...)

	return &StructData{
		Name:   structName,
		Kind:   msgInfo.TableName, // TableName is repurposed for Kind
		Fields: fields,
	}, nil
}

// fieldName converts a proto field name to a Go field name.
// Proto uses snake_case, Go uses PascalCase.
func fieldName(field *protogen.Field) string {
	return field.GoName
}

// fieldType returns the Go type for a proto field.
func fieldType(field *protogen.Field) string {
	// For now, use simple type mapping
	// This will be enhanced later for complex types
	switch field.Desc.Kind().String() {
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
	case "bytes":
		return "[]byte"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return "interface{}" // Fallback for complex types
	}
}

// buildFieldTags creates struct tags for a field.
func buildFieldTags(field *protogen.Field) string {
	// Use snake_case for datastore property names
	propName := string(field.Desc.Name())
	return fmt.Sprintf("`datastore:\"%s\"`", propName)
}

// extractPackageName extracts the Go package name from a message.
func extractPackageName(msg *protogen.Message) string {
	return string(msg.GoIdent.GoImportPath.Ident("").GoName)
}

// GenerateConverters generates converter functions for transforming between
// API messages and Datastore entities.
//
// This generates ToDatastore and FromDatastore converter functions with decorator support:
// - ToDatastore: Converts API message to Datastore entity
// - FromDatastore: Converts Datastore entity back to API message
//
// Parameters:
//   - messages: Collected Datastore messages from the collector
//
// Returns:
//   - GenerateResult containing converter files (*_converters.go)
//   - error if generation fails
func GenerateConverters(messages []*collector.MessageInfo) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Group messages by their source proto file
	fileGroups := groupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one converter file per proto file
	for protoFile, msgs := range fileGroups {
		content, err := generateConverterFileCode(msgs)
		if err != nil {
			return nil, fmt.Errorf("failed to generate converters for %s: %w", protoFile, err)
		}

		// Generate filename based on the proto file
		// e.g., datastore/user.proto -> user_converters.go
		filename := generateConverterFilenameFromProto(protoFile)

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// generateConverterFilenameFromProto creates the converter filename from a proto file path.
// e.g., "datastore/user.proto" -> "user_converters.go"
func generateConverterFilenameFromProto(protoPath string) string {
	// Extract base name without extension
	baseName := protoPath
	if idx := strings.LastIndex(baseName, "/"); idx != -1 {
		baseName = baseName[idx+1:]
	}
	if idx := strings.LastIndex(baseName, ".proto"); idx != -1 {
		baseName = baseName[:idx]
	}
	return baseName + "_converters.go"
}

// generateConverterFileCode generates the complete converter code for all messages in a proto file.
func generateConverterFileCode(messages []*collector.MessageInfo) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate converters for")
	}

	// Extract package name from the first message's target
	packageName := extractPackageName(messages[0].TargetMessage)

	// Build converter data for each Datastore message
	var converters []*ConverterData
	imports := make(map[string]bool)

	for _, msg := range messages {
		// Skip messages without a source (embedded types)
		if msg.SourceMessage == nil {
			continue
		}

		converterData := buildConverterData(msg)
		converters = append(converters, converterData)

		// Add import for source message package
		sourceImportPath := string(msg.SourceMessage.GoIdent.GoImportPath)
		imports[sourceImportPath] = true
	}

	// Build import list
	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}

	// Build template data
	data := &ConverterFileData{
		PackageName: packageName,
		Imports:     importList,
		Converters:  converters,
	}

	// Execute converter template
	content, err := executeConverterTemplate(data)
	if err != nil {
		return "", fmt.Errorf("failed to execute converter template: %w", err)
	}

	return content, nil
}

// buildConverterData builds converter metadata for a single message.
func buildConverterData(msgInfo *collector.MessageInfo) *ConverterData {
	sourceMsg := msgInfo.SourceMessage
	targetMsg := msgInfo.TargetMessage

	sourceName := string(sourceMsg.Desc.Name())
	targetName := string(targetMsg.Desc.Name())

	// Extract source package name for imports
	sourcePkgName := string(sourceMsg.GoIdent.GoImportPath.Ident("").GoName)

	// Build field mappings
	var fieldMappings []*FieldMapping
	for _, targetField := range targetMsg.Fields {
		// Find corresponding source field by name
		var sourceField *protogen.Field
		for _, sf := range sourceMsg.Fields {
			if sf.Desc.Name() == targetField.Desc.Name() {
				sourceField = sf
				break
			}
		}

		if sourceField == nil {
			// Skip fields that don't exist in source (like Key field)
			continue
		}

		mapping := &FieldMapping{
			SourceField: fieldName(sourceField),
			TargetField: fieldName(targetField),
		}
		fieldMappings = append(fieldMappings, mapping)
	}

	return &ConverterData{
		SourceType:    sourceName,
		TargetType:    targetName,
		SourcePkgName: sourcePkgName,
		FieldMappings: fieldMappings,
	}
}
