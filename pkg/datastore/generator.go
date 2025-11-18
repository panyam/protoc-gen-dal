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
	"log"
	"strings"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/converter"
	"github.com/panyam/protoc-gen-dal/pkg/generator/registry"
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

		// Add source package import if needed (for enum types)
		if msgInfo.SourceMessage != nil {
			sourceImportPath := string(msgInfo.SourceMessage.GoIdent.GoImportPath)
			// Extract base path without ;packagename suffix for import
			if idx := strings.LastIndex(sourceImportPath, ";"); idx != -1 {
				sourceImportPath = sourceImportPath[:idx]
			}
			// Use last path component as alias (e.g., "api" from ".../go/api")
			sourcePkgAlias := common.GetPackageAlias(sourceImportPath)
			importsMap.Add(common.ImportSpec{
				Alias: sourcePkgAlias,
				Path:  sourceImportPath,
			})
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
	mergedFields := common.MergeSourceFields(sourceMsg, targetMsg)

	// Extract source package alias for enum type references
	// Use import path's last component as alias (e.g., "api" from ".../go/api")
	var sourcePkgName string
	if sourceMsg != nil {
		sourceImportPath := string(sourceMsg.GoIdent.GoImportPath)
		if idx := strings.LastIndex(sourceImportPath, ";"); idx != -1 {
			sourceImportPath = sourceImportPath[:idx]
		}
		sourcePkgName = common.GetPackageAlias(sourceImportPath)
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
func buildFieldTags(field *protogen.Field) string {
	// Use snake_case for datastore property names
	propName := string(field.Desc.Name())
	return fmt.Sprintf("`datastore:\"%s\"`", propName)
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
	fileGroups := common.GroupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one converter file per proto file
	for protoFile, msgs := range fileGroups {
		content, err := generateConverterFileCode(msgs)
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
func generateConverterFileCode(messages []*collector.MessageInfo) (string, error) {
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
	var converters []*ConverterData
	importsMap := make(common.ImportMap) // Key: import path

	for _, msg := range messages {
		// Skip messages without a source (embedded types)
		if msg.SourceMessage == nil {
			continue
		}

		converterData := buildConverterData(msg, reg)
		converters = append(converters, converterData)

		// Add import for source message package with alias
		sourceImportPath := string(msg.SourceMessage.GoIdent.GoImportPath)
		// Extract base path without ;packagename suffix for import
		if idx := strings.LastIndex(sourceImportPath, ";"); idx != -1 {
			sourceImportPath = sourceImportPath[:idx]
		}
		// Use last path component as alias (e.g., "api" from ".../go/api")
		sourcePkgAlias := common.GetPackageAlias(sourceImportPath)
		importsMap.Add(common.ImportSpec{
			Alias: sourcePkgAlias,
			Path:  sourceImportPath,
		})
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
func buildConverterData(msgInfo *collector.MessageInfo, reg *registry.ConverterRegistry) *ConverterData {
	sourceMsg := msgInfo.SourceMessage
	targetMsg := msgInfo.TargetMessage

	sourceName := string(sourceMsg.Desc.Name())
	targetName := string(targetMsg.Desc.Name())

	// Extract source package alias for imports
	// Use import path's last component as alias (e.g., "api" from ".../go/api")
	sourceImportPath := string(sourceMsg.GoIdent.GoImportPath)
	if idx := strings.LastIndex(sourceImportPath, ";"); idx != -1 {
		sourceImportPath = sourceImportPath[:idx]
	}
	sourcePkgName := common.GetPackageAlias(sourceImportPath)

	// Merge source and target fields (same as buildStructData)
	// This ensures converters use the same fields as the generated struct
	mergedFields := common.MergeSourceFields(sourceMsg, targetMsg)

	// Build field mappings
	var fieldMappings []*FieldMapping
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

		mapping := buildFieldMapping(sourceField, mergedField, reg, sourcePkgName)

		// Skip fields marked as ConvertIgnore (no conversion available, decorator handles)
		if mapping.ToTargetConversionType == converter.ConvertIgnore {
			continue
		}

		fieldMappings = append(fieldMappings, mapping)
	}

	// Classify fields by render strategy
	var toTargetInline, toTargetSetter, toTargetLoop []*FieldMapping
	var fromTargetInline, fromTargetSetter, fromTargetLoop []*FieldMapping

	for _, mapping := range fieldMappings {
		// Classify ToTarget direction
		switch mapping.ToTargetRenderStrategy {
		case converter.StrategyInlineValue:
			toTargetInline = append(toTargetInline, mapping)
		case converter.StrategySetterSimple, converter.StrategySetterTransform,
			converter.StrategySetterWithError, converter.StrategySetterIgnoreError:
			toTargetSetter = append(toTargetSetter, mapping)
		case converter.StrategyLoopRepeated, converter.StrategyLoopMap:
			toTargetLoop = append(toTargetLoop, mapping)
		}

		// Classify FromTarget direction
		switch mapping.FromTargetRenderStrategy {
		case converter.StrategyInlineValue:
			fromTargetInline = append(fromTargetInline, mapping)
		case converter.StrategySetterSimple, converter.StrategySetterTransform,
			converter.StrategySetterWithError, converter.StrategySetterIgnoreError:
			fromTargetSetter = append(fromTargetSetter, mapping)
		case converter.StrategyLoopRepeated, converter.StrategyLoopMap:
			fromTargetLoop = append(fromTargetLoop, mapping)
		}
	}

	return &ConverterData{
		SourceType:    sourceName,
		TargetType:    targetName,
		SourcePkgName: sourcePkgName,
		FieldMappings: fieldMappings,

		ToTargetInlineFields: toTargetInline,
		ToTargetSetterFields: toTargetSetter,
		ToTargetLoopFields:   toTargetLoop,

		FromTargetInlineFields: fromTargetInline,
		FromTargetSetterFields: fromTargetSetter,
		FromTargetLoopFields:   fromTargetLoop,
	}
}

// addRenderStrategies calculates and populates render strategies for a field mapping.
// This determines how the field conversion should be rendered in the template based on
// the ConversionType and field characteristics (pointer, repeated, map, etc.).
func addRenderStrategies(mapping *FieldMapping) {
	if mapping == nil {
		return
	}

	// Build characteristics for ToTarget direction (API → Datastore)
	chars := converter.FieldCharacteristics{
		IsPointer:          mapping.SourceIsPointer,
		IsRepeated:         mapping.IsRepeated,
		IsMap:              mapping.IsMap,
		HasMessageElements: mapping.IsRepeated && mapping.ToTargetConverterFunc != "",
		HasMessageValues:   mapping.IsMap && mapping.ToTargetConverterFunc != "",
	}

	mapping.ToTargetRenderStrategy = converter.DetermineRenderStrategy(
		mapping.ToTargetConversionType,
		chars,
	)

	// Build characteristics for FromTarget direction (Datastore → API)
	charsFrom := chars
	charsFrom.IsPointer = mapping.TargetIsPointer
	mapping.FromTargetRenderStrategy = converter.DetermineRenderStrategy(
		mapping.FromTargetConversionType,
		charsFrom,
	)
}

// buildFieldMapping creates a field mapping with type conversion if needed.
func buildFieldMapping(sourceField, targetField *protogen.Field, reg *registry.ConverterRegistry, sourcePkgName string) *FieldMapping {
	sourceFieldName := fieldName(sourceField)
	targetFieldName := fieldName(targetField)

	sourceKind := sourceField.Desc.Kind().String()
	targetKind := targetField.Desc.Kind().String()

	mapping := &FieldMapping{
		SourceField:     sourceFieldName,
		TargetField:     targetFieldName,
		SourceIsPointer: sourceField.Desc.HasPresence(), // Proto3 optional fields are pointers
		TargetIsPointer: targetField.Desc.HasPresence(), // Datastore entity fields with HasPresence are pointers
		SourcePkgName:   sourcePkgName,                  // Package name for template rendering
	}

	// Check if this is a map field
	if sourceField.Desc.IsMap() {
		mapping.IsMap = true
		// Map fields: check the value type to determine conversion
		mapEntry := sourceField.Message
		valueField := mapEntry.Fields[1] // value field is always index 1
		valueKind := valueField.Desc.Kind().String()

		if valueKind == "message" {
			// map<K, MessageType> - needs loop-based converter for values
			sourceMsg := valueField.Message
			targetMapEntry := targetField.Message
			targetValueField := targetMapEntry.Fields[1]
			targetMsg := targetValueField.Message

			if sourceMsg != nil && targetMsg != nil {
				sourceTypeName := string(sourceMsg.Desc.Name())
				targetTypeName := string(targetMsg.Desc.Name())

				// Check if converter exists for this nested type
				if reg.HasConverter(sourceTypeName, targetTypeName) {
					mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
					mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
					mapping.TargetElementType = targetTypeName
					mapping.SourceElementType = sourceTypeName
					mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
					mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
				}
			}
			addRenderStrategies(mapping)
			return mapping
		} else {
			// map<K, primitive> - direct assignment (copy entire map)
			mapping.ToTargetCode = fmt.Sprintf("src.%s", sourceFieldName)
			mapping.FromTargetCode = fmt.Sprintf("src.%s", targetFieldName)
			mapping.ToTargetConversionType = converter.ConvertByAssignment
			mapping.FromTargetConversionType = converter.ConvertByAssignment
			addRenderStrategies(mapping)
			return mapping
		}
	} else if sourceField.Desc.IsList() {
		mapping.IsRepeated = true
		// Repeated fields: check the element type to determine conversion
		elementKind := sourceKind // The field's own kind is the element kind for repeated

		if elementKind == "message" {
			// []MessageType - needs loop-based converter for elements
			sourceMsg := sourceField.Message
			targetMsg := targetField.Message

			if sourceMsg != nil && targetMsg != nil {
				sourceTypeName := string(sourceMsg.Desc.Name())
				targetTypeName := string(targetMsg.Desc.Name())

				// Check if converter exists for this nested type
				if reg.HasConverter(sourceTypeName, targetTypeName) {
					mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
					mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
					mapping.TargetElementType = targetTypeName
					mapping.SourceElementType = sourceTypeName
					mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
					mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
				}
			}
			addRenderStrategies(mapping)
			return mapping
		} else {
			// []primitive - direct assignment (copy entire slice)
			mapping.ToTargetCode = fmt.Sprintf("src.%s", sourceFieldName)
			mapping.FromTargetCode = fmt.Sprintf("src.%s", targetFieldName)
			mapping.ToTargetConversionType = converter.ConvertByAssignment
			mapping.FromTargetConversionType = converter.ConvertByAssignment
			addRenderStrategies(mapping)
			return mapping
		}
	}

	// Handle common type conversions using the generic type mapping registry
	// This checks for known type conversions (Timestamp→int64, Timestamp→time.Time, uint32→string, etc.)
	toCode, fromCode, convType, targetIsPtr := converter.BuildConversionCode(sourceField, targetField)
	if toCode != "" && fromCode != "" {
		mapping.ToTargetCode = toCode
		mapping.FromTargetCode = fromCode
		mapping.ToTargetConversionType = convType
		mapping.FromTargetConversionType = convType
		// Apply pointer override if specified
		if targetIsPtr != nil {
			mapping.TargetIsPointer = *targetIsPtr
		}
		addRenderStrategies(mapping)
		return mapping
	}

	// Check if types match - if so, simple assignment
	if sourceKind == targetKind {
		mapping.ToTargetCode = fmt.Sprintf("src.%s", sourceFieldName)
		mapping.FromTargetCode = fmt.Sprintf("src.%s", targetFieldName)
		mapping.ToTargetConversionType = converter.ConvertByAssignment
		mapping.FromTargetConversionType = converter.ConvertByAssignment
		addRenderStrategies(mapping)
		return mapping
	}

	// If we got here, we have incompatible types with no known conversion
	// Log warning and skip field (decorator must handle it)
	log.Printf("WARNING: No type conversion found for field %q: %s (%s) → %s (%s).",
		sourceFieldName,
		converter.GetTypeName(sourceField), sourceKind,
		converter.GetTypeName(targetField), targetKind)
	log.Printf("         Field will be skipped in converter - handle in decorator function.")

	// Skip this field - decorator must handle it
	mapping.ToTargetConversionType = converter.ConvertIgnore
	mapping.FromTargetConversionType = converter.ConvertIgnore
	// NOTE: Do NOT call addRenderStrategies - we want to skip this field entirely
	return mapping
}
