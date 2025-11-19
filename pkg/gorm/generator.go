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
	"log"
	"strings"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/converter"
	"github.com/panyam/protoc-gen-dal/pkg/generator/registry"
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

	// Build message registry for source → target lookups
	// This allows us to find AuthorGORM (or any user-defined name) when we see api.Author in a field
	msgRegistry := common.NewMessageRegistry(messages, buildStructName)

	// Validate that all referenced message types have explicit definitions
	// Users must define all needed types (with flexible naming via 'source' annotation)
	if err := msgRegistry.ValidateMissingTypes(messages); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Collect all embedded types across ALL messages first
	// This ensures we generate each embedded type only once in a shared file
	allEmbeddedTypes := make(map[string]*protogen.Message)
	for _, msg := range messages {
		collectEmbeddedTypes(msg.TargetMessage, allEmbeddedTypes)
	}

	// Filter out embedded types that are already GORM messages
	// (they'll be generated with their own table)
	sharedEmbeddedTypes := make(map[string]*protogen.Message)
	for fullName, embMsg := range allEmbeddedTypes {
		isGormMsg := false
		for _, msg := range messages {
			if msg.TargetMessage == embMsg {
				isGormMsg = true
				break
			}
		}
		if !isGormMsg {
			sharedEmbeddedTypes[fullName] = embMsg
		}
	}

	// Group messages by their source proto file
	fileGroups := common.GroupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one file per proto file (without embedded types)
	for protoFile, msgs := range fileGroups {
		content, err := generateFileCodeWithoutEmbedded(msgs, msgRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code for %s: %w", protoFile, err)
		}

		// Generate filename based on the proto file
		// e.g., gorm/user.proto -> user_gorm.go
		filename := common.GenerateFilenameFromProto(protoFile, "_gorm.go")

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	// Generate a single shared file for all embedded types (if any)
	if len(sharedEmbeddedTypes) > 0 {
		content, err := generateEmbeddedTypesFile(sharedEmbeddedTypes, messages[0].TargetMessage, msgRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedded types: %w", err)
		}

		files = append(files, &GeneratedFile{
			Path:    "_embedded_gorm.go",
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// GenerateConverters generates converter functions for transforming between
// API messages and GORM structs.
//
// This generates ToGORM and FromGORM converter functions with decorator support:
// - ToGORM: converter.Converts API message to GORM struct
// - FromGORM: converter.Converts GORM struct back to API message
//
// Parameters:
//   - messages: Collected GORM messages from the collector
//
// Returns:
//   - GenerateResult containing converter files (*_converters.go)
//   - error if generation fails
func GenerateConverters(messages []*collector.MessageInfo) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Build message registry for source → target lookups
	// This is needed for nested message type resolution during converter generation
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
		// e.g., gorm/user.proto -> user_converters.go
		filename := common.GenerateConverterFilename(protoFile)

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// generateFileCodeWithoutEmbedded generates Go code for messages in a proto file.
// Embedded types are NOT included - they're generated separately in _embedded_gorm.go
func generateFileCodeWithoutEmbedded(messages []*collector.MessageInfo, registry *common.MessageRegistry) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate")
	}

	// Extract package name from the first message's target
	packageName := common.ExtractPackageName(messages[0].TargetMessage)

	// Build struct data for all messages with GORM annotations
	var structs []StructData
	importsMap := make(common.ImportMap)

	// Always add time package (for time.Time fields)
	importsMap.Add(common.ImportSpec{Path: "time"})

	for _, msg := range messages {
		structData, err := buildStructData(msg, registry)
		if err != nil {
			return "", err
		}
		structs = append(structs, structData)

		// Add source package import if needed (for enum types)
		if msg.SourceMessage != nil {
			pkgInfo := common.ExtractPackageInfo(msg.SourceMessage)
			importsMap.Add(common.ImportSpec{
				Alias: pkgInfo.Alias,
				Path:  pkgInfo.ImportPath,
			})
		}
	}

	// Convert imports map to sorted slice
	imports := importsMap.ToSlice()

	// Build template data
	data := TemplateData{
		PackageName: packageName,
		Imports:     imports,
		Structs:     structs,
	}

	// Render the file template
	return renderTemplate("file.go.tmpl", data)
}

// generateEmbeddedTypesFile generates a single file containing all shared embedded types.
// This prevents duplicate type definitions across multiple generated files.
func generateEmbeddedTypesFile(embeddedTypes map[string]*protogen.Message, sampleMsg *protogen.Message, registry *common.MessageRegistry) (string, error) {
	if len(embeddedTypes) == 0 {
		return "", fmt.Errorf("no embedded types to generate")
	}

	// Extract package name (use sample message)
	packageName := common.ExtractPackageName(sampleMsg)

	// Build struct data for embedded types
	var structs []StructData
	for _, embMsg := range embeddedTypes {
		// Build a simple struct for this embedded type (no table name)
		// No field merging for embedded types - just use fields as-is
		// Extract source package alias for enum types
		pkgInfo := common.ExtractPackageInfo(embMsg)
		fields, err := buildFields(embMsg.Fields, pkgInfo.Alias, registry)
		if err != nil {
			return "", err
		}
		structs = append(structs, StructData{
			Name:      buildStructName(embMsg),
			TableName: "", // No table for embedded types
			Fields:    fields,
		})
	}

	// Build template data
	data := TemplateData{
		PackageName: packageName,
		Imports:     []common.ImportSpec{{Path: "time"}}, // time package needed for time.Time fields
		Structs:     structs,
	}

	// Render the file template
	return renderTemplate("file.go.tmpl", data)
}

// generateConverterFileCode generates converter functions for all messages in a proto file.
func generateConverterFileCode(messages []*collector.MessageInfo, msgRegistry *common.MessageRegistry) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate converters for")
	}

	// Extract package name from the first message's target
	packageName := common.ExtractPackageName(messages[0].TargetMessage)

	// Build converter registry to track available converters
	registry := registry.NewConverterRegistry(messages, buildStructName)

	// Build converter data for each GORM message
	var converters []ConverterData
	importsMap := make(common.ImportMap) // Key: import path

	for _, msg := range messages {
		// Skip messages without a source (embedded types)
		if msg.SourceMessage == nil {
			continue
		}

		converterData, err := buildConverterData(msg, registry, msgRegistry)
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

		// Collect custom converter package imports
		common.CollectCustomConverterImports(msg.TargetMessage, importsMap)
	}

	// Build import list using ImportMap's ToSlice method
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
	data := ConverterFileData{
		PackageName:                   packageName,
		Imports:                       importList,
		Converters:                    converters,
		HasRepeatedMessageConversions: hasFmtNeeded,
	}

	// Render the converter file template
	return renderTemplate("converters.go.tmpl", data)
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

// buildStructData extracts struct information from a MessageInfo.
func buildStructData(msg *collector.MessageInfo, registry *common.MessageRegistry) (StructData, error) {
	targetMsg := msg.TargetMessage
	sourceMsg := msg.SourceMessage

	// Build struct name: e.g., "BookGorm" -> "BookGORM"
	structName := buildStructName(targetMsg)

	// Validate field merging (skip_field references, source exists, etc.)
	if err := common.ValidateFieldMerge(sourceMsg, targetMsg, msg.SourceName); err != nil {
		return StructData{}, err
	}

	// Merge source and target fields (implements opt-out field model)
	mergedFields, err := common.MergeSourceFields(sourceMsg, targetMsg)
	if err != nil {
		return StructData{}, fmt.Errorf("failed to merge fields: %w", err)
	}

	// Extract source package alias for enum type references
	var sourcePkgAlias string
	if sourceMsg != nil {
		pkgInfo := common.ExtractPackageInfo(sourceMsg)
		sourcePkgAlias = pkgInfo.Alias
	}

	// Build fields from merged list
	fields, err := buildFields(mergedFields, sourcePkgAlias, registry)
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

// buildConverterData builds converter function data from a MessageInfo.
func buildConverterData(msg *collector.MessageInfo, reg *registry.ConverterRegistry, msgRegistry *common.MessageRegistry) (ConverterData, error) {
	// Extract source type name and package
	sourceTypeName := string(msg.SourceMessage.Desc.Name())
	sourcePkgName := common.ExtractPackageName(msg.SourceMessage)

	// Build GORM type name (e.g., "UserGORM", "UserWithPermissions")
	gormTypeName := buildStructName(msg.TargetMessage)

	// Merge source and target fields (same as buildStructData)
	// This ensures converters use the same fields as the generated struct
	mergedFields, err := common.MergeSourceFields(msg.SourceMessage, msg.TargetMessage)
	if err != nil {
		return ConverterData{}, fmt.Errorf("failed to merge fields for %s: %w", msg.TargetMessage.Desc.Name(), err)
	}

	// Build field mappings between source and GORM with built-in conversions
	var fieldMappings []FieldMappingData

	// Create a map of source fields by name for quick lookup
	sourceFields := make(map[string]*protogen.Field)
	for _, field := range msg.SourceMessage.Fields {
		sourceFields[field.GoName] = field
	}

	// Iterate over merged fields (not target fields directly)
	// This ensures skip_field annotations are respected
	for _, mergedField := range mergedFields {
		// Check if source has a field with the same name
		sourceField, exists := sourceFields[mergedField.GoName]
		if !exists {
			// Field only exists in target (e.g., DeletedAt) - skip, decorator will handle
			continue
		}

		// Generate conversion code based on type compatibility
		mapping := buildFieldConversion(sourceField, mergedField, reg, msgRegistry)
		if mapping == nil {
			// No conversion possible - skip, decorator must handle
			continue
		}

		// Set source package name for type references
		mapping.SourcePkgName = sourcePkgName

		fieldMappings = append(fieldMappings, *mapping)
	}

	// Classify fields by render strategy using shared utility
	// Convert []FieldMappingData to []*FieldMappingData for generic function
	fieldMappingPtrs := make([]*FieldMappingData, len(fieldMappings))
	for i := range fieldMappings {
		fieldMappingPtrs[i] = &fieldMappings[i]
	}
	classified := converter.ClassifyFields(fieldMappingPtrs)

	// Convert back to []FieldMappingData (dereference pointers)
	toTargetInline := make([]FieldMappingData, len(classified.ToTargetInline))
	for i, ptr := range classified.ToTargetInline {
		toTargetInline[i] = *ptr
	}
	toTargetSetter := make([]FieldMappingData, len(classified.ToTargetSetter))
	for i, ptr := range classified.ToTargetSetter {
		toTargetSetter[i] = *ptr
	}
	toTargetLoop := make([]FieldMappingData, len(classified.ToTargetLoop))
	for i, ptr := range classified.ToTargetLoop {
		toTargetLoop[i] = *ptr
	}
	fromTargetInline := make([]FieldMappingData, len(classified.FromTargetInline))
	for i, ptr := range classified.FromTargetInline {
		fromTargetInline[i] = *ptr
	}
	fromTargetSetter := make([]FieldMappingData, len(classified.FromTargetSetter))
	for i, ptr := range classified.FromTargetSetter {
		fromTargetSetter[i] = *ptr
	}
	fromTargetLoop := make([]FieldMappingData, len(classified.FromTargetLoop))
	for i, ptr := range classified.FromTargetLoop {
		fromTargetLoop[i] = *ptr
	}

	return ConverterData{
		SourceType:    sourceTypeName,
		SourcePkgName: sourcePkgName,
		TargetType:    gormTypeName,
		FieldMappings: fieldMappings, // Keep for backward compatibility

		// Classified field groups
		ToTargetInlineFields: toTargetInline,
		ToTargetSetterFields: toTargetSetter,
		ToTargetLoopFields:   toTargetLoop,

		FromTargetInlineFields: fromTargetInline,
		FromTargetSetterFields: fromTargetSetter,
		FromTargetLoopFields:   fromTargetLoop,
	}, nil
}

// addRenderStrategies calculates and adds render strategies to a FieldMappingData.
// This is a thin wrapper around the shared AddRenderStrategies utility.
func addRenderStrategies(mapping *FieldMappingData) {
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
}

// buildFieldConversion generates conversion code for a field pair.
// Returns FieldMappingData with conversion type and pointer information.
func buildFieldConversion(sourceField, targetField *protogen.Field, reg *registry.ConverterRegistry, msgRegistry *common.MessageRegistry) *FieldMappingData {
	sourceKind := sourceField.Desc.Kind().String()
	targetKind := targetField.Desc.Kind().String()
	fieldName := sourceField.GoName

	// Determine if fields are pointers
	// Source (proto): message fields are always pointers, scalars are pointers only if optional
	sourceIsPointer := sourceKind == "message" || sourceField.Desc.HasPresence()

	// Target (GORM): only pointers if explicitly marked with 'optional' keyword
	// Non-optional message fields become value types in GORM
	// HasOptionalKeyword() returns true only for fields with explicit 'optional' in proto
	targetIsPointer := targetField.Desc.HasOptionalKeyword()

	mapping := &FieldMappingData{
		SourceField:     sourceField.GoName,
		TargetField:     targetField.GoName,
		SourceIsPointer: sourceIsPointer,
		TargetIsPointer: targetIsPointer,
	}

	// Step 1: Set default conversion type based on field kind
	// Scalars default to assignment, messages default to transformer with error
	// Maps and repeated fields apply "applicative" style - check contained type

	// Check if this is a map field
	if sourceField.Desc.IsMap() {
		mapping.IsMap = true
		// Map fields: check the value type to determine conversion
		isPrimitive, _ := converter.CheckMapValueType(sourceField)

		if !isPrimitive {
			// map<K, MessageType> - needs loop-based converter for values
			mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
			mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
			// Continue to find converter function for the value type
		} else {
			// map<K, primitive> - direct assignment (copy entire map)
			mapping.ToTargetCode = fmt.Sprintf("src.%s", fieldName)
			mapping.FromTargetCode = fmt.Sprintf("src.%s", fieldName)
			mapping.ToTargetConversionType = converter.ConvertByAssignment
			mapping.FromTargetConversionType = converter.ConvertByAssignment
			// Return early - no further processing needed for primitive maps
			addRenderStrategies(mapping)
			return mapping
		}
	} else if sourceField.Desc.IsList() {
		mapping.IsRepeated = true
		// Repeated fields: check the element type to determine conversion
		isPrimitive, _ := converter.CheckRepeatedElementType(sourceField)

		if !isPrimitive {
			// []MessageType - needs loop-based converter for elements
			mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
			mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
			// Continue to find converter function for the element type
		} else {
			// []primitive - direct assignment (copy entire slice)
			mapping.ToTargetCode = fmt.Sprintf("src.%s", fieldName)
			mapping.FromTargetCode = fmt.Sprintf("src.%s", fieldName)
			mapping.ToTargetConversionType = converter.ConvertByAssignment
			mapping.FromTargetConversionType = converter.ConvertByAssignment
			// Return early - no further processing needed for primitive slices
			addRenderStrategies(mapping)
			return mapping
		}
	} else if sourceKind == "message" || targetKind == "message" {
		// Regular message fields (not map or repeated) default to transformer with error
		mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
		mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
	} else {
		// Scalar types default to assignment
		mapping.ToTargetConversionType = converter.ConvertByAssignment
		mapping.FromTargetConversionType = converter.ConvertByAssignment
	}

	// Step 2: Check for custom converter functions (highest priority - overrides defaults)
	toTargetCode, fromTargetCode := common.ExtractCustomConverters(targetField, fieldName)
	if toTargetCode != "" {
		mapping.ToTargetCode = toTargetCode
		mapping.FromTargetCode = fromTargetCode
		mapping.ToTargetConversionType = converter.ConvertByTransformer
		mapping.FromTargetConversionType = converter.ConvertByTransformer
		addRenderStrategies(mapping)
		return mapping
	}

	// Step 3: Check for known type conversions using the generic type mapping registry
	// This handles Timestamp→int64, Timestamp→time.Time, uint32→string, etc.
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

	// Same types - direct assignment (override message default)
	if sourceKind == targetKind && sourceKind != "message" {
		mapping.ToTargetCode = fmt.Sprintf("src.%s", fieldName)
		mapping.FromTargetCode = fmt.Sprintf("src.%s", fieldName)
		mapping.ToTargetConversionType = converter.ConvertByAssignment
		mapping.FromTargetConversionType = converter.ConvertByAssignment
		addRenderStrategies(mapping)
		return mapping
	}

	// Both are messages - use nested converter if available
	// For repeated/map fields, we look up converter for the element/value type
	if sourceKind == "message" && targetKind == "message" {
		var sourceTypeName, targetTypeName string
		var sourceMsg, targetMsg *protogen.Message
		var targetFieldMsg *protogen.Message // The actual target field's message type

		if mapping.IsRepeated {
			// For repeated fields, get the element type
			sourceMsg = sourceField.Message
			targetFieldMsg = targetField.Message
			// Check if target is a well-known type first
			if _, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); !isWellKnown {
				// Use MessageRegistry to resolve source → target mapping
				targetMsg = msgRegistry.LookupTargetMessage(sourceMsg)
			}
		} else if mapping.IsMap {
			// For map fields, get the value type
			sourceMapEntry := sourceField.Message
			sourceMsg = sourceMapEntry.Fields[1].Message // value field
			targetMapEntry := targetField.Message
			targetFieldMsg = targetMapEntry.Fields[1].Message
			// Check if target is a well-known type first
			if _, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); !isWellKnown {
				// Use MessageRegistry to resolve source → target mapping
				targetMsg = msgRegistry.LookupTargetMessage(sourceMsg)
			}
		} else {
			// Regular message field
			sourceMsg = sourceField.Message
			targetFieldMsg = targetField.Message
			// Check if target is a well-known type first
			if _, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); !isWellKnown {
				// Use MessageRegistry to resolve source → target mapping
				targetMsg = msgRegistry.LookupTargetMessage(sourceMsg)
			}
		}

		// Check if target is a well-known type (e.g., google.protobuf.Any)
		if wkt, isWellKnown := common.GetWellKnownTypeMapping(targetFieldMsg); isWellKnown {
			// Handle well-known type conversion (e.g., Message → Any serialization)
			// For google.protobuf.Any, automatically serialize using protobuf marshaling
			if wkt.ProtoFullName == "google.protobuf.Any" {
				// Build the fully qualified type: *api.Path (not *.Path)
				// Extract package name from source message
				sourcePkgName := common.ExtractPackageName(sourceMsg)
				sourceTypeName := fmt.Sprintf("*%s.%s", sourcePkgName, sourceMsg.Desc.Name())

				if mapping.IsRepeated {
					// For repeated Any fields: []Message → [][]byte
					// Use loop-based conversion with element converter functions
					// SourceElementType should be just the type name (e.g., "WorldChange")
					// since template constructs []*api.WorldChange from SourcePkgName + SourceElementType
					mapping.SourceElementType = string(sourceMsg.Desc.Name())
					mapping.TargetElementType = "[]byte"
					// Use wrapper converters that match standard converter signature (dest, src, decorator)
					mapping.ToTargetConverterFunc = "converters.MessageToAnyBytesConverter"
					// For FromTargetConverterFunc, we need the full type with pointer and package
					mapping.FromTargetConverterFunc = fmt.Sprintf("converters.AnyBytesToMessageConverter[%s]", sourceTypeName)
					mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
					mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
				} else {
					// For single Any fields: Message → []byte
					// Generate direct conversion code
					mapping.ToTargetCode = fmt.Sprintf("converters.MessageToAnyBytes(src.%s)", fieldName)
					mapping.FromTargetCode = fmt.Sprintf("converters.AnyBytesToMessage[%s](src.%s)",
						sourceTypeName, fieldName)
					mapping.ToTargetConversionType = converter.ConvertByTransformerWithError
					mapping.FromTargetConversionType = converter.ConvertByTransformerWithError
				}
				addRenderStrategies(mapping)
				return mapping
			}
			// Other well-known types can be added here
			return nil
		}

		if sourceMsg != nil && targetMsg != nil {
			sourceTypeName = string(sourceMsg.Desc.Name())
			// Use MessageRegistry to get the correct GORM struct name
			targetTypeName = msgRegistry.GetStructName(targetMsg)

			// Check if converter exists for this nested type
			if reg.HasConverter(sourceTypeName, targetTypeName) {
				mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
				mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
				// For repeated/map fields, store the element/value types
				if mapping.IsRepeated || mapping.IsMap {
					mapping.TargetElementType = targetTypeName
					mapping.SourceElementType = sourceTypeName
				}
				// Keep default converter.ConvertByTransformerWithError (already set in Step 1)
				addRenderStrategies(mapping)
				return mapping
			}
			// No converter available - warn user and skip (decorator must handle)
			if mapping.IsRepeated {
				log.Printf("WARNING: Field '%s' is []%s but no converter found for element type.\n",
					fieldName, sourceTypeName)
			} else if mapping.IsMap {
				log.Printf("WARNING: Field '%s' is map<K, %s> but no converter found for value type.\n",
					fieldName, sourceTypeName)
			} else {
				log.Printf("WARNING: Field '%s' has matching message types (%s → %s) but no converter found.\n",
					fieldName, sourceTypeName, targetTypeName)
			}
			log.Printf("         If you want automatic conversion, add 'source' annotation to %s message.\n",
				targetMsg.Desc.Name())
			log.Printf("         Skipping field - handle in decorator function.\n\n")
			return nil
		}
	}

	// Numeric conversions - use casting
	if common.IsNumericKind(sourceKind) && common.IsNumericKind(targetKind) {
		mapping.ToTargetCode = fmt.Sprintf("%s(src.%s)", common.ProtoKindToGoType(targetKind), fieldName)
		mapping.FromTargetCode = fmt.Sprintf("%s(src.%s)", common.ProtoKindToGoType(sourceKind), fieldName)
		mapping.ToTargetConversionType = converter.ConvertByAssignment
		mapping.FromTargetConversionType = converter.ConvertByAssignment
		addRenderStrategies(mapping)
		return mapping
	}

	// No built-in conversion available - log warning and skip
	log.Printf("WARNING: No type conversion found for field %q: %s (%s) → %s (%s).",
		fieldName,
		converter.GetTypeName(sourceField), sourceKind,
		converter.GetTypeName(targetField), targetKind)
	log.Printf("         Field will be skipped in converter - handle in decorator function.")
	return nil
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

// buildFields extracts field information from a list of proto fields.
func buildFields(protoFields []*protogen.Field, sourcePkgName string, registry *common.MessageRegistry) ([]FieldData, error) {
	var fields []FieldData

	for _, field := range protoFields {
		fieldData, err := buildField(field, sourcePkgName, registry)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldData)
	}

	return fields, nil
}

// buildField converts a proto field to FieldData.
func buildField(field *protogen.Field, sourcePkgName string, registry *common.MessageRegistry) (FieldData, error) {
	// converter.Convert field name to Go naming: id -> ID, title -> Title
	goName := field.GoName

	// converter.Convert proto type to Go type using shared utility with GORM-specific naming
	// Pass the registry so it can look up target types for message fields
	goType := common.ProtoFieldToGoType(field, buildStructName, sourcePkgName, registry)

	// Extract GORM tags from column options
	gormTag := extractGormTags(field)

	return FieldData{
		Name:    goName,
		Type:    goType,
		GormTag: gormTag,
	}, nil
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

// isEmbeddedField checks if a field is marked as embedded in GORM tags.
// Embedded fields are generated as value types, not pointers.
func isEmbeddedField(field *protogen.Field) bool {
	opts := field.Desc.Options()
	if opts == nil {
		return false
	}

	v := proto.GetExtension(opts, dalv1.E_Column)
	if v == nil {
		return false
	}

	colOpts, ok := v.(*dalv1.ColumnOptions)
	if !ok || colOpts == nil {
		return false
	}

	// Check if "embedded" is in the gorm_tags
	for _, tag := range colOpts.GormTags {
		if tag == "embedded" || strings.HasPrefix(tag, "embedded:") {
			return true
		}
	}

	return false
}
