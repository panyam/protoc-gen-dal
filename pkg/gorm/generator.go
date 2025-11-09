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

// GenerateConverters generates converter functions for transforming between
// API messages and GORM structs.
//
// This generates ToGORM and FromGORM converter functions with decorator support:
// - ToGORM: Converts API message to GORM struct
// - FromGORM: Converts GORM struct back to API message
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
		// e.g., gorm/user.proto -> user_converters.go
		filename := generateConverterFilenameFromProto(protoFile)

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

// generateConverterFilenameFromProto creates the converter filename from a proto file path.
// e.g., "gorm/user.proto" -> "user_converters.go"
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

// generateConverterFileCode generates converter functions for all messages in a proto file.
func generateConverterFileCode(messages []*collector.MessageInfo) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate converters for")
	}

	// Extract package name from the first message's target
	packageName := extractPackageName(messages[0].TargetMessage)

	// Build converter registry to track available converters
	registry := newConverterRegistry(messages)

	// Build converter data for each GORM message
	var converters []ConverterData
	imports := make(map[string]bool)

	for _, msg := range messages {
		// Skip messages without a source (embedded types)
		if msg.SourceMessage == nil {
			continue
		}

		converterData := buildConverterData(msg, registry)
		converters = append(converters, converterData)

		// Add import for source message package
		sourceImportPath := string(msg.SourceMessage.GoIdent.GoImportPath)
		imports[sourceImportPath] = true

		// Collect custom converter package imports
		collectCustomImports(msg.TargetMessage, imports)
	}

	// Build import list
	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}

	// Build template data
	data := ConverterFileData{
		PackageName: packageName,
		Imports:     importList,
		Converters:  converters,
	}

	// Render the converter file template
	return renderTemplate("converters.go.tmpl", data)
}

// collectCustomImports collects import paths from custom converter functions.
func collectCustomImports(msg *protogen.Message, imports map[string]bool) {
	for _, field := range msg.Fields {
		opts := field.Desc.Options()
		if opts == nil {
			continue
		}

		v := proto.GetExtension(opts, dalv1.E_Column)
		if v == nil {
			continue
		}

		colOpts, ok := v.(*dalv1.ColumnOptions)
		if !ok || colOpts == nil {
			continue
		}

		// Add to_func package
		if colOpts.ToFunc != nil && colOpts.ToFunc.Package != "" {
			imports[colOpts.ToFunc.Package] = true
		}

		// Add from_func package
		if colOpts.FromFunc != nil && colOpts.FromFunc.Package != "" {
			imports[colOpts.FromFunc.Package] = true
		}
	}
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

// converterRegistry tracks which converter functions are being generated.
// Used to determine if nested converter calls are available.
type converterRegistry struct {
	converters map[string]bool // key: "SourceType:GormType"
}

// newConverterRegistry creates a new converter registry from messages.
func newConverterRegistry(messages []*collector.MessageInfo) *converterRegistry {
	reg := &converterRegistry{
		converters: make(map[string]bool),
	}

	for _, msg := range messages {
		if msg.SourceMessage == nil {
			continue
		}
		sourceType := string(msg.SourceMessage.Desc.Name())
		gormType := buildStructName(msg.TargetMessage)
		key := fmt.Sprintf("%s:%s", sourceType, gormType)
		reg.converters[key] = true
	}

	return reg
}

// hasConverter checks if a converter exists for the given source and gorm types.
func (r *converterRegistry) hasConverter(sourceType, gormType string) bool {
	key := fmt.Sprintf("%s:%s", sourceType, gormType)
	return r.converters[key]
}

// buildConverterData builds converter function data from a MessageInfo.
func buildConverterData(msg *collector.MessageInfo, registry *converterRegistry) ConverterData {
	// Extract source type name and package
	sourceTypeName := string(msg.SourceMessage.Desc.Name())
	sourcePkgName := extractPackageName(msg.SourceMessage)

	// Build GORM type name (e.g., "UserGORM", "UserWithPermissions")
	gormTypeName := buildStructName(msg.TargetMessage)

	// Build field mappings between source and GORM with built-in conversions
	var fieldMappings []FieldMappingData

	// Create a map of source fields by name for quick lookup
	sourceFields := make(map[string]*protogen.Field)
	for _, field := range msg.SourceMessage.Fields {
		sourceFields[field.GoName] = field
	}

	for _, targetField := range msg.TargetMessage.Fields {
		// Check if source has a field with the same name
		sourceField, exists := sourceFields[targetField.GoName]
		if !exists {
			// Field only exists in target (e.g., DeletedAt) - skip, decorator will handle
			continue
		}

		// Generate conversion code based on type compatibility
		mapping := buildFieldConversion(sourceField, targetField, registry)
		if mapping == nil {
			// No conversion possible - skip, decorator must handle
			continue
		}

		fieldMappings = append(fieldMappings, *mapping)
	}

	return ConverterData{
		SourceType:    sourceTypeName,
		SourcePkgName: sourcePkgName,
		TargetType:    gormTypeName,
		FieldMappings: fieldMappings,
	}
}

// buildFieldConversion generates conversion code for a field pair.
// Returns FieldMappingData with conversion type and pointer information.
func buildFieldConversion(sourceField, targetField *protogen.Field, registry *converterRegistry) *FieldMappingData {
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

	if sourceKind == "message" || targetKind == "message" {
		mapping.ToTargetConversionType = ConvertByTransformerWithError
		mapping.FromTargetConversionType = ConvertByTransformerWithError
	} else {
		// Scalar types default to assignment
		mapping.ToTargetConversionType = ConvertByAssignment
		mapping.FromTargetConversionType = ConvertByAssignment
	}

	// Step 2: Check for custom converter functions (highest priority - overrides defaults)
	toTargetCode, fromTargetCode := extractCustomConverters(targetField, fieldName)
	if toTargetCode != "" {
		mapping.ToTargetCode = toTargetCode
		mapping.FromTargetCode = fromTargetCode
		mapping.ToTargetConversionType = ConvertByTransformer
		mapping.FromTargetConversionType = ConvertByTransformer
		return mapping
	}

	// Step 3: Handle specific conversion patterns

	// Same types - direct assignment (override message default)
	if sourceKind == targetKind && sourceKind != "message" {
		mapping.ToTargetCode = fmt.Sprintf("src.%s", fieldName)
		mapping.FromTargetCode = fmt.Sprintf("src.%s", fieldName)
		mapping.ToTargetConversionType = ConvertByAssignment
		mapping.FromTargetConversionType = ConvertByAssignment
		return mapping
	}

	// Timestamp (message) to int64 - built-in transformer without error
	if sourceKind == "message" && targetKind == "int64" {
		if sourceField.Message != nil && string(sourceField.Message.Desc.FullName()) == "google.protobuf.Timestamp" {
			mapping.ToTargetCode = fmt.Sprintf("timestampToInt64(src.%s)", fieldName)
			mapping.FromTargetCode = fmt.Sprintf("int64ToTimestamp(src.%s)", fieldName)
			mapping.ToTargetConversionType = ConvertByTransformer
			mapping.FromTargetConversionType = ConvertByTransformer
			return mapping
		}
	}

	// Both are messages - use nested converter if available
	if sourceKind == "message" && targetKind == "message" {
		if sourceField.Message != nil && targetField.Message != nil {
			sourceTypeName := string(sourceField.Message.Desc.Name())
			targetTypeName := buildStructName(targetField.Message)

			// Check if converter exists for this nested type
			if registry.hasConverter(sourceTypeName, targetTypeName) {
				mapping.ToTargetConverterFunc = fmt.Sprintf("%sTo%s", sourceTypeName, targetTypeName)
				mapping.FromTargetConverterFunc = fmt.Sprintf("%sFrom%s", sourceTypeName, targetTypeName)
				// Keep default ConvertByTransformerWithError (already set in Step 1)
				return mapping
			}
			// No converter available - warn user and skip (decorator must handle)
			log.Printf("WARNING: Field '%s' has matching message types (%s → %s) but no converter found.\n",
				fieldName, sourceTypeName, targetTypeName)
			log.Printf("         If you want automatic conversion, add 'source' annotation to %s message.\n",
				targetField.Message.Desc.Name())
			log.Printf("         Skipping field - handle in decorator function.\n\n")
			return nil
		}
	}

	// Numeric conversions - use casting
	if isNumericKind(sourceKind) && isNumericKind(targetKind) {
		mapping.ToTargetCode = fmt.Sprintf("%s(src.%s)", protoKindToGoType(targetKind), fieldName)
		mapping.FromTargetCode = fmt.Sprintf("%s(src.%s)", protoKindToGoType(sourceKind), fieldName)
		mapping.ToTargetConversionType = ConvertByAssignment
		mapping.FromTargetConversionType = ConvertByAssignment
		return mapping
	}

	// No built-in conversion available
	return nil
}

// extractCustomConverters extracts custom converter functions from column options.
// Returns conversion code for to_func and from_func if specified.
func extractCustomConverters(field *protogen.Field, fieldName string) (toTargetCode, fromTargetCode string) {
	opts := field.Desc.Options()
	if opts == nil {
		return "", ""
	}

	// Get column options
	v := proto.GetExtension(opts, dalv1.E_Column)
	if v == nil {
		return "", ""
	}

	colOpts, ok := v.(*dalv1.ColumnOptions)
	if !ok || colOpts == nil {
		return "", ""
	}

	// Extract to_func
	if colOpts.ToFunc != nil && colOpts.ToFunc.Function != "" {
		pkgAlias := colOpts.ToFunc.Alias
		if pkgAlias == "" {
			// Use last segment of package path as alias
			pkgAlias = getPackageAlias(colOpts.ToFunc.Package)
		}
		toTargetCode = fmt.Sprintf("%s.%s(src.%s)", pkgAlias, colOpts.ToFunc.Function, fieldName)
	}

	// Extract from_func
	if colOpts.FromFunc != nil && colOpts.FromFunc.Function != "" {
		pkgAlias := colOpts.FromFunc.Alias
		if pkgAlias == "" {
			pkgAlias = getPackageAlias(colOpts.FromFunc.Package)
		}
		fromTargetCode = fmt.Sprintf("%s.%s(src.%s)", pkgAlias, colOpts.FromFunc.Function, fieldName)
	}

	return toTargetCode, fromTargetCode
}

// getPackageAlias returns the default alias for a package path.
// E.g., "github.com/myapp/converters" -> "converters"
func getPackageAlias(pkgPath string) string {
	if idx := strings.LastIndex(pkgPath, "/"); idx != -1 {
		return pkgPath[idx+1:]
	}
	return pkgPath
}

// isNumericKind checks if a proto kind is numeric.
func isNumericKind(kind string) bool {
	numericKinds := map[string]bool{
		"int32":    true,
		"int64":    true,
		"uint32":   true,
		"uint64":   true,
		"sint32":   true,
		"sint64":   true,
		"fixed32":  true,
		"fixed64":  true,
		"sfixed32": true,
		"sfixed64": true,
		"float":    true,
		"double":   true,
	}
	return numericKinds[kind]
}

// protoKindToGoType converts a proto kind to its Go type for casting.
func protoKindToGoType(kind string) string {
	switch kind {
	case "int32", "sint32", "sfixed32":
		return "int32"
	case "int64", "sint64", "sfixed64":
		return "int64"
	case "uint32", "fixed32":
		return "uint32"
	case "uint64", "fixed64":
		return "uint64"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return kind
	}
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
