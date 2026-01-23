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
	"context"
	"testing"

	"cloud.google.com/go/datastore"
	dsgen "github.com/panyam/protoc-gen-dal/tests/gen/datastore"
)

const testRecord3Kind = "test_records3"

func init() {
	// Add test_records3 to testKinds for cleanup
	testKinds = append(testKinds, testRecord3Kind)
}

// TestMapStringInt64_Put tests that a struct with map[string]int64 can be stored
// in Google Cloud Datastore.
//
// FAILING TEST: Google Cloud Datastore does NOT natively support Go map types.
// The generated struct must implement the PropertyLoadSaver interface to
// serialize maps as JSON or flattened properties.
//
// This test documents the issue and will pass once the fix is implemented.
func TestMapStringInt64_Put(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()
	namespace := getTestNamespace()

	// Clean up before and after
	if namespace != "" {
		cleanupKindInNamespace(ctx, client, testRecord3Kind, namespace)
		t.Cleanup(func() {
			cleanupKindInNamespace(ctx, client, testRecord3Kind, namespace)
		})
	}

	// Create entity with map[string]int64 field
	record := &dsgen.TestRecord3Datastore{
		Id:         "test-1",
		EntityType: "post",
		EntityId:   "post-123",
		TotalCount: 5,
		CountsByType: map[string]int64{
			"like":  3,
			"love":  2,
		},
	}

	// Build key
	key := datastore.NameKey(testRecord3Kind, record.Id, nil)
	if namespace != "" {
		key.Namespace = namespace
	}

	// Try to put - this will fail without PropertyLoadSaver implementation
	_, err := client.Put(ctx, key, record)
	if err != nil {
		// Expected error: "datastore: unsupported struct field type: map[string]int64"
		t.Fatalf("Put failed (map[string]int64 not supported without PropertyLoadSaver): %v", err)
	}

	// Get back and verify
	var retrieved dsgen.TestRecord3Datastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify fields
	if retrieved.Id != record.Id {
		t.Errorf("Id mismatch: got %s, want %s", retrieved.Id, record.Id)
	}
	if retrieved.EntityType != record.EntityType {
		t.Errorf("EntityType mismatch: got %s, want %s", retrieved.EntityType, record.EntityType)
	}
	if retrieved.TotalCount != record.TotalCount {
		t.Errorf("TotalCount mismatch: got %d, want %d", retrieved.TotalCount, record.TotalCount)
	}

	// Verify map contents
	if len(retrieved.CountsByType) != len(record.CountsByType) {
		t.Errorf("CountsByType length mismatch: got %d, want %d", len(retrieved.CountsByType), len(record.CountsByType))
	}
	for k, v := range record.CountsByType {
		if got, ok := retrieved.CountsByType[k]; !ok {
			t.Errorf("CountsByType missing key %q", k)
		} else if got != v {
			t.Errorf("CountsByType[%q] mismatch: got %d, want %d", k, got, v)
		}
	}

	t.Log("SUCCESS: map[string]int64 field was stored and retrieved correctly")
}

// TestMapStringInt64_EmptyMap tests handling of empty maps.
func TestMapStringInt64_EmptyMap(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()
	namespace := getTestNamespace()

	// Create entity with empty map
	record := &dsgen.TestRecord3Datastore{
		Id:           "test-empty",
		EntityType:   "post",
		EntityId:     "post-456",
		TotalCount:   0,
		CountsByType: map[string]int64{},
	}

	key := datastore.NameKey(testRecord3Kind, record.Id, nil)
	if namespace != "" {
		key.Namespace = namespace
	}

	_, err := client.Put(ctx, key, record)
	if err != nil {
		t.Fatalf("Put failed for empty map: %v", err)
	}

	var retrieved dsgen.TestRecord3Datastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Empty map should be retrieved as empty (or nil)
	if retrieved.CountsByType != nil && len(retrieved.CountsByType) != 0 {
		t.Errorf("Expected empty/nil CountsByType, got %v", retrieved.CountsByType)
	}
}

// TestMapStringInt64_NilMap tests handling of nil maps.
func TestMapStringInt64_NilMap(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()
	namespace := getTestNamespace()

	// Create entity with nil map
	record := &dsgen.TestRecord3Datastore{
		Id:           "test-nil",
		EntityType:   "post",
		EntityId:     "post-789",
		TotalCount:   0,
		CountsByType: nil,
	}

	key := datastore.NameKey(testRecord3Kind, record.Id, nil)
	if namespace != "" {
		key.Namespace = namespace
	}

	_, err := client.Put(ctx, key, record)
	if err != nil {
		t.Fatalf("Put failed for nil map: %v", err)
	}

	var retrieved dsgen.TestRecord3Datastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Nil map should be retrieved as nil or empty
	if retrieved.CountsByType != nil && len(retrieved.CountsByType) != 0 {
		t.Errorf("Expected nil/empty CountsByType, got %v", retrieved.CountsByType)
	}
}
