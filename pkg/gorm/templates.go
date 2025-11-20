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
	"github.com/panyam/protoc-gen-dal/pkg/generator/converter"
	"github.com/panyam/protoc-gen-dal/pkg/generator/types"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateData contains all data needed to render a complete Go file.
type TemplateData struct {
	PackageName string
	Imports     []common.ImportSpec // Import specifications with optional aliases
	Structs     []StructData        // Multiple structs per file
}

// StructData contains data for generating a GORM struct.
type StructData struct {
	Name       string      // GORM struct name (e.g., "BookGORM")
	SourceName string      // Source API message name (e.g., "library.v1.Book")
	TableName  string      // Database table name (e.g., "books")
	Fields     []FieldData // Struct fields
}

type FieldData = types.FieldData
type ConverterFileData = types.ConverterFileData

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
		// Render strategy helpers for ToTarget direction
		"isInlineValue": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategyInlineValue
		},
		"isSetterSimple": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategySetterSimple
		},
		"isSetterTransform": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategySetterTransform
		},
		"isSetterWithError": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategySetterWithError
		},
		"isSetterIgnoreError": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategySetterIgnoreError
		},
		"isLoopRepeated": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategyLoopRepeated
		},
		"isLoopMap": func(strategy converter.FieldRenderStrategy) bool {
			return strategy == converter.StrategyLoopMap
		},
		// Convenience helpers for checking if error handling is needed
		"needsErrorCheck": func(convType converter.ConversionType) bool {
			return convType == converter.ConvertByTransformerWithError
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
