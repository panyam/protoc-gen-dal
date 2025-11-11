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
func Generate(messages []*collector.MessageInfo) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Group messages by their source proto file
	fileGroups := common.GroupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one file per proto file
	for protoFile, msgs := range fileGroups {
		content, err := generateFileCode(msgs)
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
	pkgName := common.ExtractPackageName(firstMsg)

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
	kind := field.Desc.Kind().String()

	// Handle map fields (proto maps generate as message types with IsMap() == true)
	if field.Desc.IsMap() {
		// Extract key and value types from the map entry message
		mapEntry := field.Message
		keyField := mapEntry.Fields[0]   // maps always have key at index 0
		valueField := mapEntry.Fields[1] // maps always have value at index 1

		keyType := common.ProtoScalarToGo(keyField.Desc.Kind().String())

		// Check if value is a message type or scalar
		var valueType string
		if valueField.Desc.Kind().String() == "message" {
			// Map value is a message type - use the struct name
			valueType = string(valueField.Message.Desc.Name())
		} else {
			// Map value is a scalar type
			valueType = common.ProtoScalarToGo(valueField.Desc.Kind().String())
		}

		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	}

	// Handle message types (embedded structs, etc.)
	if kind == "message" {
		// For repeated message fields
		if field.Desc.Cardinality().String() == "repeated" {
			return "[]" + string(field.Message.Desc.Name())
		}
		// For singular message fields, use the struct name
		return string(field.Message.Desc.Name())
	}

	// Handle repeated scalar fields
	if field.Desc.Cardinality().String() == "repeated" {
		elemType := common.ProtoScalarToGo(kind)
		return "[]" + elemType
	}

	return common.ProtoScalarToGo(kind)
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
		sourcePkgName := common.ExtractPackageName(msg.SourceMessage)
		importsMap.Add(common.ImportSpec{
			Alias: sourcePkgName,
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

	// Extract source package name for imports (use same logic as extractPackageName)
	sourcePkgName := common.ExtractPackageName(sourceMsg)

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

		mapping := buildFieldMapping(sourceField, targetField, reg, sourcePkgName)
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

	// Check if types match - if so, simple assignment
	if sourceKind == targetKind {
		mapping.ToTargetConversionType = converter.ConvertByAssignment
		mapping.FromTargetConversionType = converter.ConvertByAssignment
		addRenderStrategies(mapping)
		return mapping
	}

	// Handle common type conversions
	// Note: ToTargetCode is for API → Datastore, FromTargetCode is for Datastore → API

	// uint32 (API) ↔ string (Datastore) for IDs
	if sourceKind == "uint32" && targetKind == "string" {
		mapping.ToTargetCode = fmt.Sprintf("strconv.FormatUint(uint64(src.%s), 10)", sourceFieldName)
		mapping.FromTargetCode = fmt.Sprintf("uint32(mustParseUint(src.%s))", targetFieldName)
		mapping.ToTargetConversionType = converter.ConvertByTransformer
		mapping.FromTargetConversionType = converter.ConvertByTransformer
		addRenderStrategies(mapping)
		return mapping
	}

	// Timestamp (API) ↔ int64 (Datastore)
	if sourceKind == "message" && sourceField.Message != nil &&
		string(sourceField.Message.Desc.FullName()) == "google.protobuf.Timestamp" &&
		targetKind == "int64" {
		mapping.ToTargetCode = fmt.Sprintf("timestampToInt64(src.%s)", sourceFieldName)
		mapping.FromTargetCode = fmt.Sprintf("int64ToTimestamp(src.%s)", targetFieldName)
		mapping.ToTargetConversionType = converter.ConvertByTransformer
		mapping.FromTargetConversionType = converter.ConvertByTransformer
		addRenderStrategies(mapping)
		return mapping
	}

	// Default: simple assignment (may fail at compile time if incompatible)
	mapping.ToTargetConversionType = converter.ConvertByAssignment
	mapping.FromTargetConversionType = converter.ConvertByAssignment
	addRenderStrategies(mapping)
	return mapping
}
