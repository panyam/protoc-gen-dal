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
	"github.com/panyam/protoc-gen-dal/pkg/generator/testutil"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// TestGenerateGORM_SimpleMessage tests that a simple message generates
// a correct GORM struct with basic fields
//
// This is the first test that drives the GORM generator implementation.
// It verifies that we can generate:
// - A properly named struct (BookGORM)
// - Basic fields with correct GORM tags
// - A TableName() method
func TestGenerateGORM_SimpleMessage(t *testing.T) {
	// Given: A simple Book message with basic fields
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			// API proto
			{
				Name: "library/v1/book.proto",
				Pkg:  "library.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "Book",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "title", Number: 2, TypeName: "string"},
						},
					},
				},
			},
			// GORM DAL proto
			{
				Name: "library/v1/dal/book_gorm.proto",
				Pkg:  "library.v1.dal",
				Messages: []testutil.TestMessage{
					{
						Name: "BookGorm",
						GormOpts: &dalv1.GormOptions{
							Source: "library.v1.Book",
							Table:  "books",
						},
						Fields: []testutil.TestField{
							{
								Name:     "id",
								Number:   1,
								TypeName: "string",
								ColumnOpts: &dalv1.ColumnOptions{
									GormTags: []string{"primaryKey"},
								},
							},
							{Name: "title", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	// Collect GORM messages
	messages, err := collector.CollectMessages(plugin, collector.TargetGorm)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("Expected 1 GORM message, got %d", len(messages))
	}

	// When: Generate GORM code
	result, err := Generate(messages)

	// Then: Should succeed
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Then: Should generate at least one file
	if len(result.Files) == 0 {
		t.Fatal("Expected at least one generated file")
	}

	generatedCode := result.Files[0].Content

	// Then: Should generate GORM struct with correct name
	if !strings.Contains(generatedCode, "type BookGORM struct") {
		t.Error("Expected 'type BookGORM struct' in generated code")
	}

	// Then: Should generate fields with GORM tags
	if !strings.Contains(generatedCode, "ID") && !strings.Contains(generatedCode, "Id") {
		t.Error("Expected ID field in generated struct")
	}
	if !strings.Contains(generatedCode, "Title") {
		t.Error("Expected Title field in generated struct")
	}

	// Then: Should generate primary key tag
	if !strings.Contains(generatedCode, "primaryKey") {
		t.Error("Expected primaryKey tag for ID field")
	}

	// Then: Should generate TableName method with pointer receiver
	if !strings.Contains(generatedCode, "func (*BookGORM) TableName() string") {
		t.Error("Expected TableName() method with pointer receiver in generated code")
	}

	if !strings.Contains(generatedCode, `return "books"`) {
		t.Error("Expected TableName() to return 'books'")
	}
}

// TestGenerateConverters tests that converter functions are generated
// for converting between API messages and GORM structs.
//
// This test verifies:
// - ToGORM converter function with decorator parameter
// - FromGORM converter function with decorator parameter
// - Pointer receivers to avoid struct copying
func TestGenerateConverters(t *testing.T) {
	// Given: A Book message with GORM mapping
	plugin := testutil.CreateTestPlugin(t, &testutil.TestProtoSet{
		Files: []testutil.TestFile{
			// API proto
			{
				Name: "library/v1/book.proto",
				Pkg:  "library.v1",
				Messages: []testutil.TestMessage{
					{
						Name: "Book",
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "title", Number: 2, TypeName: "string"},
						},
					},
				},
			},
			// GORM DAL proto
			{
				Name: "library/v1/dal/book_gorm.proto",
				Pkg:  "library.v1.dal",
				Messages: []testutil.TestMessage{
					{
						Name: "BookGorm",
						GormOpts: &dalv1.GormOptions{
							Source: "library.v1.Book",
							Table:  "books",
						},
						Fields: []testutil.TestField{
							{Name: "id", Number: 1, TypeName: "string"},
							{Name: "title", Number: 2, TypeName: "string"},
						},
					},
				},
			},
		},
	})

	messages, err := collector.CollectMessages(plugin, collector.TargetGorm)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}

	// When: Generate converters
	result, err := GenerateConverters(messages)

	// Then: Should succeed
	if err != nil {
		t.Fatalf("GenerateConverters failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("Expected at least one converter file")
	}

	converterCode := result.Files[0].Content

	// Then: Should generate ToGORM function with full name (SourceToGormType)
	if !strings.Contains(converterCode, "func BookToBookGORM(") {
		t.Error("Expected BookToBookGORM function")
	}

	// Should accept API message pointer as src
	if !strings.Contains(converterCode, "src *") {
		t.Error("Expected src parameter in ToGORM")
	}

	// Should accept dest parameter for in-place conversion
	if !strings.Contains(converterCode, "dest *BookGORM") {
		t.Error("Expected dest *BookGORM parameter in ToGORM")
	}

	// Should accept decorator function
	if !strings.Contains(converterCode, "decorator func(*") {
		t.Error("Expected decorator parameter in ToGORM")
	}

	// Should return named GORM struct pointer and error
	if !strings.Contains(converterCode, "(out *BookGORM, err error)") {
		t.Error("Expected (out *BookGORM, err error) return in ToGORM")
	}

	// Then: Should generate FromGORM function with full name (SourceFromTargetType)
	if !strings.Contains(converterCode, "func BookFromBookGORM(") {
		t.Error("Expected BookFromBookGORM function")
	}

	// Should accept dest parameter (for proto message)
	if !strings.Contains(converterCode, "dest *") {
		t.Error("Expected dest parameter in FromGORM")
	}

	// Should accept GORM struct pointer as src
	if !strings.Contains(converterCode, "src *BookGORM") {
		t.Error("Expected src *BookGORM parameter in FromGORM")
	}

	// Should return named API message pointer and error (with package qualifier)
	if !strings.Contains(converterCode, ".Book") || !strings.Contains(converterCode, "(out *") {
		t.Error("Expected qualified Book type and named return in FromGORM")
	}
}

// Test helpers have been moved to pkg/generator/testutil
