package gorm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	gormgen "github.com/panyam/protoc-gen-dal/tests/gen/gorm"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm/dal"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a database connection for testing.
// If PROTOC_GEN_DAL_TEST_PGDB environment variable is set, it connects to PostgreSQL.
// Otherwise, it creates a temporary SQLite database.
//
// PostgreSQL environment variables:
//   - PROTOC_GEN_DAL_TEST_PGDB: Database name (required to use PostgreSQL)
//   - PROTOC_GEN_DAL_TEST_PGHOST: Host (default: localhost)
//   - PROTOC_GEN_DAL_TEST_PGPORT: Port (default: 5432)
//   - PROTOC_GEN_DAL_TEST_PGUSER: Username (default: postgres)
//   - PROTOC_GEN_DAL_TEST_PGPASSWORD: Password (default: empty)
func setupTestDB(t *testing.T) *gorm.DB {
	pgDB := os.Getenv("PROTOC_GEN_DAL_TEST_PGDB")
	if pgDB != "" {
		return setupPostgresDB(t, pgDB)
	}
	return setupSQLiteDB(t)
}

// setupSQLiteDB creates a temporary SQLite database for testing
func setupSQLiteDB(t *testing.T) *gorm.DB {
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
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	return db
}

// setupPostgresDB connects to a PostgreSQL database for testing.
// It creates a test schema for isolation and drops it on cleanup.
func setupPostgresDB(t *testing.T, dbName string) *gorm.DB {
	host := os.Getenv("PROTOC_GEN_DAL_TEST_PGHOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PROTOC_GEN_DAL_TEST_PGPORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("PROTOC_GEN_DAL_TEST_PGUSER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("PROTOC_GEN_DAL_TEST_PGPASSWORD")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		host, port, user, dbName)
	if password != "" {
		dsn += " password=" + password
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Create a unique test schema for isolation
	// Using test name (sanitized) to allow parallel tests
	schemaName := "test_" + sanitizeSchemaName(t.Name())

	// Create schema
	if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Set search path to use the test schema
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error; err != nil {
		t.Fatalf("Failed to set search_path: %v", err)
	}

	// Clean up schema on test completion
	t.Cleanup(func() {
		db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	})

	return db
}

// sanitizeSchemaName converts a test name to a valid PostgreSQL schema name
func sanitizeSchemaName(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	// Ensure it doesn't start with a number
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = append([]byte("s_"), result...)
	}
	// Limit length (PostgreSQL max identifier is 63 chars)
	if len(result) > 60 {
		result = result[:60]
	}
	return string(result)
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
	if err == nil { // this is an auto update time field so cannot be set
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

// TestDALLateBinding tests the DAL's TableName override functionality.
// This demonstrates using the same struct with different tables at runtime.
func TestDALLateBinding(t *testing.T) {
	db := setupTestDB(t)

	// Create two separate tables using AutoMigrate with custom table names
	// This ensures cross-database compatibility (SQLite and PostgreSQL)
	if err := db.Table("user_addresses").AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to create user_addresses table: %v", err)
	}

	if err := db.Table("company_addresses").AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to create company_addresses table: %v", err)
	}

	ctx := context.Background()

	// Create two DAL instances pointing to different tables
	userAddrDAL := dal.NewUserGORMDAL("user_addresses")
	companyAddrDAL := dal.NewUserGORMDAL("company_addresses")

	// Create records in both tables using the same struct type
	userAddr := &gormgen.UserGORM{
		Id:    1,
		Name:  "Home Address",
		Email: "home@example.com",
		Age:   1,
	}

	companyAddr := &gormgen.UserGORM{
		Id:    1, // Same ID but different table
		Name:  "Office Address",
		Email: "office@example.com",
		Age:   2,
	}

	// Create in user_addresses
	if err := userAddrDAL.Create(ctx, db, userAddr); err != nil {
		t.Fatalf("Failed to create in user_addresses: %v", err)
	}

	// Create in company_addresses
	if err := companyAddrDAL.Create(ctx, db, companyAddr); err != nil {
		t.Fatalf("Failed to create in company_addresses: %v", err)
	}

	// Retrieve from both tables and verify isolation
	retrieved, err := userAddrDAL.Get(ctx, db, 1)
	if err != nil {
		t.Fatalf("Failed to get from user_addresses: %v", err)
	}
	if retrieved.Name != "Home Address" {
		t.Errorf("Wrong record from user_addresses: got %s, want 'Home Address'", retrieved.Name)
	}

	retrieved, err = companyAddrDAL.Get(ctx, db, 1)
	if err != nil {
		t.Fatalf("Failed to get from company_addresses: %v", err)
	}
	if retrieved.Name != "Office Address" {
		t.Errorf("Wrong record from company_addresses: got %s, want 'Office Address'", retrieved.Name)
	}

	// Update in one table shouldn't affect the other
	userAddr.Name = "Updated Home"
	if err := userAddrDAL.Save(ctx, db, userAddr); err != nil {
		t.Fatalf("Failed to update user_addresses: %v", err)
	}

	// Verify company_addresses unchanged
	retrieved, _ = companyAddrDAL.Get(ctx, db, 1)
	if retrieved.Name != "Office Address" {
		t.Errorf("Company address was changed unexpectedly: got %s", retrieved.Name)
	}

	// Delete from one table
	if err := userAddrDAL.Delete(ctx, db, 1); err != nil {
		t.Fatalf("Failed to delete from user_addresses: %v", err)
	}

	// Verify user_addresses is empty but company_addresses still has the record
	retrieved, _ = userAddrDAL.Get(ctx, db, 1)
	if retrieved != nil {
		t.Error("Record should be deleted from user_addresses")
	}

	retrieved, _ = companyAddrDAL.Get(ctx, db, 1)
	if retrieved == nil {
		t.Error("Record should still exist in company_addresses")
	}
}

// TestDALTableNameEmpty tests that empty TableName uses the struct's TableName() method
func TestDALTableNameEmpty(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.UserGORM{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	// Create DAL with empty TableName (uses struct's TableName() method)
	userDAL := dal.NewUserGORMDAL("")

	// Create user using the default table
	user := &gormgen.UserGORM{
		Id:    1,
		Name:  "Default Table User",
		Email: "default@example.com",
		Age:   25,
	}

	if err := userDAL.Create(ctx, db, user); err != nil {
		t.Fatalf("Failed to create with empty TableName: %v", err)
	}

	// Verify record exists in the default table ("users" via TableName() method)
	var count int64
	db.Table("users").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record in 'users' table, got %d", count)
	}
}

// TestTestRecord1WithEnums tests AutoMigrate and CRUD operations for TestRecord1GORM
// which contains enum, repeated enum, and map<string, enum> fields.
func TestTestRecord1WithEnums(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.TestRecord1GORM{}); err != nil {
		t.Fatalf("Failed to auto-migrate TestRecord1GORM: %v", err)
	}

	// Create a record with all enum field types
	record := &gormgen.TestRecord1GORM{
		TimeField:   time.Now().Truncate(time.Second),
		ExtraData:   []byte("test data"),
		AnEnum:      api.SampleEnum_SAMPLE_ENUM_B,
		ListOfEnums: []api.SampleEnum{
			api.SampleEnum_SAMPLE_ENUM_A,
			api.SampleEnum_SAMPLE_ENUM_B,
			api.SampleEnum_SAMPLE_ENUM_C,
		},
		MapStringToEnum: map[string]api.SampleEnum{
			"key1": api.SampleEnum_SAMPLE_ENUM_A,
			"key2": api.SampleEnum_SAMPLE_ENUM_B,
			"key3": api.SampleEnum_SAMPLE_ENUM_C,
		},
	}

	// Create
	if err := db.Create(record).Error; err != nil {
		t.Fatalf("Failed to create TestRecord1GORM: %v", err)
	}

	// Read back
	var retrieved gormgen.TestRecord1GORM
	if err := db.First(&retrieved).Error; err != nil {
		t.Fatalf("Failed to retrieve TestRecord1GORM: %v", err)
	}

	// Verify enum field
	if retrieved.AnEnum != record.AnEnum {
		t.Errorf("AnEnum mismatch: got %v, want %v", retrieved.AnEnum, record.AnEnum)
	}

	// Verify repeated enum field
	if len(retrieved.ListOfEnums) != len(record.ListOfEnums) {
		t.Errorf("ListOfEnums length mismatch: got %d, want %d", len(retrieved.ListOfEnums), len(record.ListOfEnums))
	} else {
		for i, v := range retrieved.ListOfEnums {
			if v != record.ListOfEnums[i] {
				t.Errorf("ListOfEnums[%d] mismatch: got %v, want %v", i, v, record.ListOfEnums[i])
			}
		}
	}

	// Verify map<string, enum> field
	if len(retrieved.MapStringToEnum) != len(record.MapStringToEnum) {
		t.Errorf("MapStringToEnum length mismatch: got %d, want %d", len(retrieved.MapStringToEnum), len(record.MapStringToEnum))
	} else {
		for k, v := range record.MapStringToEnum {
			if retrieved.MapStringToEnum[k] != v {
				t.Errorf("MapStringToEnum[%s] mismatch: got %v, want %v", k, retrieved.MapStringToEnum[k], v)
			}
		}
	}
}

// TestTestRecord1EnumRoundTrip tests conversion between API and GORM types with enum fields
func TestTestRecord1EnumRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&gormgen.TestRecord1GORM{}); err != nil {
		t.Fatalf("Failed to auto-migrate TestRecord1GORM: %v", err)
	}

	// Create via API -> GORM conversion (using the test from testany_converters_test.go pattern)
	// Note: This tests that the GORM struct can be stored and retrieved correctly
	gormRecord := &gormgen.TestRecord1GORM{
		TimeField:   time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		ExtraData:   []byte("test extra data"),
		AnEnum:      api.SampleEnum_SAMPLE_ENUM_C,
		ListOfEnums: []api.SampleEnum{
			api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED,
			api.SampleEnum_SAMPLE_ENUM_A,
			api.SampleEnum_SAMPLE_ENUM_B,
			api.SampleEnum_SAMPLE_ENUM_C,
		},
		MapStringToEnum: map[string]api.SampleEnum{
			"unspecified": api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED,
			"a":           api.SampleEnum_SAMPLE_ENUM_A,
			"b":           api.SampleEnum_SAMPLE_ENUM_B,
			"c":           api.SampleEnum_SAMPLE_ENUM_C,
		},
	}

	// Store in database
	if err := db.Create(gormRecord).Error; err != nil {
		t.Fatalf("Failed to create: %v", err)
	}

	// Retrieve from database
	var retrieved gormgen.TestRecord1GORM
	if err := db.First(&retrieved).Error; err != nil {
		t.Fatalf("Failed to retrieve: %v", err)
	}

	// Verify all fields
	if !retrieved.TimeField.Equal(gormRecord.TimeField) {
		t.Errorf("TimeField mismatch: got %v, want %v", retrieved.TimeField, gormRecord.TimeField)
	}

	if string(retrieved.ExtraData) != string(gormRecord.ExtraData) {
		t.Errorf("ExtraData mismatch: got %s, want %s", retrieved.ExtraData, gormRecord.ExtraData)
	}

	if retrieved.AnEnum != gormRecord.AnEnum {
		t.Errorf("AnEnum mismatch: got %v, want %v", retrieved.AnEnum, gormRecord.AnEnum)
	}

	// Check ListOfEnums
	if len(retrieved.ListOfEnums) != 4 {
		t.Errorf("ListOfEnums length: got %d, want 4", len(retrieved.ListOfEnums))
	}

	// Check MapStringToEnum
	if len(retrieved.MapStringToEnum) != 4 {
		t.Errorf("MapStringToEnum length: got %d, want 4", len(retrieved.MapStringToEnum))
	}
	if retrieved.MapStringToEnum["c"] != api.SampleEnum_SAMPLE_ENUM_C {
		t.Errorf("MapStringToEnum['c']: got %v, want %v", retrieved.MapStringToEnum["c"], api.SampleEnum_SAMPLE_ENUM_C)
	}
}
