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
	"strings"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"google.golang.org/protobuf/compiler/protogen"
)

// DALOptions contains configuration for DAL helper generation
type DALOptions struct {
	FilenameSuffix   string // e.g., "_dal" -> "world_gorm_dal.go"
	FilenamePrefix   string // e.g., "dal_" -> "dal_world_gorm.go"
	OutputDir        string // e.g., "dal" -> files go to "gen/gorm/dal/" (relative to main output)
	EntityImportPath string // e.g., "github.com/example/gen/gorm" (auto-detected if empty)
}

// PrimaryKeyField represents a primary key field in a message
type PrimaryKeyField struct {
	Name       string // Go field name (e.g., "Id", "BookId")
	ProtoName  string // Proto field name (e.g., "id", "book_id")
	Type       string // Go type (e.g., "string", "int32")
	ColumnName string // Database column name (from tags or snake_case of proto name)
}

// DALData holds the template data for DAL helper generation
type DALData struct {
	StructName     string            // e.g., "WorldGORM"
	DALTypeName    string            // e.g., "WorldGORMDAL"
	PrimaryKeys    []PrimaryKeyField // Primary key fields (in order)
	HasCompositePK bool              // Whether there are multiple primary keys
	PKStructName   string            // Composite key struct name (e.g., "WorldKey")
}

// GenerateDALHelpers generates DAL helper methods for GORM messages.
//
// This generates Save, Get, Delete, List, and BatchGet methods for each message:
// - Save: Upsert operation with WillCreate hook
// - Get: Fetch by primary key(s)
// - Delete: Delete by primary key(s)
// - List: Fetch multiple records using a query
// - BatchGet: Fetch multiple records by primary key values
//
// Parameters:
//   - messages: Collected GORM messages from the collector
//   - options: Configuration for filename generation
//
// Returns:
//   - GenerateResult containing DAL helper files
//   - error if generation fails
func GenerateDALHelpers(messages []*collector.MessageInfo, options *DALOptions) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Group messages by their source proto file
	fileGroups := common.GroupMessagesByFile(messages)

	var files []*GeneratedFile

	// Generate one DAL file per proto file
	for protoFile, msgs := range fileGroups {
		// Extract entity package info from the first message
		entityPkgInfo := common.ExtractPackageInfo(msgs[0].TargetMessage)

		// Override with explicit entity import path if provided
		if options.EntityImportPath != "" {
			entityPkgInfo.ImportPath = options.EntityImportPath
			// Extract package name from path (last component)
			entityPkgInfo.Alias = common.GetPackageAlias(options.EntityImportPath)
		}

		content, err := generateDALFileCodeWithOptions(msgs, entityPkgInfo, options)
		if err != nil {
			return nil, fmt.Errorf("failed to generate DAL helpers for %s: %w", protoFile, err)
		}

		// Skip empty files (messages without primary keys)
		if content == "" {
			continue
		}

		// Generate filename based on the proto file and options
		filename := generateDALFilename(protoFile, options)

		files = append(files, &GeneratedFile{
			Path:    filename,
			Content: content,
		})
	}

	return &GenerateResult{Files: files}, nil
}

// generateDALFilename generates the filename for DAL helpers based on options
func generateDALFilename(protoFile string, options *DALOptions) string {
	// Get base filename from proto file (e.g., "gorm/user.proto" -> "user_gorm")
	base := common.GenerateFilenameFromProto(protoFile, "_gorm.go")
	base = strings.TrimSuffix(base, ".go")

	// Apply prefix/suffix
	var filename string
	if options.FilenamePrefix != "" {
		filename = options.FilenamePrefix + base + ".go"
	} else {
		filename = base + options.FilenameSuffix + ".go"
	}

	// Apply output directory if specified
	if options.OutputDir != "" {
		filename = options.OutputDir + "/" + filename
	}

	return filename
}

// generateDALFileCode generates the DAL helper code for messages in a proto file
// (backward compatibility wrapper)
func generateDALFileCode(messages []*collector.MessageInfo) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate DAL helpers for")
	}
	entityPkgInfo := common.ExtractPackageInfo(messages[0].TargetMessage)
	return generateDALFileCodeWithOptions(messages, entityPkgInfo, &DALOptions{})
}

// generateDALFileCodeWithOptions generates the DAL helper code with subdirectory support
func generateDALFileCodeWithOptions(messages []*collector.MessageInfo, entityPkgInfo common.PackageInfo, options *DALOptions) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to generate DAL helpers for")
	}

	// Extract package name from the first message's target
	entityPackageName := common.ExtractPackageName(messages[0].TargetMessage)

	// Build DAL data for messages that:
	// 1. Have GenerateDAL set to true
	// 2. Have primary keys
	var dals []DALData
	for _, msg := range messages {
		// Skip messages that don't want DAL generation
		if !msg.GenerateDAL {
			continue
		}

		dalData, err := buildDALData(msg)
		if err != nil {
			// Skip messages that don't have primary keys
			// (e.g., embedded types or messages without id fields)
			continue
		}
		dals = append(dals, dalData)
	}

	// If no messages have primary keys, skip generating this file
	if len(dals) == 0 {
		return "", nil
	}

	// Determine package name, imports, and prefix based on OutputDir
	packageName := entityPackageName
	imports := common.ImportMap{}
	entityPrefix := ""
	gormAlias := "gorm"

	// Always add standard imports
	imports.Add(common.ImportSpec{Path: "context"})
	imports.Add(common.ImportSpec{Path: "errors"})

	if options.OutputDir != "" {
		// When OutputDir is specified, use subdirectory name as package
		// and import the entity package
		packageName = strings.TrimSuffix(options.OutputDir, "/")
		if entityPkgInfo.ImportPath != "" {
			imports.Add(common.ImportSpec{
				Alias: entityPkgInfo.Alias,
				Path:  entityPkgInfo.ImportPath,
			})
		}
		entityPrefix = entityPkgInfo.Alias + "."

		// Handle naming clash: if entity package is "gorm", alias the gorm library
		if entityPkgInfo.Alias == "gorm" {
			gormAlias = "gormlib"
			imports.Add(common.ImportSpec{
				Alias: gormAlias,
				Path:  "gorm.io/gorm",
			})
		} else {
			imports.Add(common.ImportSpec{Path: "gorm.io/gorm"})
		}
	} else {
		// No subdirectory, no entity import needed
		imports.Add(common.ImportSpec{Path: "gorm.io/gorm"})
	}

	// Build template data
	data := DALTemplateData{
		PackageName:  packageName,
		DALs:         dals,
		Imports:      imports.ToSlice(),
		EntityPrefix: entityPrefix,
		GormAlias:    gormAlias,
	}

	// Render the DAL template
	return renderTemplate("dal.go.tmpl", data)
}

// buildDALData builds the template data for a single message's DAL helper
func buildDALData(msg *collector.MessageInfo) (DALData, error) {
	structName := buildStructName(msg.TargetMessage)
	dalTypeName := structName + "DAL"

	// Detect primary key fields
	primaryKeys, err := detectPrimaryKeys(msg.TargetMessage)
	if err != nil {
		return DALData{}, fmt.Errorf("failed to detect primary keys for %s: %w", structName, err)
	}

	hasCompositePK := len(primaryKeys) > 1
	pkStructName := ""
	if hasCompositePK {
		// Generate composite key struct name by removing "GORM" suffix
		// e.g., "BookEditionGORM" -> "BookEditionKey"
		pkStructName = strings.TrimSuffix(structName, "GORM") + "Key"
	}

	return DALData{
		StructName:     structName,
		DALTypeName:    dalTypeName,
		PrimaryKeys:    primaryKeys,
		HasCompositePK: hasCompositePK,
		PKStructName:   pkStructName,
	}, nil
}

// detectPrimaryKeys detects primary key fields from GORM tags or defaults to "id" field
func detectPrimaryKeys(msg *protogen.Message) ([]PrimaryKeyField, error) {
	var primaryKeys []PrimaryKeyField

	// First pass: look for fields with "primaryKey" in gorm_tags
	for _, field := range msg.Fields {
		if hasPrimaryKeyTag(field) {
			pkField := PrimaryKeyField{
				Name:       field.GoName,
				ProtoName:  string(field.Desc.Name()),
				Type:       getGoType(field),
				ColumnName: common.GetColumnName(field),
			}
			primaryKeys = append(primaryKeys, pkField)
		}
	}

	// If no primary keys found, default to "id" field
	if len(primaryKeys) == 0 {
		for _, field := range msg.Fields {
			if strings.ToLower(string(field.Desc.Name())) == "id" {
				pkField := PrimaryKeyField{
					Name:       field.GoName,
					ProtoName:  string(field.Desc.Name()),
					Type:       getGoType(field),
					ColumnName: "id",
				}
				primaryKeys = append(primaryKeys, pkField)
				break
			}
		}
	}

	if len(primaryKeys) == 0 {
		return nil, fmt.Errorf("no primary key found (no 'primaryKey' tag and no 'id' field)")
	}

	return primaryKeys, nil
}

// hasPrimaryKeyTag checks if a field has "primaryKey" in its gorm_tags
func hasPrimaryKeyTag(field *protogen.Field) bool {
	// Get column options from the field
	opts := common.GetColumnOptions(field)
	if opts == nil {
		return false
	}

	// Check if any gorm tag contains "primaryKey"
	for _, tag := range opts.GormTags {
		if tag == "primaryKey" {
			return true
		}
	}

	return false
}

// getGoType returns the Go type for a field (simplified version)
func getGoType(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case 9: // string
		return "string"
	case 3: // int64
		return "int64"
	case 5: // int32
		return "int32"
	case 13: // uint32
		return "uint32"
	case 4: // uint64
		return "uint64"
	case 8: // bool
		return "bool"
	default:
		return "interface{}"
	}
}

// DALTemplateData is the root template data for DAL file generation
type DALTemplateData struct {
	PackageName  string
	DALs         []DALData
	Imports      []common.ImportSpec // Standard imports (always includes gorm)
	EntityPrefix string              // Prefix for entity types (e.g., "v1." or "")
	GormAlias    string              // Alias for gorm library (e.g., "gorm" or "gormlib")
}
