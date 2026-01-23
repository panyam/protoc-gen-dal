// Copyright 2025 Sri Panyam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/converter"
)

// GeneratedFile represents a generated code file with its path and content.
type GeneratedFile struct {
	// Path is the output file path (e.g., "user_gorm.go", "user_datastore.go")
	Path string

	// Content is the generated Go code
	Content string
}

// GenerateResult contains all generated files.
type GenerateResult struct {
	// Files is the list of generated files
	Files []*GeneratedFile
}

// ConverterData contains metadata for generating converter functions.
// This structure is shared across all generator targets (GORM, Datastore, etc.).
type ConverterData struct {
	// SourceType is the API message type name (e.g., "User")
	SourceType string

	// TargetType is the target entity type name (e.g., "UserGORM", "UserDatastore")
	TargetType string

	// SourcePkgName is the source package name for imports (e.g., "api", "testapi")
	SourcePkgName string

	// FieldMappings is the list of field conversions (for backward compatibility)
	FieldMappings []*converter.FieldMapping

	// Field groups by render strategy (for struct literal + setters pattern)
	ToTargetInlineFields []*converter.FieldMapping // Fields for struct literal initialization (ToTarget)
	ToTargetSetterFields []*converter.FieldMapping // Fields needing setter statements (ToTarget)
	ToTargetLoopFields   []*converter.FieldMapping // Fields needing loop-based conversion (ToTarget)

	FromTargetInlineFields []*converter.FieldMapping // Fields for struct literal initialization (FromTarget)
	FromTargetSetterFields []*converter.FieldMapping // Fields needing setter statements (FromTarget)
	FromTargetLoopFields   []*converter.FieldMapping // Fields needing loop-based conversion (FromTarget)
}

// FieldData contains data for a single struct field.
type FieldData struct {
	Name       string // Go field name (e.g., "ID", "Title")
	Type       string // Go type (e.g., "string", "int64", "[]string")
	Tags       string // struct tag content (e.g., "primaryKey;type:uuid")
	IsOptional bool   // Whether field is marked optional in proto (affects pointer generation)
	IsMap      bool   // Whether field is a map type (e.g., map[string]int64)
}

// ConverterFileData contains all data for generating a converter file.
type ConverterFileData struct {
	// PackageName is the Go package name
	PackageName string

	// Imports is the list of import specs with optional aliases
	Imports []common.ImportSpec

	// Converters is the list of converter functions to generate
	Converters []*ConverterData

	// HasRepeatedMessageConversions indicates if any converter has repeated/map message conversions (needs fmt)
	HasRepeatedMessageConversions bool
}
