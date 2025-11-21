package gorm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	gormgen "github.com/panyam/protoc-gen-dal/tests/gen/gorm"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm/dal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	// Create temp file in /tmp
	tmpFile := filepath.Join(os.TempDir(), "test_protoc_gen_dal_"+t.Name()+"_*.db")
	f, err := os.CreateTemp(os.TempDir(), "test_protoc_gen_dal_"+t.Name()+"_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpFile = f.Name()
	f.Close()

	// Clean up on test completion
	t.Cleanup(func() {
		os.Remove(tmpFile)
	})

	// Open SQLite database
	db, err := gorm.Open(sqlite.Open(tmpFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	return db
}

// TestAutoMigrate tests that all generated models can be auto-migrated
func TestAutoMigrate(t *testing.T) {
	db := setupTestDB(t)

	// Test migrating all models that have a "table" in their option
	models := []any{
		&gormgen.UserGORM{},
		&gormgen.UserWithPermissions{},
		&gormgen.UserWithCustomTimestamps{},
		&gormgen.UserWithIndexes{},
		&gormgen.UserWithDefaults{},
		&gormgen.BlogGORM{},
		&gormgen.ProductGORM{},
		&gormgen.LibraryGORM{},
		&gormgen.OrganizationGORM{},
		&gormgen.WorldGORM{},
		&gormgen.WorldDataGORM{},
		&gormgen.DocumentGormExtra{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			t.Errorf("Failed to auto-migrate %T: %v", model, err)
		}
	}
}

// TestDALCreate tests the Create method
func TestDALCreate(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{}

	// Create a new user
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	err := userDAL.Create(ctx, db, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to create duplicate (should fail)
	duplicate := &gormgen.UserGORM{
		Id:    1,
		Name:  "Bob",
		Email: "bob@example.com",
	}

	err = userDAL.Create(ctx, db, duplicate)
	if err == nil {
		t.Error("Expected error when creating duplicate, got nil")
	}
}

// TestDALUpdate tests the Update method
func TestDALUpdate(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{}

	// Create initial user
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update the user
	user.Name = "Alice Updated"
	user.Age = 31
	err := userDAL.Update(ctx, db, user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	var retrieved gormgen.UserGORM
	db.First(&retrieved, 1)
	if retrieved.Name != "Alice Updated" || retrieved.Age != 31 {
		t.Errorf("Update didn't persist: got %v", retrieved)
	}

	// Try to update non-existent record
	nonExistent := &gormgen.UserGORM{
		Id:   999,
		Name: "Nobody",
	}
	err = userDAL.Update(ctx, db, nonExistent)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound, got: %v", err)
	}
}

// TestDALSave tests the Save (upsert) method
func TestDALSave(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{
		WillCreate: func(ctx context.Context, user *gormgen.UserGORM) error {
			user.MemberNumber = "MEMBER-001"
			return nil
		},
	}

	// Save new user (create)
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	err := userDAL.Save(ctx, db, user)
	if err != nil {
		t.Fatalf("Save (create) failed: %v", err)
	}

	// Verify WillCreate was called
	var retrieved gormgen.UserGORM
	db.First(&retrieved, 1)
	if retrieved.MemberNumber != "MEMBER-001" {
		t.Errorf("WillCreate hook not called: got MemberNumber=%s", retrieved.MemberNumber)
	}

	// Save existing user (update)
	user.Name = "Alice Updated"
	user.Age = 31
	err = userDAL.Save(ctx, db, user)
	if err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	// Verify update
	db.First(&retrieved, 1)
	if retrieved.Name != "Alice Updated" || retrieved.Age != 31 {
		t.Errorf("Save (update) didn't persist: got %v", retrieved)
	}
}

// TestDALGetAndDelete tests Get and Delete methods
func TestDALGetAndDelete(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{}

	// Create test user
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Alice",
		Email: "alice@example.com",
	}
	db.Create(user)

	// Test Get
	retrieved, err := userDAL.Get(ctx, db, 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Get returned nil for existing record")
	}
	if retrieved.Name != "Alice" {
		t.Errorf("Get returned wrong user: got %v", retrieved)
	}

	// Test Get non-existent (should return nil, nil)
	notFound, err := userDAL.Get(ctx, db, 999)
	if err != nil {
		t.Errorf("Get non-existent should not error: %v", err)
	}
	if notFound != nil {
		t.Errorf("Get non-existent should return nil, got: %v", notFound)
	}

	// Test Delete
	err = userDAL.Delete(ctx, db, 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	retrieved, _ = userDAL.Get(ctx, db, 1)
	if retrieved != nil {
		t.Error("Record still exists after delete")
	}
}

// TestDALListAndBatchGet tests List and BatchGet methods
func TestDALListAndBatchGet(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{}

	// Create multiple users
	users := []*gormgen.UserGORM{
		{Id: 1, Name: "Alice", Email: "alice@example.com", Age: 30},
		{Id: 2, Name: "Bob", Email: "bob@example.com", Age: 25},
		{Id: 3, Name: "Charlie", Email: "charlie@example.com", Age: 35},
		{Id: 4, Name: "Diana", Email: "diana@example.com", Age: 28},
	}
	for _, u := range users {
		db.Create(u)
	}

	// Test List with filter
	results, err := userDAL.List(ctx, db.Where("age >= ?", 30).Order("age"))
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].Name != "Alice" || results[1].Name != "Charlie" {
		t.Errorf("List returned unexpected results: %v", results)
	}

	// Test BatchGet
	ids := []uint32{1, 3, 999} // 999 doesn't exist
	batch, err := userDAL.BatchGet(ctx, db, ids)
	if err != nil {
		t.Fatalf("BatchGet failed: %v", err)
	}
	if len(batch) != 2 {
		t.Errorf("Expected 2 results from BatchGet, got %d", len(batch))
	}

	// Test BatchGet with empty slice
	empty, err := userDAL.BatchGet(ctx, db, []uint32{})
	if err != nil {
		t.Errorf("BatchGet with empty slice failed: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("BatchGet with empty slice should return empty, got %d results", len(empty))
	}
}

// TestOptimisticLocking tests conditional updates with timestamp checking
func TestOptimisticLocking(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{}

	// Create a user
	now := time.Now().Truncate(time.Second)
	user := &gormgen.UserGORM{
		Id:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		UpdatedAt: now,
	}
	db.Create(user)

	// Simulate concurrent update: Update with correct timestamp
	user.Name = "Alice Updated"
	user.Age = 31
	newTime := now.Add(time.Second)
	user.UpdatedAt = newTime
	err := userDAL.Update(ctx, db.Where("updated_at = ?", now), user)
	if err != nil {
		t.Fatalf("Update with correct timestamp failed: %v", err)
	}

	// Try to update with stale timestamp (should fail)
	user.Name = "Stale update"
	err = userDAL.Update(ctx, db.Where("updated_at = ?", now), user)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound for stale timestamp, got: %v", err)
	}

	// Verify name wasn't updated
	var retrieved gormgen.UserGORM
	db.First(&retrieved, 1)
	if retrieved.Name == "Stale update" {
		t.Error("Stale update should not have been applied")
	}
	if retrieved.Name != "Alice Updated" {
		t.Errorf("Expected 'Alice Updated', got: %s", retrieved.Name)
	}
}

// TestSaveWithOptimisticLocking tests Save with conditional updates
func TestSaveWithOptimisticLocking(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()
	userDAL := &dal.UserGORMDAL{}

	// Create a user
	now := time.Now().Truncate(time.Second)
	user := &gormgen.UserGORM{
		Id:        1,
		Name:      "Bob",
		Email:     "bob@example.com",
		Age:       25,
		UpdatedAt: now,
	}
	db.Create(user)

	// Save with correct timestamp (should succeed)
	user.Name = "Bob Updated"
	user.Age = 26
	newTime := now.Add(time.Second)
	user.UpdatedAt = newTime
	err := userDAL.Save(ctx, db.Where("updated_at = ?", now), user)
	if err != nil {
		t.Fatalf("Save with correct timestamp failed: %v", err)
	}

	// Try Save with stale timestamp (should fail)
	user.Name = "Stale save"
	err = userDAL.Save(ctx, db.Where("updated_at = ?", now), user)
	if err == nil {
		t.Error("Expected error for Save with stale timestamp, got nil")
	}
}

// TestWillCreateHook tests that WillCreate hook is called appropriately
func TestWillCreateHook(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	hookCalled := false
	userDAL := &dal.UserGORMDAL{
		WillCreate: func(ctx context.Context, user *gormgen.UserGORM) error {
			hookCalled = true
			user.MemberNumber = "AUTO-" + user.Name
			user.CreatedAt = time.Now()
			return nil
		},
	}

	// Test that hook is called during Save (create path)
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Alice",
		Email: "alice@example.com",
	}

	err := userDAL.Save(ctx, db, user)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !hookCalled {
		t.Error("WillCreate hook was not called")
	}

	// Verify hook modifications were applied
	var retrieved gormgen.UserGORM
	db.First(&retrieved, 1)
	if retrieved.MemberNumber != "AUTO-Alice" {
		t.Errorf("Hook modifications not applied: got MemberNumber=%s", retrieved.MemberNumber)
	}

	// Hook should not be called on update
	hookCalled = false
	user.Name = "Alice Updated"
	err = userDAL.Save(ctx, db, user)
	if err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	if hookCalled {
		t.Error("WillCreate hook should not be called on update")
	}
}

// TestWillCreateHookError tests that errors from WillCreate prevent creation
func TestWillCreateHookError(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	userDAL := &dal.UserGORMDAL{
		WillCreate: func(ctx context.Context, user *gormgen.UserGORM) error {
			if user.Age < 18 {
				return errors.New("user must be 18 or older")
			}
			return nil
		},
	}

	// Try to create user under 18 (should fail)
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Minor",
		Email: "minor@example.com",
		Age:   16,
	}

	err := userDAL.Save(ctx, db, user)
	if err == nil {
		t.Error("Expected error from WillCreate hook, got nil")
	}
	if err.Error() != "user must be 18 or older" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify record was not created
	var count int64
	db.Model(&gormgen.UserGORM{}).Count(&count)
	if count != 0 {
		t.Error("Record was created despite WillCreate error")
	}
}
