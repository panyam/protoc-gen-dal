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
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/converter"
	"github.com/panyam/protoc-gen-dal/pkg/generator/registry"
	"github.com/panyam/protoc-gen-dal/pkg/generator/types"
	"google.golang.org/protobuf/compiler/protogen"
)

// GeneratedFile is an alias for the shared type
type GeneratedFile = types.GeneratedFile
type GenerateResult = types.GenerateResult

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
//
// buildStructName generates the Datastore struct name from the target message name.
// For Datastore, we keep the name as-is (e.g., "UserDatastore" stays "UserDatastore")
func buildStructName(msg *protogen.Message) string {
	return string(msg.Desc.Name())
}

func Generate(messages []*collector.MessageInfo) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Build message registry for source → target lookups
	// This allows us to find AuthorDatastore (or any user-defined name) when we see api.Author in a field
	msgRegistry := common.NewMessageRegistry(messages, buildStructName)

	// Validate that all referenced message types have explicit definitions
	// Users must define all needed types (with flexible naming via 'source' annotation)
	if err := msgRegistry.ValidateMissingTypes(messages); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Group messages by their source proto file
	fileGroups := common.GroupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one file per proto file
	for protoFile, msgs := range fileGroups {
		content, err := generateFileCode(msgs, msgRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code for %s: %w", protoFile, err)
		}

		// Generate filename based on the proto file
		// e.g., datastore/user.proto -> user_datastore.go
		filename := common.GenerateFilenameFromProto(protoFile, ".go")

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// generateFileCode generates the complete Go code for all messages in a proto file.
func generateFileCode(messages []*collector.MessageInfo, registry *common.MessageRegistry) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate")
	}

	// Build template data
	data, err := buildTemplateData(messages, registry)
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
func buildTemplateData(messages []*collector.MessageInfo, registry *common.MessageRegistry) (*TemplateData, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	// Extract package name from first message
	// All messages in same file should have same package
	firstMsg := messages[0].TargetMessage
	pkgName := common.ExtractPackageName(firstMsg)

	var structs []*StructData
	importsMap := make(common.ImportMap)

	// Always add required packages
	importsMap.Add(common.ImportSpec{Path: "time"})
	importsMap.Add(common.ImportSpec{Path: "cloud.google.com/go/datastore"})

	// Build struct data for each message
	for _, msgInfo := range messages {
		structData, err := buildStructData(msgInfo, registry)
		if err != nil {
			return nil, fmt.Errorf("failed to build struct data for %s: %w",
				msgInfo.TargetMessage.Desc.Name(), err)
		}
		structs = append(structs, structData)

		// Add source package import only if actually needed (for enum types)
		// Check if any field type references the source package
		if msgInfo.SourceMessage != nil {
			pkgInfo := common.ExtractPackageInfo(msgInfo.SourceMessage)
			if pkgInfo.Alias != "" {
				prefix := pkgInfo.Alias + "."
				for _, field := range structData.Fields {
					if strings.Contains(field.Type, prefix) {
						importsMap.Add(common.ImportSpec{
							Alias: pkgInfo.Alias,
							Path:  pkgInfo.ImportPath,
						})
						break
					}
				}
			}
		}
	}

	// Convert imports map to sorted slice
	imports := importsMap.ToSlice()

	return &TemplateData{
		Package: pkgName,
		Imports: imports,
		Structs: structs,
	}, nil
}

// buildStructData extracts metadata for a single struct/entity.
func buildStructData(msgInfo *collector.MessageInfo, registry *common.MessageRegistry) (*StructData, error) {
	targetMsg := msgInfo.TargetMessage
	sourceMsg := msgInfo.SourceMessage
	structName := string(targetMsg.Desc.Name())

	// Validate field merging (skip_field references, source exists, etc.)
	if err := common.ValidateFieldMerge(sourceMsg, targetMsg, msgInfo.SourceName); err != nil {
		return nil, err
	}

	// Merge source and target fields (implements opt-out field model)
	mergedFields, err := common.MergeSourceFields(sourceMsg, targetMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to merge fields: %w", err)
	}

	// Extract source package alias for enum type references
	var sourcePkgName string
	if sourceMsg != nil {
		pkgInfo := common.ExtractPackageInfo(sourceMsg)
		sourcePkgName = pkgInfo.Alias
	}

	// Extract fields from merged list
	var fields []*FieldData
	for _, field := range mergedFields {
		fieldData := &FieldData{
			Name: fieldName(field),
			Type: fieldType(field, sourcePkgName, registry),
			Tags: buildFieldTags(field),
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
func fieldType(field *protogen.Field, sourcePkgName string, registry *common.MessageRegistry) string {
	// Use shared utility with Datastore-specific struct naming
	// This now handles google.protobuf.Timestamp → time.Time automatically
	// The registry allows looking up target types for message fields (e.g., api.Author → AuthorDatastore)
	return common.ProtoFieldToGoType(field, buildStructName, sourcePkgName, registry)
}

// buildFieldTags creates struct tags for a field.
// It reads datastore_tags from the column annotation and joins them with the field name.
// Example: datastore_tags: ["noindex", "omitempty"] generates `datastore:"field_name,noindex,omitempty"`
func buildFieldTags(field *protogen.Field) string {
	// Use snake_case for datastore property names
	propName := string(field.Desc.Name())

	// Get datastore_tags from column options
	colOpts := common.GetColumnOptions(field)
	if colOpts != nil && len(colOpts.DatastoreTags) > 0 {
		// Check for "-" tag (ignore field)
		for _, tag := range colOpts.DatastoreTags {
			if tag == "-" {
				return "`datastore:\"-\"`"
			}
		}
		// Join field name with additional tags
		tags := append([]string{propName}, colOpts.DatastoreTags...)
		return fmt.Sprintf("`datastore:\"%s\"`", joinDatastoreTags(tags))
	}

	return fmt.Sprintf("`datastore:\"%s\"`", propName)
}

// joinDatastoreTags joins datastore tags with commas.
func joinDatastoreTags(tags []string) string {
	var result []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return strings.Join(result, ",")
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

	// Build message registry for source → target type lookups
	msgRegistry := common.NewMessageRegistry(messages, buildStructName)

	// Group messages by their source proto file
	fileGroups := common.GroupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one converter file per proto file
	for protoFile, msgs := range fileGroups {
		content, err := generateConverterFileCode(msgs, msgRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to generate converters for %s: %w", protoFile, err)
		}

		// Generate filename based on the proto file
		// e.g., datastore/user.proto -> user_converters.go
		filename := common.GenerateConverterFilename(protoFile)

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// generateConverterFileCode generates the complete converter code for all messages in a proto file.
func generateConverterFileCode(messages []*collector.MessageInfo, msgRegistry *common.MessageRegistry) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate converters for")
	}

	// Extract package name from the first message's target
	packageName := common.ExtractPackageName(messages[0].TargetMessage)

	// Build converter registry for nested message conversions
	// Datastore uses raw message names (no suffix transformation like GORM)
	datastoreNameFunc := func(msg *protogen.Message) string {
		return string(msg.Desc.Name())
	}
	reg := registry.NewConverterRegistry(messages, datastoreNameFunc)

	// Build converter data for each Datastore message
	var converters []*types.ConverterData
	importsMap := make(common.ImportMap) // Key: import path

	for _, msg := range messages {
		// Skip messages without a source (embedded types)
		if msg.SourceMessage == nil {
			continue
		}

		converterData, err := buildConverterData(msg, reg, msgRegistry)
		if err != nil {
			return "", fmt.Errorf("failed to build converter data for %s: %w", msg.TargetMessage.Desc.Name(), err)
		}
		converters = append(converters, converterData)

		// Add import for source message package with alias
		pkgInfo := common.ExtractPackageInfo(msg.SourceMessage)
		importsMap.Add(common.ImportSpec{
			Alias: pkgInfo.Alias,
			Path:  pkgInfo.ImportPath,
		})

		// Collect custom converter package imports (new for Datastore!)
		common.CollectCustomConverterImports(msg.TargetMessage, importsMap)
	}

	// Build import list
	importList := importsMap.ToSlice()

	// Check if we need fmt import (for repeated/map message conversions)
	hasFmtNeeded := false
	for _, conv := range converters {
		for _, field := range conv.FieldMappings {
			if (field.IsRepeated || field.IsMap) && field.ToTargetConverterFunc != "" {
				hasFmtNeeded = true
				break
			}
		}
		if hasFmtNeeded {
			break
		}
	}

	// Build template data
	data := &ConverterFileData{
		PackageName:                   packageName,
		Imports:                       importList,
		Converters:                    converters,
		HasRepeatedMessageConversions: hasFmtNeeded,
	}

	// Execute converter template
	content, err := executeConverterTemplate(data)
	if err != nil {
		return "", fmt.Errorf("failed to execute converter template: %w", err)
	}

	return content, nil
}

// buildConverterData builds converter metadata for a single message.
func buildConverterData(msgInfo *collector.MessageInfo, reg *registry.ConverterRegistry, msgRegistry *common.MessageRegistry) (*types.ConverterData, error) {
	sourceMsg := msgInfo.SourceMessage
	targetMsg := msgInfo.TargetMessage

	sourceName := string(sourceMsg.Desc.Name())
	targetName := string(targetMsg.Desc.Name())

	// Extract source package alias for imports
	pkgInfo := common.ExtractPackageInfo(sourceMsg)
	sourcePkgName := pkgInfo.Alias

	// Merge source and target fields (same as buildStructData)
	// This ensures converters use the same fields as the generated struct
	mergedFields, err := common.MergeSourceFields(sourceMsg, targetMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to merge fields: %w", err)
	}

	// Build field mappings
	var fieldMappings []*converter.FieldMapping
	for _, mergedField := range mergedFields {
		// Find corresponding source field by name
		var sourceField *protogen.Field
		for _, sf := range sourceMsg.Fields {
			if sf.Desc.Name() == mergedField.Desc.Name() {
				sourceField = sf
				break
			}
		}

		if sourceField == nil {
			// Skip fields that don't exist in source (like Key field added by Datastore)
			continue
		}

		mapping := buildFieldMapping(sourceField, mergedField, reg, sourcePkgName, msgRegistry)

		// Skip fields marked as ConvertIgnore (no conversion available, decorator handles)
		if mapping.ToTargetConversionType == converter.ConvertIgnore {
			continue
		}

		fieldMappings = append(fieldMappings, mapping)
	}

	// Classify fields by render strategy using shared utility
	classified := converter.ClassifyFields(fieldMappings)

	// Filter out oneof members from FromTarget lists.
	// Proto oneofs don't expose fields directly in struct literals, so we can't
	// generate automatic conversion code for them. Users must handle these in decorators.
	fromInline := filterNonOneofFields(classified.FromTargetInline)
	fromSetter := filterNonOneofFields(classified.FromTargetSetter)
	fromLoop := filterNonOneofFields(classified.FromTargetLoop)

	return &types.ConverterData{
		SourceType:    sourceName,
		TargetType:    targetName,
		SourcePkgName: sourcePkgName,
		FieldMappings: fieldMappings,

		ToTargetInlineFields: classified.ToTargetInline,
		ToTargetSetterFields: classified.ToTargetSetter,
		ToTargetLoopFields:   classified.ToTargetLoop,

		FromTargetInlineFields: fromInline,
		FromTargetSetterFields: fromSetter,
		FromTargetLoopFields:   fromLoop,
	}, nil
}

// filterNonOneofFields removes oneof members from a field mapping slice.
// This is needed because proto oneofs can't be set via direct struct field assignment.
func filterNonOneofFields(fields []*converter.FieldMapping) []*converter.FieldMapping {
	result := make([]*converter.FieldMapping, 0, len(fields))
	for _, f := range fields {
		if !f.SourceIsOneofMember {
			result = append(result, f)
		}
	}
	return result
}

// addRenderStrategies calculates and adds render strategies to a FieldMapping.
// This is a thin wrapper around the shared AddRenderStrategies utility.
func addRenderStrategies(mapping *converter.FieldMapping) {
	if mapping == nil {
		return
	}

	// Use shared utility to calculate render strategies
	toTargetStrategy, fromTargetStrategy := converter.AddRenderStrategies(
		mapping.ToTargetConversionType,
		mapping.FromTargetConversionType,
		mapping.SourceIsPointer,
		mapping.TargetIsPointer,
		mapping.IsRepeated,
		mapping.IsMap,
		mapping.ToTargetConverterFunc != "",
		mapping.FromTargetConverterFunc != "",
	)

	mapping.ToTargetRenderStrategy = toTargetStrategy
	mapping.FromTargetRenderStrategy = fromTargetStrategy

	// Oneof fields cannot be initialized inline in proto struct literals.
	// For FromTarget direction (Target → Proto), force oneof members to setter strategy.
	// This is because proto messages with oneofs don't expose oneof fields as named struct fields.
	if mapping.SourceIsOneofMember && mapping.FromTargetRenderStrategy == converter.StrategyInlineValue {
		mapping.FromTargetRenderStrategy = converter.StrategySetterSimple
	}
}

// buildFieldMapping creates a field mapping with type conversion if needed.
// This is a thin wrapper around the shared BuildFieldMapping function.
func buildFieldMapping(sourceField, targetField *protogen.Field, reg *registry.ConverterRegistry, sourcePkgName string, msgRegistry *common.MessageRegistry) *converter.FieldMapping {
	return converter.BuildFieldMapping(sourceField, targetField, reg, msgRegistry, sourcePkgName, addRenderStrategies)
}
