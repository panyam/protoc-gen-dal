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
	"strings"
	"testing"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/testutil"
	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

func TestDetectPrimaryKeys_SingleKey(t *testing.T) {
	// Create a message with a single primary key
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "User",
						Fields: []testutil.TestField{
							{
								Name:       "id",
								Number:     1,
								TypeName:   "uint32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{Name: "name", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	msg := plugin.Files[0].Messages[0]

	// Test primary key detection
	primaryKeys, err := detectPrimaryKeys(msg)
	if err != nil {
		t.Fatalf("detectPrimaryKeys failed: %v", err)
	}

	if len(primaryKeys) != 1 {
		t.Errorf("Expected 1 primary key, got %d", len(primaryKeys))
	}

	pk := primaryKeys[0]
	if pk.Name != "Id" {
		t.Errorf("Expected primary key name 'Id', got '%s'", pk.Name)
	}
	if pk.Type != "uint32" {
		t.Errorf("Expected primary key type 'uint32', got '%s'", pk.Type)
	}
	if pk.ProtoName != "id" {
		t.Errorf("Expected proto name 'id', got '%s'", pk.ProtoName)
	}
	if pk.ColumnName != "id" {
		t.Errorf("Expected column name 'id', got '%s'", pk.ColumnName)
	}
}

func TestDetectPrimaryKeys_CompositeKey(t *testing.T) {
	// Create a message with composite primary keys
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/book.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "BookEdition",
						Fields: []testutil.TestField{
							{
								Name:       "book_id",
								Number:     1,
								TypeName:   "string",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{
								Name:       "edition_number",
								Number:     2,
								TypeName:   "int32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{Name: "title", Number: 3, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	msg := plugin.Files[0].Messages[0]

	// Test primary key detection
	primaryKeys, err := detectPrimaryKeys(msg)
	if err != nil {
		t.Fatalf("detectPrimaryKeys failed: %v", err)
	}

	if len(primaryKeys) != 2 {
		t.Errorf("Expected 2 primary keys, got %d", len(primaryKeys))
	}

	// Verify first key
	if primaryKeys[0].Name != "BookId" {
		t.Errorf("Expected first key name 'BookId', got '%s'", primaryKeys[0].Name)
	}
	if primaryKeys[0].Type != "string" {
		t.Errorf("Expected first key type 'string', got '%s'", primaryKeys[0].Type)
	}

	// Verify second key
	if primaryKeys[1].Name != "EditionNumber" {
		t.Errorf("Expected second key name 'EditionNumber', got '%s'", primaryKeys[1].Name)
	}
	if primaryKeys[1].Type != "int32" {
		t.Errorf("Expected second key type 'int32', got '%s'", primaryKeys[1].Type)
	}
}

func TestDetectPrimaryKeys_FallbackToId(t *testing.T) {
	// Create a message without explicit primaryKey tag but with "id" field
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/product.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "Product",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "uint32"},
							{Name: "name", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	msg := plugin.Files[0].Messages[0]

	// Test primary key detection
	primaryKeys, err := detectPrimaryKeys(msg)
	if err != nil {
		t.Fatalf("detectPrimaryKeys failed: %v", err)
	}

	if len(primaryKeys) != 1 {
		t.Errorf("Expected 1 primary key (fallback to 'id'), got %d", len(primaryKeys))
	}

	if primaryKeys[0].Name != "Id" {
		t.Errorf("Expected primary key name 'Id', got '%s'", primaryKeys[0].Name)
	}
}

func TestDetectPrimaryKeys_NoPrimaryKey(t *testing.T) {
	// Create a message without any primary key
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/address.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "Address",
						Fields: []testutil.TestField{
							{Name: "street", Number: 1, TypeName: "string"},
							{Name: "city", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	msg := plugin.Files[0].Messages[0]

	// Test primary key detection - should fail
	_, err := detectPrimaryKeys(msg)
	if err == nil {
		t.Error("Expected error when no primary key found, got nil")
	}

	if !strings.Contains(err.Error(), "no primary key found") {
		t.Errorf("Expected error message about no primary key, got: %v", err)
	}
}

func TestBuildDALData_SingleKey(t *testing.T) {
	// Create test message info
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserGorm",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.User",
							Table:  "users",
						},
						Fields: []testutil.TestField{
							{
								Name:       "id",
								Number:     1,
								TypeName:   "uint32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
						},
					},
				},
			},
		},
	})

	msgInfo := &collector.MessageInfo{
		TargetMessage: plugin.Files[0].Messages[0],
	}

	// Build DAL data
	dalData, err := buildDALData(msgInfo)
	if err != nil {
		t.Fatalf("buildDALData failed: %v", err)
	}

	// Verify DAL data
	if dalData.StructName != "UserGORM" {
		t.Errorf("Expected StructName 'UserGORM', got '%s'", dalData.StructName)
	}

	if dalData.DALTypeName != "UserGORMDAL" {
		t.Errorf("Expected DALTypeName 'UserGORMDAL', got '%s'", dalData.DALTypeName)
	}

	if dalData.HasCompositePK {
		t.Error("Expected HasCompositePK to be false for single key")
	}

	if dalData.PKStructName != "" {
		t.Errorf("Expected empty PKStructName for single key, got '%s'", dalData.PKStructName)
	}

	if len(dalData.PrimaryKeys) != 1 {
		t.Errorf("Expected 1 primary key, got %d", len(dalData.PrimaryKeys))
	}
}

func TestBuildDALData_CompositeKey(t *testing.T) {
	// Create test message info with composite key
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/book.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "BookEditionGORM",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.BookEdition",
							Table:  "book_editions",
						},
						Fields: []testutil.TestField{
							{
								Name:       "book_id",
								Number:     1,
								TypeName:   "string",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{
								Name:       "edition_number",
								Number:     2,
								TypeName:   "int32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
						},
					},
				},
			},
		},
	})

	msgInfo := &collector.MessageInfo{
		TargetMessage: plugin.Files[0].Messages[0],
	}

	// Build DAL data
	dalData, err := buildDALData(msgInfo)
	if err != nil {
		t.Fatalf("buildDALData failed: %v", err)
	}

	// Verify composite key handling
	if !dalData.HasCompositePK {
		t.Error("Expected HasCompositePK to be true for composite key")
	}

	if dalData.PKStructName != "BookEditionKey" {
		t.Errorf("Expected PKStructName 'BookEditionKey', got '%s'", dalData.PKStructName)
	}

	if len(dalData.PrimaryKeys) != 2 {
		t.Errorf("Expected 2 primary keys, got %d", len(dalData.PrimaryKeys))
	}
}

func TestGenerateDALFilename_DefaultSuffix(t *testing.T) {
	options := &DALOptions{
		FilenameSuffix: "_dal",
		FilenamePrefix: "",
		OutputDir:      "",
	}

	filename := generateDALFilename("gorm/user.proto", options)
	expected := "user_gorm_dal.go"

	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

func TestGenerateDALFilename_CustomPrefix(t *testing.T) {
	options := &DALOptions{
		FilenameSuffix: "_dal",
		FilenamePrefix: "dal_",
		OutputDir:      "",
	}

	filename := generateDALFilename("gorm/user.proto", options)
	expected := "dal_user_gorm.go"

	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

func TestGenerateDALFilename_WithOutputDir(t *testing.T) {
	options := &DALOptions{
		FilenameSuffix: "_dal",
		FilenamePrefix: "",
		OutputDir:      "dal",
	}

	filename := generateDALFilename("gorm/user.proto", options)
	expected := "dal/user_gorm_dal.go"

	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

func TestGenerateDALFilename_WithOutputDirAndPrefix(t *testing.T) {
	options := &DALOptions{
		FilenameSuffix: "_dal",
		FilenamePrefix: "helpers_",
		OutputDir:      "dal",
	}

	filename := generateDALFilename("gorm/user.proto", options)
	expected := "dal/helpers_user_gorm.go"

	if filename != expected {
		t.Errorf("Expected filename '%s', got '%s'", expected, filename)
	}
}

func TestGenerateDALFileCode_SkipsMessagesWithoutPK(t *testing.T) {
	// Create messages with and without primary keys
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/models.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserGorm",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.User",
							Table:  "users",
						},
						Fields: []testutil.TestField{
							{
								Name:       "id",
								Number:     1,
								TypeName:   "uint32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
						},
					},
					{
						Name: "AddressGorm",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.Address",
							Table:  "addresses",
						},
						Fields: []testutil.TestField{
							{Name: "street", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{TargetMessage: plugin.Files[0].Messages[0]},
		{TargetMessage: plugin.Files[0].Messages[1]},
	}

	// Generate DAL code
	content, err := generateDALFileCode(messages)
	if err != nil {
		t.Fatalf("generateDALFileCode failed: %v", err)
	}

	// Should contain UserGORMDAL but not AddressGORMDAL (no PK)
	if !strings.Contains(content, "UserGORMDAL") {
		t.Error("Expected generated code to contain 'UserGORMDAL'")
	}

	if strings.Contains(content, "AddressGORMDAL") {
		t.Error("Expected generated code to skip 'AddressGORMDAL' (no primary key)")
	}
}

func TestGenerateDALFileCode_EmptyWhenAllMessagesLackPK(t *testing.T) {
	// Create only messages without primary keys
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/address.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "AddressGorm",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.Address",
							Table:  "addresses",
						},
						Fields: []testutil.TestField{
							{Name: "street", Number: 1, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{TargetMessage: plugin.Files[0].Messages[0]},
	}

	// Generate DAL code
	content, err := generateDALFileCode(messages)
	if err != nil {
		t.Fatalf("generateDALFileCode failed: %v", err)
	}

	// Should return empty string when all messages lack primary keys
	if content != "" {
		t.Errorf("Expected empty content when all messages lack primary keys, got: %s", content)
	}
}

func TestGenerateDALFileCode_BasicStructure(t *testing.T) {
	// Create a simple message with primary key
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/product.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "ProductGORM",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.Product",
							Table:  "products",
						},
						Fields: []testutil.TestField{
							{
								Name:       "id",
								Number:     1,
								TypeName:   "uint32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{Name: "name", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{TargetMessage: plugin.Files[0].Messages[0]},
	}

	// Generate DAL code
	content, err := generateDALFileCode(messages)
	if err != nil {
		t.Fatalf("generateDALFileCode failed: %v", err)
	}

	// Verify basic structure
	expectedElements := []string{
		"package v1",
		"type ProductGORMDAL struct",
		"WillCreate func(context.Context, *ProductGORM) error",
		"func (d *ProductGORMDAL) Save(",
		"func (d *ProductGORMDAL) Get(",
		"func (d *ProductGORMDAL) Delete(",
		"func (d *ProductGORMDAL) List(",
		"func (d *ProductGORMDAL) BatchGet(",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected generated code to contain '%s'", expected)
		}
	}
}

func TestGenerateDALFileCode_CreateUpdateMethods(t *testing.T) {
	// Create a message to test Create and Update methods
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/product.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "ProductGORM",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.Product",
							Table:  "products",
						},
						Fields: []testutil.TestField{
							{
								Name:       "id",
								Number:     1,
								TypeName:   "uint32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{Name: "name", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{TargetMessage: plugin.Files[0].Messages[0]},
	}

	// Generate DAL code
	content, err := generateDALFileCode(messages)
	if err != nil {
		t.Fatalf("generateDALFileCode failed: %v", err)
	}

	// Verify Create method exists
	if !strings.Contains(content, "func (d *ProductGORMDAL) Create(") {
		t.Error("Expected generated code to contain Create method")
	}

	// Verify Update method exists
	if !strings.Contains(content, "func (d *ProductGORMDAL) Update(") {
		t.Error("Expected generated code to contain Update method")
	}

	// Verify Create returns error on duplicate
	if !strings.Contains(content, "db.Create(obj).Error") {
		t.Error("Expected Create method to call db.Create")
	}

	// Verify Update checks RowsAffected
	if !strings.Contains(content, "result.RowsAffected == 0") {
		t.Error("Expected Update method to check RowsAffected")
	}

	// Verify Save still has WillCreate hook
	if !strings.Contains(content, "d.WillCreate") {
		t.Error("Expected Save method to call WillCreate hook")
	}
}

func TestGenerateDALFileCode_WithSubdirectory(t *testing.T) {
	// Create a simple message with primary key
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			{
				Name: "test/user.proto",
				Pkg:  "test.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "UserGORM",
						GormOpts: &dalv1.GormOptions{
							Source: "test.v1.User",
							Table:  "users",
						},
						Fields: []testutil.TestField{
							{
								Name:       "id",
								Number:     1,
								TypeName:   "uint32",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{Name: "name", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages := []*collector.MessageInfo{
		{TargetMessage: plugin.Files[0].Messages[0]},
	}

	// Generate DAL code with subdirectory
	options := &DALOptions{
		OutputDir: "dal",
	}
	entityPkgInfo := common.PackageInfo{
		ImportPath: "github.com/test/gen/v1",
		Alias:      "v1",
	}
	content, err := generateDALFileCodeWithOptions(messages, entityPkgInfo, options)
	if err != nil {
		t.Fatalf("generateDALFileCodeWithOptions failed: %v", err)
	}

	// Verify package name changed to subdirectory name
	if !strings.Contains(content, "package dal") {
		t.Error("Expected package name to be 'dal' when OutputDir is specified")
	}

	// Verify import for entity package
	if !strings.Contains(content, `v1 "github.com/test/gen/v1"`) {
		t.Error("Expected import for entity package when OutputDir is specified")
	}

	// Verify entity types are prefixed with package alias
	expectedPrefixed := []string{
		"*v1.UserGORM",
		"v1.UserGORM",
		"[]*v1.UserGORM",
	}

	for _, expected := range expectedPrefixed {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected entity types to be prefixed with 'v1.', missing: %s", expected)
		}
	}

	// Verify no unprefixed entity references
	if strings.Contains(content, "*UserGORM)") || strings.Contains(content, "[]UserGORM") {
		t.Error("Found unprefixed entity type references (should all be prefixed with 'v1.')")
	}
}
