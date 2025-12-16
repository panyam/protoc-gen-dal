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
	"github.com/panyam/protoc-gen-dal/tests/gen/datastore/dal"
)

const testUserKind = "TestUser"

// TestDALPut tests the Put method with various key scenarios.
func TestDALPut(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Clean up before and after test
	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	t.Run("PutWithStringID", func(t *testing.T) {
		user := &dsgen.UserDatastore{
			Id:    "user-1",
			Name:  "Alice",
			Email: "alice@example.com",
			Age:   30,
		}

		key, err := userDAL.Put(ctx, client, user)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		if key.Name != "user-1" {
			t.Errorf("Expected key name 'user-1', got '%s'", key.Name)
		}

		if user.Key == nil {
			t.Error("Expected user.Key to be set after Put")
		}
	})

	t.Run("PutWithExistingKey", func(t *testing.T) {
		existingKey := datastore.NameKey(testUserKind, "user-2", nil)
		user := &dsgen.UserDatastore{
			Key:   existingKey,
			Name:  "Bob",
			Email: "bob@example.com",
		}

		key, err := userDAL.Put(ctx, client, user)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		if key.Name != "user-2" {
			t.Errorf("Expected key name 'user-2', got '%s'", key.Name)
		}
	})

	t.Run("PutWithIncompleteKey", func(t *testing.T) {
		user := &dsgen.UserDatastore{
			Name:  "Charlie",
			Email: "charlie@example.com",
		}

		key, err := userDAL.Put(ctx, client, user)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		// Incomplete key should result in an auto-generated numeric ID
		if key.ID == 0 && key.Name == "" {
			t.Error("Expected key to have either ID or Name after Put")
		}
	})
}

// TestDALGet tests the Get method.
func TestDALGet(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create a test user first
	user := &dsgen.UserDatastore{
		Id:    "get-test-user",
		Name:  "GetTest",
		Email: "gettest@example.com",
		Age:   25,
	}
	key, err := userDAL.Put(ctx, client, user)
	if err != nil {
		t.Fatalf("Setup Put failed: %v", err)
	}

	t.Run("GetExisting", func(t *testing.T) {
		retrieved, err := userDAL.Get(ctx, client, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected entity, got nil")
		}

		if retrieved.Name != "GetTest" {
			t.Errorf("Expected Name 'GetTest', got '%s'", retrieved.Name)
		}

		if retrieved.Key == nil {
			t.Error("Expected Key to be set on retrieved entity")
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		nonExistentKey := datastore.NameKey(testUserKind, "does-not-exist", nil)
		retrieved, err := userDAL.Get(ctx, client, nonExistentKey)
		if err != nil {
			t.Fatalf("Get failed with error: %v", err)
		}

		if retrieved != nil {
			t.Errorf("Expected nil for non-existent entity, got %v", retrieved)
		}
	})
}

// TestDALGetByID tests the GetByID convenience method.
func TestDALGetByID(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create a test user
	user := &dsgen.UserDatastore{
		Id:    "getbyid-user",
		Name:  "GetByIDTest",
		Email: "getbyid@example.com",
	}
	_, err := userDAL.Put(ctx, client, user)
	if err != nil {
		t.Fatalf("Setup Put failed: %v", err)
	}

	t.Run("GetByIDExisting", func(t *testing.T) {
		retrieved, err := userDAL.GetByID(ctx, client, "getbyid-user")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected entity, got nil")
		}

		if retrieved.Name != "GetByIDTest" {
			t.Errorf("Expected Name 'GetByIDTest', got '%s'", retrieved.Name)
		}
	})

	t.Run("GetByIDNonExistent", func(t *testing.T) {
		retrieved, err := userDAL.GetByID(ctx, client, "does-not-exist")
		if err != nil {
			t.Fatalf("GetByID failed with error: %v", err)
		}

		if retrieved != nil {
			t.Errorf("Expected nil for non-existent entity, got %v", retrieved)
		}
	})
}

// TestDALDelete tests the Delete method.
func TestDALDelete(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create a test user
	user := &dsgen.UserDatastore{
		Id:    "delete-user",
		Name:  "DeleteTest",
		Email: "delete@example.com",
	}
	key, err := userDAL.Put(ctx, client, user)
	if err != nil {
		t.Fatalf("Setup Put failed: %v", err)
	}

	// Delete the user
	err = userDAL.Delete(ctx, client, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	retrieved, err := userDAL.Get(ctx, client, key)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil after deletion, entity still exists")
	}
}

// TestDALDeleteByID tests the DeleteByID convenience method.
func TestDALDeleteByID(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create a test user
	user := &dsgen.UserDatastore{
		Id:    "deletebyid-user",
		Name:  "DeleteByIDTest",
		Email: "deletebyid@example.com",
	}
	_, err := userDAL.Put(ctx, client, user)
	if err != nil {
		t.Fatalf("Setup Put failed: %v", err)
	}

	// Delete by ID
	err = userDAL.DeleteByID(ctx, client, "deletebyid-user")
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	// Verify deletion
	retrieved, err := userDAL.GetByID(ctx, client, "deletebyid-user")
	if err != nil {
		t.Fatalf("GetByID after delete failed: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil after deletion, entity still exists")
	}
}

// TestDALPutMulti tests the PutMulti method.
func TestDALPutMulti(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	users := []*dsgen.UserDatastore{
		{Id: "multi-1", Name: "User1", Email: "user1@example.com"},
		{Id: "multi-2", Name: "User2", Email: "user2@example.com"},
		{Id: "multi-3", Name: "User3", Email: "user3@example.com"},
	}

	keys, err := userDAL.PutMulti(ctx, client, users)
	if err != nil {
		t.Fatalf("PutMulti failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify all entities were created
	for i, user := range users {
		if user.Key == nil {
			t.Errorf("User %d Key not set after PutMulti", i)
		}
	}
}

// TestDALGetMulti tests the GetMulti method.
func TestDALGetMulti(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create test users
	users := []*dsgen.UserDatastore{
		{Id: "getmulti-1", Name: "User1", Email: "user1@example.com"},
		{Id: "getmulti-2", Name: "User2", Email: "user2@example.com"},
	}
	keys, err := userDAL.PutMulti(ctx, client, users)
	if err != nil {
		t.Fatalf("Setup PutMulti failed: %v", err)
	}

	// Add a non-existent key
	allKeys := append(keys, datastore.NameKey(testUserKind, "does-not-exist", nil))

	retrieved, err := userDAL.GetMulti(ctx, client, allKeys)
	if err != nil {
		t.Fatalf("GetMulti failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 results, got %d", len(retrieved))
	}

	// First two should exist
	if retrieved[0] == nil || retrieved[0].Name != "User1" {
		t.Errorf("Expected User1, got %v", retrieved[0])
	}
	if retrieved[1] == nil || retrieved[1].Name != "User2" {
		t.Errorf("Expected User2, got %v", retrieved[1])
	}

	// Third should be nil (non-existent)
	if retrieved[2] != nil {
		t.Errorf("Expected nil for non-existent key, got %v", retrieved[2])
	}
}

// TestDALGetMultiByIDs tests the GetMultiByIDs convenience method.
func TestDALGetMultiByIDs(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create test users
	users := []*dsgen.UserDatastore{
		{Id: "ids-1", Name: "User1", Email: "user1@example.com"},
		{Id: "ids-2", Name: "User2", Email: "user2@example.com"},
	}
	_, err := userDAL.PutMulti(ctx, client, users)
	if err != nil {
		t.Fatalf("Setup PutMulti failed: %v", err)
	}

	ids := []string{"ids-1", "ids-2", "does-not-exist"}
	retrieved, err := userDAL.GetMultiByIDs(ctx, client, ids)
	if err != nil {
		t.Fatalf("GetMultiByIDs failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 results, got %d", len(retrieved))
	}

	if retrieved[0] == nil || retrieved[0].Name != "User1" {
		t.Errorf("Expected User1, got %v", retrieved[0])
	}
	if retrieved[1] == nil || retrieved[1].Name != "User2" {
		t.Errorf("Expected User2, got %v", retrieved[1])
	}
	if retrieved[2] != nil {
		t.Errorf("Expected nil for non-existent ID, got %v", retrieved[2])
	}
}

// TestDALDeleteMulti tests the DeleteMulti method.
func TestDALDeleteMulti(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create test users
	users := []*dsgen.UserDatastore{
		{Id: "delmulti-1", Name: "User1", Email: "user1@example.com"},
		{Id: "delmulti-2", Name: "User2", Email: "user2@example.com"},
	}
	keys, err := userDAL.PutMulti(ctx, client, users)
	if err != nil {
		t.Fatalf("Setup PutMulti failed: %v", err)
	}

	// Delete all
	err = userDAL.DeleteMulti(ctx, client, keys)
	if err != nil {
		t.Fatalf("DeleteMulti failed: %v", err)
	}

	// Verify deletion
	retrieved, err := userDAL.GetMulti(ctx, client, keys)
	if err != nil {
		t.Fatalf("GetMulti after delete failed: %v", err)
	}

	for i, r := range retrieved {
		if r != nil {
			t.Errorf("Expected nil at index %d after deletion, got %v", i, r)
		}
	}
}

// TestDALQuery tests the Query method.
func TestDALQuery(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create test users with different ages
	users := []*dsgen.UserDatastore{
		{Id: "query-1", Name: "Young", Email: "young@example.com", Age: 20},
		{Id: "query-2", Name: "Middle", Email: "middle@example.com", Age: 30},
		{Id: "query-3", Name: "Older", Email: "older@example.com", Age: 40},
	}
	_, err := userDAL.PutMulti(ctx, client, users)
	if err != nil {
		t.Fatalf("Setup PutMulti failed: %v", err)
	}

	t.Run("QueryAll", func(t *testing.T) {
		q := datastore.NewQuery(testUserKind)
		results, err := userDAL.Query(ctx, client, q)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("QueryWithFilter", func(t *testing.T) {
		q := datastore.NewQuery(testUserKind).FilterField("Age", ">=", uint32(30))
		results, err := userDAL.Query(ctx, client, q)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results (age >= 30), got %d", len(results))
		}
	})
}

// TestDALCount tests the Count method.
func TestDALCount(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Create test users
	users := []*dsgen.UserDatastore{
		{Id: "count-1", Name: "User1", Email: "user1@example.com"},
		{Id: "count-2", Name: "User2", Email: "user2@example.com"},
		{Id: "count-3", Name: "User3", Email: "user3@example.com"},
	}
	_, err := userDAL.PutMulti(ctx, client, users)
	if err != nil {
		t.Fatalf("Setup PutMulti failed: %v", err)
	}

	q := datastore.NewQuery(testUserKind)
	count, err := userDAL.Count(ctx, client, q)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

// TestDALWillPutHook tests the WillPut hook functionality.
func TestDALWillPutHook(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)

	// Track hook calls
	hookCalled := false
	userDAL.WillPut = func(ctx context.Context, u *dsgen.UserDatastore) error {
		hookCalled = true
		// Modify the entity in the hook
		u.Name = "Modified by hook"
		return nil
	}

	user := &dsgen.UserDatastore{
		Id:    "hook-user",
		Name:  "Original",
		Email: "hook@example.com",
	}

	_, err := userDAL.Put(ctx, client, user)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if !hookCalled {
		t.Error("WillPut hook was not called")
	}

	// Verify the modification was persisted
	retrieved, err := userDAL.GetByID(ctx, client, "hook-user")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != "Modified by hook" {
		t.Errorf("Expected 'Modified by hook', got '%s'", retrieved.Name)
	}
}

// TestDALNamespaceOverride tests the namespace override functionality.
func TestDALNamespaceOverride(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	namespace := "test-namespace"

	// Clean up the namespace
	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := dal.NewUserDatastoreDAL(testUserKind)
	userDAL.Namespace = namespace

	user := &dsgen.UserDatastore{
		Id:    "ns-user",
		Name:  "NamespaceTest",
		Email: "ns@example.com",
	}

	key, err := userDAL.Put(ctx, client, user)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if key.Namespace != namespace {
		t.Errorf("Expected namespace '%s', got '%s'", namespace, key.Namespace)
	}
}
