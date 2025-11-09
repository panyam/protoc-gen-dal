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
)

// Template data structures

// TemplateData contains all data for generating a complete Go file.
type TemplateData struct {
	// Package is the Go package name
	Package string

	// Imports is the list of import paths
	Imports []string

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

// FieldData contains metadata for a single struct field.
type FieldData struct {
	// Name is the Go field name (e.g., "Name")
	Name string

	// Type is the Go type (e.g., "string")
	Type string

	// Tags is the struct tag string (e.g., "`datastore:\"name\"`")
	Tags string
}

// ImportSpec represents a Go import with optional alias.
type ImportSpec struct {
	Alias string // Optional alias (e.g., "models")
	Path  string // Import path (e.g., "github.com/...")
}

// ConverterFileData contains all data for generating a converter file.
type ConverterFileData struct {
	// PackageName is the Go package name
	PackageName string

	// Imports is the list of import specs with optional aliases
	Imports []ImportSpec

	// Converters is the list of converter functions to generate
	Converters []*ConverterData

	// HasRepeatedMessageConversions indicates if any converter has repeated/map message conversions (needs fmt)
	HasRepeatedMessageConversions bool
}

// ConverterData contains metadata for generating a pair of converter functions.
type ConverterData struct {
	// SourceType is the API message type name (e.g., "User")
	SourceType string

	// TargetType is the Datastore entity type name (e.g., "UserDatastore")
	TargetType string

	// SourcePkgName is the source package name for imports (e.g., "api")
	SourcePkgName string

	// FieldMappings is the list of field conversions
	FieldMappings []*FieldMapping
}

// FieldMapping describes how to convert a single field.
type FieldMapping struct {
	// SourceField is the source field name
	SourceField string

	// TargetField is the target field name
	TargetField string

	// ToTargetCode is the conversion code for API → Datastore (empty string = simple assignment)
	ToTargetCode string

	// FromTargetCode is the conversion code for Datastore → API (empty string = simple assignment)
	FromTargetCode string

	// IsRepeated indicates this is a repeated field (slice/array)
	IsRepeated bool

	// IsMap indicates this is a map field
	IsMap bool

	// TargetElementType is the element type for repeated fields or value type for maps
	TargetElementType string

	// SourceElementType is the source element type for repeated fields or value type for maps
	SourceElementType string

	// ToTargetConverterFunc is the converter function name for nested message conversions (API → Datastore)
	ToTargetConverterFunc string

	// FromTargetConverterFunc is the converter function name for nested message conversions (Datastore → API)
	FromTargetConverterFunc string
}

// Embedded templates

//go:embed templates/file.go.tmpl
var fileTemplate string

//go:embed templates/converters.go.tmpl
var converterTemplate string

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
	tmpl, err := template.New("converters").Parse(converterTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse converter template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute converter template: %w", err)
	}

	return buf.String(), nil
}
