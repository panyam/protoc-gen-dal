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
	"bytes"
	"embed"
	"text/template"

	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateData contains all data needed to render a complete Go file.
type TemplateData struct {
	PackageName string
	Imports     []string // TODO: Change to []ImportSpec
	Structs     []StructData // Multiple structs per file
}

// StructData contains data for generating a GORM struct.
type StructData struct {
	Name       string      // GORM struct name (e.g., "BookGORM")
	SourceName string      // Source API message name (e.g., "library.v1.Book")
	TableName  string      // Database table name (e.g., "books")
	Fields     []FieldData // Struct fields
}

// FieldData contains data for a single struct field.
type FieldData struct {
	Name       string // Go field name (e.g., "ID", "Title")
	Type       string // Go type (e.g., "string", "int64", "[]string")
	GormTag    string // GORM tag content (e.g., "primaryKey;type:uuid")
	IsOptional bool   // Whether field is marked optional in proto (affects pointer generation)
}

// ConverterFileData contains all data needed to render a converter file.
type ConverterFileData struct {
	PackageName                   string               // Package name (e.g., "gorm")
	Imports                       []common.ImportSpec  // Import specs with optional aliases
	Converters                    []ConverterData      // Converter functions to generate
	HasRepeatedMessageConversions bool                 // Whether any converter has repeated message fields (needs fmt)
}

// ConverterData contains data for generating converter functions.
type ConverterData struct {
	SourceType    string             // Source API type (e.g., "User")
	SourcePkgName string             // Source package name (e.g., "testapi")
	TargetType    string             // Target type (e.g., "UserGORM", "UserWithPermissions")
	FieldMappings []FieldMappingData // Field conversion mappings
}

// ConversionType represents how a field should be converted.
type ConversionType int

const (
	ConvertIgnore                          ConversionType = iota // Field only in target, skip conversion
	ConvertByAssignment                                          // Direct assignment: out.Field = src.Field
	ConvertByTransformer                                         // No-error transformer: out.Field = converter(src.Field)
	ConvertByTransformerWithError                                // Error-returning transformer: out.Field, err = converter(src.Field)
	ConvertByTransformerWithIgnorableError                       // Error-ignoring transformer: out.Field, _ = converter(src.Field)
)

// FieldMappingData contains data for mapping a single field between API and target.
type FieldMappingData struct {
	SourceField              string         // Source field name (e.g., "Birthday")
	TargetField              string         // Target field name (e.g., "Birthday")
	ToTargetConversionType   ConversionType // How to convert source → target (or element conversion for collections)
	FromTargetConversionType ConversionType // How to convert target → source (or element conversion for collections)
	ToTargetCode             string         // Code to convert source to target (for assignment/transformer)
	FromTargetCode           string         // Code to convert target to source (for assignment/transformer)
	ToTargetConverterFunc    string         // Converter function name for ToTarget (e.g., "AuthorToAuthorGORM")
	FromTargetConverterFunc  string         // Converter function name for FromTarget (e.g., "AuthorFromAuthorGORM")
	SourceIsPointer          bool           // Whether source field is a pointer type (needs nil check)
	TargetIsPointer          bool           // Whether target field is a pointer type (affects assignment)
	IsRepeated               bool           // Whether this is a repeated field (needs loop-based conversion)
	IsMap                    bool           // Whether this is a map field (needs loop-based conversion)
	TargetElementType        string         // For repeated/map: Go type of target element/value (e.g., "AuthorGORM")
	SourceElementType        string         // For repeated/map: Go type of source element/value (e.g., "Author")
	SourcePkgName            string         // Source package name (e.g., "api" or "testapi") - needed for type references
}

var tmpl *template.Template

// loadTemplates loads and parses all templates.
// This is called once during initialization.
func loadTemplates() (*template.Template, error) {
	if tmpl != nil {
		return tmpl, nil
	}

	// Create template with helper functions
	t := template.New("").Funcs(template.FuncMap{
		// fieldRef generates the correct field reference expression for converter parameters.
		// For pointer fields: returns "varName.fieldName" (pass pointer as-is)
		// For value fields: returns "&varName.fieldName" (take address for in-place modification)
		"fieldRef": func(varName, fieldName string, isPointer bool) string {
			if isPointer {
				return varName + "." + fieldName
			}
			return "&" + varName + "." + fieldName
		},
	})

	// Parse all template files
	t, err := t.ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	tmpl = t
	return tmpl, nil
}

// renderTemplate executes a template with the given data.
func renderTemplate(name string, data any) (string, error) {
	t, err := loadTemplates()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
