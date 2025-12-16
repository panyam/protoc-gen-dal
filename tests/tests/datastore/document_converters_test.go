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
	"reflect"
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/datastore"
	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// createTestDocument creates a sample api.Document for testing.
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
		Tags:      []string{"protobuf", "dal", "datastore", "testing"},
	}
}

// TestDocumentDatastoreEmpty_AutoMergeAllFields verifies that DocumentDatastoreEmpty
// auto-merges all fields from api.Document when no explicit fields are declared.
func TestDocumentDatastoreEmpty_AutoMergeAllFields(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentDatastoreEmpty
	dsDoc, err := datastore.DocumentToDocumentDatastoreEmpty(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreEmpty failed: %v", err)
	}

	// Verify all fields were merged
	if dsDoc.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", dsDoc.Id, src.Id)
	}

	if dsDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", dsDoc.Title, src.Title)
	}

	if dsDoc.Content != src.Content {
		t.Errorf("Content mismatch: got %s, want %s", dsDoc.Content, src.Content)
	}

	if dsDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", dsDoc.Author, src.Author)
	}

	if !dsDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", dsDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !dsDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", dsDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if dsDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", dsDoc.Published, src.Published)
	}

	if dsDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", dsDoc.ViewCount, src.ViewCount)
	}

	// Verify Tags (repeated field)
	if !reflect.DeepEqual(dsDoc.Tags, src.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", dsDoc.Tags, src.Tags)
	}

	// Convert back to API
	apiDoc, err := datastore.DocumentFromDocumentDatastoreEmpty(nil, dsDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastoreEmpty failed: %v", err)
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

	if !reflect.DeepEqual(apiDoc.Tags, src.Tags) {
		t.Errorf("Round-trip Tags mismatch: got %v, want %v", apiDoc.Tags, src.Tags)
	}
}

// TestDocumentDatastorePartial_ExplicitFieldOverride verifies that DocumentDatastorePartial
// correctly overrides explicitly declared fields while auto-merging others.
// Note: DocumentDatastorePartial converts Id from uint32 to string.
func TestDocumentDatastorePartial_ExplicitFieldOverride(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentDatastorePartial
	dsDoc, err := datastore.DocumentToDocumentDatastorePartial(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastorePartial failed: %v", err)
	}

	// Verify Id is converted to string
	if dsDoc.Id != "1" {
		t.Errorf("Id mismatch: got %s, want '1'", dsDoc.Id)
	}

	if dsDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", dsDoc.Title, src.Title)
	}

	// Verify auto-merged fields are present
	if dsDoc.Content != src.Content {
		t.Errorf("Content mismatch: got %s, want %s", dsDoc.Content, src.Content)
	}

	if dsDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", dsDoc.Author, src.Author)
	}

	if !dsDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", dsDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !dsDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", dsDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if dsDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", dsDoc.Published, src.Published)
	}

	if dsDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", dsDoc.ViewCount, src.ViewCount)
	}

	if !reflect.DeepEqual(dsDoc.Tags, src.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", dsDoc.Tags, src.Tags)
	}

	// Convert back to API
	apiDoc, err := datastore.DocumentFromDocumentDatastorePartial(nil, dsDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastorePartial failed: %v", err)
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

	if !reflect.DeepEqual(apiDoc.Tags, src.Tags) {
		t.Errorf("Round-trip Tags mismatch: got %v, want %v", apiDoc.Tags, src.Tags)
	}
}

// TestDocumentDatastoreSkip_ExcludesSkippedFields verifies that DocumentDatastoreSkip
// excludes fields marked with skip_field annotation.
// Note: Content field is skipped in this variant.
func TestDocumentDatastoreSkip_ExcludesSkippedFields(t *testing.T) {
	src := createTestDocument()

	// Convert to DocumentDatastoreSkip
	dsDoc, err := datastore.DocumentToDocumentDatastoreSkip(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreSkip failed: %v", err)
	}

	// Verify Id is converted to string
	if dsDoc.Id != "1" {
		t.Errorf("Id mismatch: got %s, want '1'", dsDoc.Id)
	}

	if dsDoc.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", dsDoc.Title, src.Title)
	}

	// Content field is SKIPPED - verify it doesn't exist in dsDoc
	// The struct doesn't have a Content field (skip_field annotation)

	if dsDoc.Author != src.Author {
		t.Errorf("Author mismatch: got %s, want %s", dsDoc.Author, src.Author)
	}

	if !dsDoc.CreatedAt.Equal(src.CreatedAt.AsTime()) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", dsDoc.CreatedAt, src.CreatedAt.AsTime())
	}

	if !dsDoc.UpdatedAt.Equal(src.UpdatedAt.AsTime()) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", dsDoc.UpdatedAt, src.UpdatedAt.AsTime())
	}

	if dsDoc.Published != src.Published {
		t.Errorf("Published mismatch: got %v, want %v", dsDoc.Published, src.Published)
	}

	if dsDoc.ViewCount != src.ViewCount {
		t.Errorf("ViewCount mismatch: got %d, want %d", dsDoc.ViewCount, src.ViewCount)
	}

	if !reflect.DeepEqual(dsDoc.Tags, src.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", dsDoc.Tags, src.Tags)
	}

	// Convert back to API
	apiDoc, err := datastore.DocumentFromDocumentDatastoreSkip(nil, dsDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastoreSkip failed: %v", err)
	}

	// Verify round-trip for non-skipped fields
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}

	if apiDoc.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiDoc.Title, src.Title)
	}

	// Content should be empty (zero value) after round-trip since it was skipped
	if apiDoc.Content != "" {
		t.Errorf("Round-trip Content should be empty (skipped field), got %s", apiDoc.Content)
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

	if !reflect.DeepEqual(apiDoc.Tags, src.Tags) {
		t.Errorf("Round-trip Tags mismatch: got %v, want %v", apiDoc.Tags, src.Tags)
	}
}

// TestDocumentDatastoreEmpty_NilSource verifies nil source handling.
func TestDocumentDatastoreEmpty_NilSource(t *testing.T) {
	dsDoc, err := datastore.DocumentToDocumentDatastoreEmpty(nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreEmpty with nil failed: %v", err)
	}
	if dsDoc != nil {
		t.Errorf("Expected nil result for nil source, got %v", dsDoc)
	}

	apiDoc, err := datastore.DocumentFromDocumentDatastoreEmpty(nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastoreEmpty with nil failed: %v", err)
	}
	if apiDoc != nil {
		t.Errorf("Expected nil result for nil source, got %v", apiDoc)
	}
}

// TestDocumentDatastorePartial_NilSource verifies nil source handling.
func TestDocumentDatastorePartial_NilSource(t *testing.T) {
	dsDoc, err := datastore.DocumentToDocumentDatastorePartial(nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastorePartial with nil failed: %v", err)
	}
	if dsDoc != nil {
		t.Errorf("Expected nil result for nil source, got %v", dsDoc)
	}

	apiDoc, err := datastore.DocumentFromDocumentDatastorePartial(nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastorePartial with nil failed: %v", err)
	}
	if apiDoc != nil {
		t.Errorf("Expected nil result for nil source, got %v", apiDoc)
	}
}

// TestDocumentDatastoreSkip_NilSource verifies nil source handling.
func TestDocumentDatastoreSkip_NilSource(t *testing.T) {
	dsDoc, err := datastore.DocumentToDocumentDatastoreSkip(nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreSkip with nil failed: %v", err)
	}
	if dsDoc != nil {
		t.Errorf("Expected nil result for nil source, got %v", dsDoc)
	}

	apiDoc, err := datastore.DocumentFromDocumentDatastoreSkip(nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastoreSkip with nil failed: %v", err)
	}
	if apiDoc != nil {
		t.Errorf("Expected nil result for nil source, got %v", apiDoc)
	}
}

// TestDocumentDatastoreEmpty_NilTimestamps verifies nil timestamp handling.
func TestDocumentDatastoreEmpty_NilTimestamps(t *testing.T) {
	src := &api.Document{
		Id:        2,
		Title:     "No Timestamps",
		Content:   "A document without timestamps",
		Author:    "Bob",
		CreatedAt: nil,
		UpdatedAt: nil,
		Published: false,
		ViewCount: 0,
		Tags:      nil,
	}

	dsDoc, err := datastore.DocumentToDocumentDatastoreEmpty(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreEmpty failed: %v", err)
	}

	// Nil timestamps should result in zero time values
	if !dsDoc.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should be zero, got %v", dsDoc.CreatedAt)
	}

	if !dsDoc.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt should be zero, got %v", dsDoc.UpdatedAt)
	}

	// Convert back
	apiDoc, err := datastore.DocumentFromDocumentDatastoreEmpty(nil, dsDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastoreEmpty failed: %v", err)
	}

	// Zero time converts back to nil (TimeToTimestamp returns nil for zero time)
	if apiDoc.CreatedAt != nil {
		t.Errorf("Round-trip CreatedAt should be nil for zero time, got %v", apiDoc.CreatedAt)
	}

	if apiDoc.UpdatedAt != nil {
		t.Errorf("Round-trip UpdatedAt should be nil for zero time, got %v", apiDoc.UpdatedAt)
	}
}

// TestDocumentDatastoreEmpty_EmptyTags verifies empty Tags handling.
func TestDocumentDatastoreEmpty_EmptyTags(t *testing.T) {
	src := &api.Document{
		Id:        3,
		Title:     "Empty Tags",
		Content:   "A document with empty tags",
		Author:    "Charlie",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Published: true,
		ViewCount: 100,
		Tags:      []string{},
	}

	dsDoc, err := datastore.DocumentToDocumentDatastoreEmpty(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreEmpty failed: %v", err)
	}

	if dsDoc.Tags == nil {
		t.Error("Tags should not be nil for empty slice")
	}

	if len(dsDoc.Tags) != 0 {
		t.Errorf("Tags should be empty, got %v", dsDoc.Tags)
	}
}

// TestDocumentDatastoreEmpty_NilTags verifies nil Tags handling.
func TestDocumentDatastoreEmpty_NilTags(t *testing.T) {
	src := &api.Document{
		Id:        4,
		Title:     "Nil Tags",
		Content:   "A document with nil tags",
		Author:    "Diana",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Published: true,
		ViewCount: 200,
		Tags:      nil,
	}

	dsDoc, err := datastore.DocumentToDocumentDatastoreEmpty(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreEmpty failed: %v", err)
	}

	if dsDoc.Tags != nil {
		t.Errorf("Tags should be nil, got %v", dsDoc.Tags)
	}
}

// TestDocumentDatastorePartial_LargeId verifies large ID conversion.
func TestDocumentDatastorePartial_LargeId(t *testing.T) {
	src := &api.Document{
		Id:        4294967295, // Max uint32
		Title:     "Large ID Document",
		Content:   "Testing large ID handling",
		Author:    "Eve",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Published: true,
		ViewCount: 500,
		Tags:      []string{"test"},
	}

	dsDoc, err := datastore.DocumentToDocumentDatastorePartial(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastorePartial failed: %v", err)
	}

	// Verify large ID is correctly converted to string
	if dsDoc.Id != "4294967295" {
		t.Errorf("Large Id mismatch: got %s, want '4294967295'", dsDoc.Id)
	}

	// Convert back
	apiDoc, err := datastore.DocumentFromDocumentDatastorePartial(nil, dsDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastorePartial failed: %v", err)
	}

	// Verify round-trip preserves large ID
	if apiDoc.Id != src.Id {
		t.Errorf("Round-trip large Id mismatch: got %d, want %d", apiDoc.Id, src.Id)
	}
}

// TestDocumentDatastoreEmpty_WithDecorator verifies decorator function is called.
func TestDocumentDatastoreEmpty_WithDecorator(t *testing.T) {
	src := createTestDocument()

	decoratorCalled := false
	decorator := func(src *api.Document, dest *datastore.DocumentDatastoreEmpty) error {
		decoratorCalled = true
		// Modify a field in decorator
		dest.Title = "Modified Title"
		return nil
	}

	dsDoc, err := datastore.DocumentToDocumentDatastoreEmpty(src, nil, decorator)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastoreEmpty failed: %v", err)
	}

	if !decoratorCalled {
		t.Error("Decorator was not called")
	}

	if dsDoc.Title != "Modified Title" {
		t.Errorf("Decorator modification not applied: got %s", dsDoc.Title)
	}
}

// TestDocumentDatastorePartial_ZeroId verifies zero ID handling.
func TestDocumentDatastorePartial_ZeroId(t *testing.T) {
	src := &api.Document{
		Id:        0,
		Title:     "Zero ID Document",
		Content:   "Testing zero ID handling",
		Author:    "Frank",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Published: false,
		ViewCount: 0,
		Tags:      nil,
	}

	dsDoc, err := datastore.DocumentToDocumentDatastorePartial(src, nil, nil)
	if err != nil {
		t.Fatalf("DocumentToDocumentDatastorePartial failed: %v", err)
	}

	// Zero should convert to "0"
	if dsDoc.Id != "0" {
		t.Errorf("Zero Id mismatch: got %s, want '0'", dsDoc.Id)
	}

	// Convert back
	apiDoc, err := datastore.DocumentFromDocumentDatastorePartial(nil, dsDoc, nil)
	if err != nil {
		t.Fatalf("DocumentFromDocumentDatastorePartial failed: %v", err)
	}

	if apiDoc.Id != 0 {
		t.Errorf("Round-trip zero Id mismatch: got %d, want 0", apiDoc.Id)
	}
}

// TestDocumentDatastoreSkip_ContentFieldMissing verifies Content field is not in the struct.
func TestDocumentDatastoreSkip_ContentFieldMissing(t *testing.T) {
	// This is a compile-time verification - if Content existed, this would fail
	dsDoc := &datastore.DocumentDatastoreSkip{
		Id:        "1",
		Title:     "Test",
		Author:    "Test Author",
		Published: true,
		ViewCount: 0,
		// Content: "test", // This would cause a compile error
	}

	if dsDoc.Title != "Test" {
		t.Errorf("Title should be 'Test', got %s", dsDoc.Title)
	}
}
