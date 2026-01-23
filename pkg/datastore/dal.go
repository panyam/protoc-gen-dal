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
	"sort"
	"strings"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// DALOptions contains configuration for Datastore DAL helper generation.
type DALOptions struct {
	FilenameSuffix   string // e.g., "_dal" -> "user_datastore_dal.go"
	FilenamePrefix   string // e.g., "dal_" -> "dal_user_datastore.go"
	OutputDir        string // e.g., "dal" -> files go to "gen/datastore/dal/"
	EntityImportPath string // e.g., "github.com/example/gen/datastore" (auto-detected if empty)
}

// DALData holds the template data for a single Datastore DAL helper.
type DALData struct {
	StructName  string // e.g., "UserDatastore"
	DALTypeName string // e.g., "UserDatastoreDAL"
	HasIDField  bool   // Whether the struct has an "id" field for convenience methods
	IDFieldType string // Type of the ID field (usually "string")
	HasStringID bool   // Whether the struct has a string Id field (for key derivation in Put)
}

// DALTemplateData is the root template data for DAL file generation.
type DALTemplateData struct {
	PackageName   string
	DALs          []DALData
	Imports       []common.ImportSpec
	EntityPrefix  string // Prefix for entity types (e.g., "ds." or "")
	DatastoreLib  string // Datastore library reference (e.g., "dslib" or "datastore")
}

// GenerateDALHelpers generates DAL helper methods for Datastore messages.
//
// This generates Put, Get, Delete, GetMulti, PutMulti, DeleteMulti, Query, and Count
// methods for each message, along with ID-based convenience methods.
//
// Parameters:
//   - messages: Collected Datastore messages from the collector
//   - options: Configuration for filename generation
//
// Returns:
//   - GenerateResult containing DAL helper files
//   - error if generation fails
func GenerateDALHelpers(messages []*collector.MessageInfo, options *DALOptions) (*GenerateResult, error) {
	if len(messages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Filter to only messages with GenerateDAL = true
	var dalMessages []*collector.MessageInfo
	for _, msg := range messages {
		if msg.GenerateDAL {
			dalMessages = append(dalMessages, msg)
		}
	}

	if len(dalMessages) == 0 {
		return &GenerateResult{Files: []*GeneratedFile{}}, nil
	}

	// Group messages by their source proto file
	fileGroups := common.GroupMessagesByFile(dalMessages)

	// Get sorted proto file paths for deterministic output
	protoFiles := make([]string, 0, len(fileGroups))
	for protoFile := range fileGroups {
		protoFiles = append(protoFiles, protoFile)
	}
	sort.Strings(protoFiles)

	var files []*GeneratedFile

	// Generate one DAL file per proto file
	for _, protoFile := range protoFiles {
		msgs := fileGroups[protoFile]
		// Extract entity package info from the first message
		entityPkgInfo := common.ExtractPackageInfo(msgs[0].TargetMessage)

		// Override with explicit entity import path if provided
		if options.EntityImportPath != "" {
			entityPkgInfo.ImportPath = options.EntityImportPath
			entityPkgInfo.Alias = common.GetPackageAlias(options.EntityImportPath)
		}

		content, err := generateDALFileCodeWithOptions(msgs, entityPkgInfo, options)
		if err != nil {
			return nil, fmt.Errorf("failed to generate DAL helpers for %s: %w", protoFile, err)
		}

		// Skip empty files
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

// generateDALFilename generates the filename for DAL helpers based on options.
func generateDALFilename(protoFile string, options *DALOptions) string {
	// Get base filename from proto file (e.g., "datastore/user.proto" -> "user_datastore")
	base := common.GenerateFilenameFromProto(protoFile, "_datastore.go")
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

// generateDALFileCodeWithOptions generates the DAL helper code for messages.
func generateDALFileCodeWithOptions(messages []*collector.MessageInfo, entityPkgInfo common.PackageInfo, options *DALOptions) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// Extract package name from the first message's target
	entityPackageName := common.ExtractPackageName(messages[0].TargetMessage)

	// Build DAL data for each message
	var dals []DALData
	for _, msg := range messages {
		dalData := buildDALData(msg)
		dals = append(dals, dalData)
	}

	if len(dals) == 0 {
		return "", nil
	}

	// Determine package name, imports, and prefix based on OutputDir
	packageName := entityPackageName
	imports := common.ImportMap{}
	entityPrefix := ""
	datastoreLib := "datastore" // Default library reference

	// Always add standard imports
	imports.Add(common.ImportSpec{Path: "context"})

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
		// Add datastore library with alias to avoid collision with entity package
		imports.Add(common.ImportSpec{Alias: "dslib", Path: "cloud.google.com/go/datastore"})
		datastoreLib = "dslib"
	} else {
		// No OutputDir - entity package is in the same package, no alias needed for dslib
		imports.Add(common.ImportSpec{Path: "cloud.google.com/go/datastore"})
	}

	// Build template data
	data := DALTemplateData{
		PackageName:  packageName,
		DALs:         dals,
		Imports:      imports.ToSlice(),
		EntityPrefix: entityPrefix,
		DatastoreLib: datastoreLib,
	}

	// Render the DAL template
	return renderTemplate("dal.go.tmpl", data)
}

// buildDALData builds the template data for a single message's DAL helper.
func buildDALData(msg *collector.MessageInfo) DALData {
	structName := buildStructName(msg.TargetMessage)
	dalTypeName := structName + "DAL"

	// Check for ID field
	hasIDField := false
	idFieldType := "string"
	for _, field := range msg.TargetMessage.Fields {
		if strings.ToLower(string(field.Desc.Name())) == "id" {
			hasIDField = true
			idFieldType = getGoType(field)
			break
		}
	}

	return DALData{
		StructName:  structName,
		DALTypeName: dalTypeName,
		HasIDField:  hasIDField,
		IDFieldType: idFieldType,
		HasStringID: hasIDField && idFieldType == "string",
	}
}

// getGoType returns the Go type for a protogen.Field.
func getGoType(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.BoolKind:
		return "bool"
	default:
		return "interface{}"
	}
}
