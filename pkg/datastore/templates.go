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
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/converter"
	"github.com/panyam/protoc-gen-dal/pkg/generator/types"
)

// Template data structures

// TemplateData contains all data for generating a complete Go file.
type TemplateData struct {
	// Package is the Go package name
	Package string

	// Imports is the list of import specifications with optional aliases
	Imports []common.ImportSpec

	// Structs is the list of entity structs to generate
	Structs []*StructData
}

// StructData contains metadata for generating a single entity struct.
type StructData struct {
	// Name is the struct name (e.g., "UserDatastore")
	Name string

	// Kind is the Datastore kind name (e.g., "User")
	Kind string

	// Fields is the list of struct fields
	Fields []*FieldData
}

type FieldData = types.FieldData
type ConverterFileData = types.ConverterFileData

// Embedded templates

//go:embed templates/file.go.tmpl
var fileTemplate string

//go:embed templates/converters.go.tmpl
var converterTemplate string

//go:embed templates/dal.go.tmpl
var dalTemplate string

// executeTemplate executes the file template with the given data.
func executeTemplate(data *TemplateData) (string, error) {
	tmpl, err := template.New("file").Parse(fileTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// executeConverterTemplate executes the converter template with the given data.
func executeConverterTemplate(data *ConverterFileData) (string, error) {
	// Create template with helper functions
	tmpl := template.New("converters").Funcs(template.FuncMap{
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

	tmpl, err := tmpl.Parse(converterTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse converter template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute converter template: %w", err)
	}

	return buf.String(), nil
}

// renderTemplate renders a template by name with the given data.
// This is used by the DAL generator to render the DAL template.
func renderTemplate(name string, data interface{}) (string, error) {
	var templateContent string
	switch name {
	case "dal.go.tmpl":
		templateContent = dalTemplate
	default:
		return "", fmt.Errorf("unknown template: %s", name)
	}

	tmpl, err := template.New(name).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
