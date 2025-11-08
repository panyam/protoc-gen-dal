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
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateData contains all data needed to render a complete Go file.
type TemplateData struct {
	PackageName string
	Imports     []string
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
	Name    string // Go field name (e.g., "ID", "Title")
	Type    string // Go type (e.g., "string", "int64", "[]string")
	GormTag string // GORM tag content (e.g., "primaryKey;type:uuid")
}

var tmpl *template.Template

// loadTemplates loads and parses all templates.
// This is called once during initialization.
func loadTemplates() (*template.Template, error) {
	if tmpl != nil {
		return tmpl, nil
	}

	// Parse all template files
	t, err := template.ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	tmpl = t
	return tmpl, nil
}

// renderTemplate executes a template with the given data.
func renderTemplate(name string, data interface{}) (string, error) {
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
