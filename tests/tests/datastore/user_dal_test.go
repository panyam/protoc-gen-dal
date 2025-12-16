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
)

const testUserKind = "TestUser"

// TestUser is a simple test entity that only uses Datastore-compatible types.
// The generated UserDatastore has uint32 fields which Datastore doesn't support.
// Datastore only supports signed integers (int, int8, int16, int32, int64).
type TestUser struct {
	Key   *datastore.Key `datastore:"-"`
	Id    string         `datastore:"id"`
	Name  string         `datastore:"name"`
	Email string         `datastore:"email"`
}

// TestUserDAL provides DAL methods for TestUser - a minimal implementation
// to test the DAL pattern without depending on generated code with unsupported types.
type TestUserDAL struct {
	Kind      string
	Namespace string
	WillPut   func(context.Context, *TestUser) error
}

func NewTestUserDAL(kind string) *TestUserDAL {
	return &TestUserDAL{Kind: kind}
}

func (d *TestUserDAL) getKind() string {
	if d.Kind != "" {
		return d.Kind
	}
	return testUserKind
}

func (d *TestUserDAL) newKey(id string) *datastore.Key {
	key := datastore.NameKey(d.getKind(), id, nil)
	if d.Namespace != "" {
		key.Namespace = d.Namespace
	}
	return key
}

func (d *TestUserDAL) newIncompleteKey() *datastore.Key {
	key := datastore.IncompleteKey(d.getKind(), nil)
	if d.Namespace != "" {
		key.Namespace = d.Namespace
	}
	return key
}

func (d *TestUserDAL) Put(ctx context.Context, client *datastore.Client, obj *TestUser) (*datastore.Key, error) {
	if d.WillPut != nil {
		if err := d.WillPut(ctx, obj); err != nil {
			return nil, err
		}
	}

	var key *datastore.Key
	if obj.Key != nil {
		key = obj.Key
		if d.Namespace != "" {
			key.Namespace = d.Namespace
		}
	} else if obj.Id != "" {
		key = d.newKey(obj.Id)
	} else {
		key = d.newIncompleteKey()
	}

	resultKey, err := client.Put(ctx, key, obj)
	if err != nil {
		return nil, err
	}

	obj.Key = resultKey
	return resultKey, nil
}

func (d *TestUserDAL) Get(ctx context.Context, client *datastore.Client, key *datastore.Key) (*TestUser, error) {
	var entity TestUser
	err := client.Get(ctx, key, &entity)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, err
	}
	entity.Key = key
	return &entity, nil
}

func (d *TestUserDAL) GetByID(ctx context.Context, client *datastore.Client, id string) (*TestUser, error) {
	key := d.newKey(id)
	return d.Get(ctx, client, key)
}

func (d *TestUserDAL) Delete(ctx context.Context, client *datastore.Client, key *datastore.Key) error {
	return client.Delete(ctx, key)
}

func (d *TestUserDAL) DeleteByID(ctx context.Context, client *datastore.Client, id string) error {
	key := d.newKey(id)
	return d.Delete(ctx, client, key)
}

func (d *TestUserDAL) PutMulti(ctx context.Context, client *datastore.Client, objs []*TestUser) ([]*datastore.Key, error) {
	if len(objs) == 0 {
		return []*datastore.Key{}, nil
	}

	if d.WillPut != nil {
		for _, obj := range objs {
			if err := d.WillPut(ctx, obj); err != nil {
				return nil, err
			}
		}
	}

	keys := make([]*datastore.Key, len(objs))
	for i, obj := range objs {
		if obj.Key != nil {
			keys[i] = obj.Key
			if d.Namespace != "" {
				keys[i].Namespace = d.Namespace
			}
		} else if obj.Id != "" {
			keys[i] = d.newKey(obj.Id)
		} else {
			keys[i] = d.newIncompleteKey()
		}
	}

	resultKeys, err := client.PutMulti(ctx, keys, objs)
	if err != nil {
		return nil, err
	}

	for i, key := range resultKeys {
		objs[i].Key = key
	}

	return resultKeys, nil
}

func (d *TestUserDAL) GetMulti(ctx context.Context, client *datastore.Client, keys []*datastore.Key) ([]*TestUser, error) {
	if len(keys) == 0 {
		return []*TestUser{}, nil
	}

	entities := make([]TestUser, len(keys))
	err := client.GetMulti(ctx, keys, entities)
	if err != nil {
		if multiErr, ok := err.(datastore.MultiError); ok {
			result := make([]*TestUser, len(keys))
			for i, e := range multiErr {
				if e == nil {
					entities[i].Key = keys[i]
					result[i] = &entities[i]
				} else if e != datastore.ErrNoSuchEntity {
					return nil, err
				}
			}
			return result, nil
		}
		return nil, err
	}

	result := make([]*TestUser, len(keys))
	for i := range entities {
		entities[i].Key = keys[i]
		result[i] = &entities[i]
	}
	return result, nil
}

func (d *TestUserDAL) GetMultiByIDs(ctx context.Context, client *datastore.Client, ids []string) ([]*TestUser, error) {
	if len(ids) == 0 {
		return []*TestUser{}, nil
	}

	keys := make([]*datastore.Key, len(ids))
	for i, id := range ids {
		keys[i] = d.newKey(id)
	}

	return d.GetMulti(ctx, client, keys)
}

func (d *TestUserDAL) DeleteMulti(ctx context.Context, client *datastore.Client, keys []*datastore.Key) error {
	if len(keys) == 0 {
		return nil
	}
	return client.DeleteMulti(ctx, keys)
}

func (d *TestUserDAL) Query(ctx context.Context, client *datastore.Client, q *datastore.Query) ([]*TestUser, error) {
	var entities []*TestUser
	keys, err := client.GetAll(ctx, q, &entities)
	if err != nil {
		return nil, err
	}

	for i, key := range keys {
		entities[i].Key = key
	}

	return entities, nil
}

func (d *TestUserDAL) Count(ctx context.Context, client *datastore.Client, q *datastore.Query) (int, error) {
	return client.Count(ctx, q)
}

// TestDALPut tests the Put method with various key scenarios.
func TestDALPut(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Clean up before and after test
	cleanupKind(ctx, client, testUserKind)
	t.Cleanup(func() {
		cleanupKind(ctx, client, testUserKind)
	})

	userDAL := NewTestUserDAL(testUserKind)

	t.Run("PutWithStringID", func(t *testing.T) {
		user := &TestUser{
			Id:    "user-1",
			Name:  "Alice",
			Email: "alice@example.com",
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
		user := &TestUser{
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
		user := &TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create a test user first
	user := &TestUser{
		Id:    "get-test-user",
		Name:  "GetTest",
		Email: "gettest@example.com",
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create a test user
	user := &TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create a test user
	user := &TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create a test user
	user := &TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	users := []*TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create test users
	users := []*TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create test users
	users := []*TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create test users
	users := []*TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create test users with different names
	users := []*TestUser{
		{Id: "query-1", Name: "Alice", Email: "alice@example.com"},
		{Id: "query-2", Name: "Bob", Email: "bob@example.com"},
		{Id: "query-3", Name: "Charlie", Email: "charlie@example.com"},
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
		q := datastore.NewQuery(testUserKind).FilterField("name", ">=", "Bob")
		results, err := userDAL.Query(ctx, client, q)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results (name >= Bob), got %d", len(results))
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

	userDAL := NewTestUserDAL(testUserKind)

	// Create test users
	users := []*TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)

	// Track hook calls
	hookCalled := false
	userDAL.WillPut = func(ctx context.Context, u *TestUser) error {
		hookCalled = true
		// Modify the entity in the hook
		u.Name = "Modified by hook"
		return nil
	}

	user := &TestUser{
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

	userDAL := NewTestUserDAL(testUserKind)
	userDAL.Namespace = namespace

	user := &TestUser{
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
