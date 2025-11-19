package gorm

import (
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// createTestDocument creates a sample api.Document for testing
func createTestDocument() *api.Document {
	createdAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

	return &api.Document{
		Id:        1,
		Title:     "Understanding Protobuf Field Merging",
		Content:   "This document explains how field merging works in DAL generation...",
		Author:    "Alice Smith",
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
		Published: true,
		ViewCount: 1234,
		Tags:      []string{"protobuf", "dal", "gorm", "testing"},
	}
}

// TestDocumentGormEmpty_AutoMergeAllFields verifies that DocumentGormEmpty
// auto-merges all fields from api.Document when no explicit fields are declared.
func TestDocumentGormEmpty_AutoMergeAllFields(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentGormEmpty
	gormDoc, err := gorm.DocumentToDocumentGormEmpty(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentGormEmpty failed: %v", err)
	}

	// Verify all fields were merged
	if gormDoc.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormDoc.Id, src.Id)
	}

	if gormDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", gormDoc.Title, src.Title)
	}

	if gormDoc.Content != src.Content {
		t.Errorf("Content mismatch: got %s, want %s", gormDoc.Content, src.Content)
	}

	if gormDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", gormDoc.Author, src.Author)
	}

	if !gormDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", gormDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !gormDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", gormDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if gormDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", gormDoc.Published, src.Published)
	}

	if gormDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", gormDoc.ViewCount, src.ViewCount)
	}

	// Verify Tags (repeated field)
	if len(gormDoc.Tags) != len(src.Tags) {
		t.Fatalf("Tags length mismatch: got %d, want %d", len(gormDoc.Tags), len(src.Tags))
	}

	for i, tag := range src.Tags {
		if gormDoc.Tags[i] != tag {
			t.Errorf("Tags[%d] mismatch: got %s, want %s", i, gormDoc.Tags[i], tag)
		}
	}

	// Convert back to API
	apiDoc, err := gorm.DocumentFromDocumentGormEmpty(nil, gormDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentGormEmpty failed: %v", err)
	}

	// Verify round-trip
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}

	if apiDoc.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiDoc.Title, src.Title)
	}

	if apiDoc.Content != src.Content {
		t.Errorf("Round-trip Content mismatch: got %s, want %s", apiDoc.Content, src.Content)
	}

	if apiDoc.Author != src.Author {
		t.Errorf("Round-trip Author mismatch: got %s, want %s", apiDoc.Author, src.Author)
	}

	if !apiDoc.CreatedAt.AsTime().Equal(src.CreatedAt.AsTime()) {
		t.Errorf("Round-trip CreatedAt mismatch: got %v, want %v", apiDoc.CreatedAt.AsTime(), src.CreatedAt.AsTime())
	}

	if !apiDoc.UpdatedAt.AsTime().Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("Round-trip UpdatedAt mismatch: got %v, want %v", apiDoc.UpdatedAt.AsTime(), src.UpdatedAt.AsTime())
	}

	if apiDoc.Published != src.Published {
		t.Errorf("Round-trip Published mismatch: got %v, want %v", apiDoc.Published, src.Published)
	}

	if apiDoc.ViewCount != src.ViewCount {
		t.Errorf("Round-trip ViewCount mismatch: got %d, want %d", apiDoc.ViewCount, src.ViewCount)
	}

	// Verify Tags round-trip
	if len(apiDoc.Tags) != len(src.Tags) {
		t.Fatalf("Round-trip Tags length mismatch: got %d, want %d", len(apiDoc.Tags), len(src.Tags))
	}

	for i, tag := range src.Tags {
		if apiDoc.Tags[i] != tag {
			t.Errorf("Round-trip Tags[%d] mismatch: got %s, want %s", i, apiDoc.Tags[i], tag)
		}
	}
}

// TestDocumentGormPartial_ExplicitFieldOverride verifies that DocumentGormPartial
// correctly overrides explicitly declared fields (id, title) while auto-merging others.
func TestDocumentGormPartial_ExplicitFieldOverride(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentGormPartial
	gormDoc, err := gorm.DocumentToDocumentGormPartial(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentGormPartial failed: %v", err)
	}

	// Verify explicitly overridden fields (id, title) are present
	if gormDoc.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormDoc.Id, src.Id)
	}

	if gormDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", gormDoc.Title, src.Title)
	}

	// Verify auto-merged fields are present
	if gormDoc.Content != src.Content {
		t.Errorf("Content mismatch: got %s, want %s", gormDoc.Content, src.Content)
	}

	if gormDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", gormDoc.Author, src.Author)
	}

	if !gormDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", gormDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !gormDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", gormDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if gormDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", gormDoc.Published, src.Published)
	}

	if gormDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", gormDoc.ViewCount, src.ViewCount)
	}

	// Verify Tags (repeated field)
	if len(gormDoc.Tags) != len(src.Tags) {
		t.Fatalf("Tags length mismatch: got %d, want %d", len(gormDoc.Tags), len(src.Tags))
	}

	for i, tag := range src.Tags {
		if gormDoc.Tags[i] != tag {
			t.Errorf("Tags[%d] mismatch: got %s, want %s", i, gormDoc.Tags[i], tag)
		}
	}

	// Convert back to API
	apiDoc, err := gorm.DocumentFromDocumentGormPartial(nil, gormDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentGormPartial failed: %v", err)
	}

	// Verify round-trip for all fields
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}

	if apiDoc.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiDoc.Title, src.Title)
	}

	if apiDoc.Content != src.Content {
		t.Errorf("Round-trip Content mismatch: got %s, want %s", apiDoc.Content, src.Content)
	}

	if apiDoc.Published != src.Published {
		t.Errorf("Round-trip Published mismatch: got %v, want %v", apiDoc.Published, src.Published)
	}
}

// TestDocumentGormSkip_SkipContentField verifies that DocumentGormSkip
// excludes the 'content' field (marked with skip_field=true) while merging others.
func TestDocumentGormSkip_SkipContentField(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentGormSkip
	gormDoc, err := gorm.DocumentToDocumentGormSkip(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentGormSkip failed: %v", err)
	}

	// Verify id field is present (explicitly declared)
	if gormDoc.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormDoc.Id, src.Id)
	}

	// Verify auto-merged fields are present
	if gormDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", gormDoc.Title, src.Title)
	}

	if gormDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", gormDoc.Author, src.Author)
	}

	if !gormDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", gormDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !gormDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", gormDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if gormDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", gormDoc.Published, src.Published)
	}

	if gormDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", gormDoc.ViewCount, src.ViewCount)
	}

	// Verify Tags (repeated field)
	if len(gormDoc.Tags) != len(src.Tags) {
		t.Fatalf("Tags length mismatch: got %d, want %d", len(gormDoc.Tags), len(src.Tags))
	}

	// NOTE: The DocumentGormSkip struct does NOT have a Content field
	// because it was marked with skip_field=true. This is the key test:
	// we verify that the converter doesn't try to convert it, and the struct doesn't have it.

	// Convert back to API
	apiDoc, err := gorm.DocumentFromDocumentGormSkip(nil, gormDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentGormSkip failed: %v", err)
	}

	// Verify round-trip for non-skipped fields
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}

	if apiDoc.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiDoc.Title, src.Title)
	}

	// Content field should be empty (zero value) on round-trip since it was skipped
	if apiDoc.Content != "" {
		t.Errorf("Round-trip Content should be empty (skipped), got %s", apiDoc.Content)
	}

	if apiDoc.Published != src.Published {
		t.Errorf("Round-trip Published mismatch: got %v, want %v", apiDoc.Published, src.Published)
	}
}

// TestDocumentGormExtra_AdditionalDBFields verifies that DocumentGormExtra
// auto-merges all source fields AND adds extra DB-only fields (deleted_at, version).
func TestDocumentGormExtra_AdditionalDBFields(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentGormExtra
	gormDoc, err := gorm.DocumentToDocumentGormExtra(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentGormExtra failed: %v", err)
	}

	// Verify all auto-merged fields from source
	if gormDoc.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormDoc.Id, src.Id)
	}

	if gormDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", gormDoc.Title, src.Title)
	}

	if gormDoc.Content != src.Content {
		t.Errorf("Content mismatch: got %s, want %s", gormDoc.Content, src.Content)
	}

	if gormDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", gormDoc.Author, src.Author)
	}

	if !gormDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", gormDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !gormDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", gormDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if gormDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", gormDoc.Published, src.Published)
	}

	if gormDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", gormDoc.ViewCount, src.ViewCount)
	}

	// Verify Tags (repeated field)
	if len(gormDoc.Tags) != len(src.Tags) {
		t.Fatalf("Tags length mismatch: got %d, want %d", len(gormDoc.Tags), len(src.Tags))
	}

	for i, tag := range src.Tags {
		if gormDoc.Tags[i] != tag {
			t.Errorf("Tags[%d] mismatch: got %s, want %s", i, gormDoc.Tags[i], tag)
		}
	}

	// Verify extra DB-only fields exist with zero values
	// (they're not converted from source since they don't exist in api.Document)
	if !gormDoc.DeletedAt.IsZero() {
		t.Logf("DeletedAt has non-zero value: %v (expected zero, but OK for DB field)", gormDoc.DeletedAt)
	}

	if gormDoc.Version != 0 {
		t.Logf("Version has non-zero value: %d (expected zero, but OK for DB field)", gormDoc.Version)
	}

	// Convert back to API
	apiDoc, err := gorm.DocumentFromDocumentGormExtra(nil, gormDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentGormExtra failed: %v", err)
	}

	// Verify round-trip for all source fields
	// Note: extra DB fields (deleted_at, version) are NOT in api.Document, so they don't round-trip
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}

	if apiDoc.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiDoc.Title, src.Title)
	}

	if apiDoc.Content != src.Content {
		t.Errorf("Round-trip Content mismatch: got %s, want %s", apiDoc.Content, src.Content)
	}

	if apiDoc.Published != src.Published {
		t.Errorf("Round-trip Published mismatch: got %v, want %v", apiDoc.Published, src.Published)
	}
}

// TestDocumentGormExtra_ExtraFieldsManipulation verifies that extra DB-only fields
// can be manipulated without affecting API conversion.
func TestDocumentGormExtra_ExtraFieldsManipulation(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentGormExtra
	gormDoc, err := gorm.DocumentToDocumentGormExtra(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentGormExtra failed: %v", err)
	}

	// Manually set extra DB fields (simulating DB operations)
	deletedAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	gormDoc.DeletedAt = deletedAt
	gormDoc.Version = 5

	// Verify extra fields were set
	if !gormDoc.DeletedAt.Equal(deletedAt) {
		t.Errorf("DeletedAt not set correctly: got %v, want %v", gormDoc.DeletedAt, deletedAt)
	}

	if gormDoc.Version != 5 {
		t.Errorf("Version not set correctly: got %d, want %d", gormDoc.Version, 5)
	}

	// Convert back to API - extra fields should not appear
	apiDoc, err := gorm.DocumentFromDocumentGormExtra(nil, gormDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentGormExtra failed: %v", err)
	}

	// Verify that API document only has source fields (no extra fields)
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}

	if apiDoc.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiDoc.Title, src.Title)
	}

	// The API message doesn't have deleted_at or version fields,
	// so we just verify that the conversion succeeded without errors
}

// TestDocumentConversions_NilTimestamps verifies that all document variants
// handle nil timestamps correctly.
func TestDocumentConversions_NilTimestamps(t *testing.T) {
	src := &api.Document{
		Id:        2,
		Title:     "Document with nil timestamps",
		Content:   "Content",
		Author:    "Bob Jones",
		CreatedAt: nil,
		UpdatedAt: nil,
		Published: false,
		ViewCount: 0,
		Tags:      nil,
	}

	testCases := []struct {
		name      string
		convertTo func(*api.Document, interface{}, interface{}) (interface{}, error)
		check     func(t *testing.T, gormDoc interface{})
	}{
		{
			name: "DocumentGormEmpty",
			convertTo: func(src *api.Document, dest interface{}, decorator interface{}) (interface{}, error) {
				return gorm.DocumentToDocumentGormEmpty(src, nil, nil)
			},
			check: func(t *testing.T, gormDoc interface{}) {
				doc := gormDoc.(*gorm.DocumentGormEmpty)
				if !doc.CreatedAt.IsZero() {
					t.Errorf("Expected zero CreatedAt, got %v", doc.CreatedAt)
				}
				if !doc.UpdatedAt.IsZero() {
					t.Errorf("Expected zero UpdatedAt, got %v", doc.UpdatedAt)
				}
			},
		},
		{
			name: "DocumentGormPartial",
			convertTo: func(src *api.Document, dest interface{}, decorator interface{}) (interface{}, error) {
				return gorm.DocumentToDocumentGormPartial(src, nil, nil)
			},
			check: func(t *testing.T, gormDoc interface{}) {
				doc := gormDoc.(*gorm.DocumentGormPartial)
				if !doc.CreatedAt.IsZero() {
					t.Errorf("Expected zero CreatedAt, got %v", doc.CreatedAt)
				}
				if !doc.UpdatedAt.IsZero() {
					t.Errorf("Expected zero UpdatedAt, got %v", doc.UpdatedAt)
				}
			},
		},
		{
			name: "DocumentGormSkip",
			convertTo: func(src *api.Document, dest interface{}, decorator interface{}) (interface{}, error) {
				return gorm.DocumentToDocumentGormSkip(src, nil, nil)
			},
			check: func(t *testing.T, gormDoc interface{}) {
				doc := gormDoc.(*gorm.DocumentGormSkip)
				if !doc.CreatedAt.IsZero() {
					t.Errorf("Expected zero CreatedAt, got %v", doc.CreatedAt)
				}
				if !doc.UpdatedAt.IsZero() {
					t.Errorf("Expected zero UpdatedAt, got %v", doc.UpdatedAt)
				}
			},
		},
		{
			name: "DocumentGormExtra",
			convertTo: func(src *api.Document, dest interface{}, decorator interface{}) (interface{}, error) {
				return gorm.DocumentToDocumentGormExtra(src, nil, nil)
			},
			check: func(t *testing.T, gormDoc interface{}) {
				doc := gormDoc.(*gorm.DocumentGormExtra)
				if !doc.CreatedAt.IsZero() {
					t.Errorf("Expected zero CreatedAt, got %v", doc.CreatedAt)
				}
				if !doc.UpdatedAt.IsZero() {
					t.Errorf("Expected zero UpdatedAt, got %v", doc.UpdatedAt)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gormDoc, err := tc.convertTo(src, nil, nil)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			tc.check(t, gormDoc)
		})
	}
}

// TestDocumentConversions_EmptyTags verifies that all document variants
// handle empty/nil tag slices correctly.
func TestDocumentConversions_EmptyTags(t *testing.T) {
	testCases := []struct {
		name string
		tags []string
	}{
		{"NilTags", nil},
		{"EmptyTags", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := &api.Document{
				Id:      3,
				Title:   "Document with " + tc.name,
				Content: "Content",
				Author:  "Carol White",
				Tags:    tc.tags,
			}

			// Test with DocumentGormEmpty
			gormDoc, err := gorm.DocumentToDocumentGormEmpty(src, nil, nil)
			if err != nil {
				t.Fatalf("DocumentToDocumentGormEmpty failed: %v", err)
			}

			if tc.tags == nil && gormDoc.Tags != nil {
				t.Errorf("Expected nil Tags, got %v", gormDoc.Tags)
			}

			if tc.tags != nil && len(gormDoc.Tags) != 0 {
				t.Errorf("Expected empty Tags, got %v", gormDoc.Tags)
			}

			// Round-trip
			apiDoc, err := gorm.DocumentFromDocumentGormEmpty(nil, gormDoc, nil)
			if err != nil {
				t.Fatalf("DocumentFromDocumentGormEmpty failed: %v", err)
			}

			if tc.tags == nil && apiDoc.Tags != nil {
				t.Errorf("Round-trip: Expected nil Tags, got %v", apiDoc.Tags)
			}

			if tc.tags != nil && len(apiDoc.Tags) != 0 {
				t.Errorf("Round-trip: Expected empty Tags, got %v", apiDoc.Tags)
			}
		})
	}
}
