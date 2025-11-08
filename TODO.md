# TODO: Implementation Roadmap

## Phase 1: Foundation (Current)

### 1.1 Update Proto Annotations ✅
- [x] Add `source` field to `TableOptions`
- [x] Add `source` field to `GormOptions`
- [x] Add `source` field to `DatastoreOptions`
- [x] Add `PostgresOptions` annotation
- [x] Add `FirestoreOptions` annotation
- [x] Regenerate proto Go code

### 1.2 Shared Collector Package
- [ ] **TEST**: Collector finds postgres messages across multiple files
- [ ] Implement `collector.CollectMessages()`
- [ ] **TEST**: Collector builds message index correctly
- [ ] Implement `buildMessageIndex()`
- [ ] **TEST**: Collector extracts PostgresOptions correctly
- [ ] Implement `extractMessageInfo()` for postgres
- [ ] **TEST**: Collector links source message via `source` field
- [ ] Implement source message lookup

### 1.3 First Generator: postgres+gorm (Go)
- [ ] **TEST**: Simple message generates BookGORM struct
- [ ] Implement basic GORM struct generation
- [ ] **TEST**: Primary key generates correct GORM tag
- [ ] Implement primary key tag generation
- [ ] **TEST**: Column annotation generates correct GORM tag
- [ ] Implement column name/type tag generation
- [ ] **TEST**: Generates TableName() method
- [ ] Implement TableName() method generation
- [ ] **TEST**: Generates BookToGORM converter
- [ ] Implement typed converter generation
- [ ] **TEST**: Generates BookFromGORM converter
- [ ] Implement reverse converter generation

### 1.4 First Binary: protoc-gen-go-dal-postgres-gorm
- [ ] **TEST**: Binary collects and generates from test protos
- [ ] Create thin main.go that delegates to collector + generator
- [ ] **TEST**: Integration test with buf generate
- [ ] Wire up end-to-end flow

## Phase 2: Core Features

### 2.1 Field Transformations
- [ ] **TEST**: Timestamp field transforms to int64
- [ ] Implement default transformations in converter
- [ ] **TEST**: Decorator function allows custom transformation
- [ ] Implement decorator parameter in converters
- [ ] **TEST**: Nested message transforms to JSON
- [ ] Implement JSON serialization for complex types

### 2.2 Repeated/Map Fields
- [ ] **TEST**: Repeated primitive as JSONB
- [ ] Implement JSON storage strategy
- [ ] **TEST**: Repeated primitive as array (TEXT[])
- [ ] Implement array storage strategy
- [ ] **TEST**: Map field as JSONB
- [ ] Implement map to JSONB conversion

### 2.3 Foreign Keys
- [ ] **TEST**: Foreign key annotation generates correct GORM tag
- [ ] Implement foreign key tag generation
- [ ] **TEST**: Cross-file foreign key resolution
- [ ] Implement relationship analysis across messages
- [ ] **TEST**: CASCADE/RESTRICT actions in GORM tags
- [ ] Implement referential action mapping

### 2.4 Composite Keys
- [ ] **TEST**: Multiple primary_key fields generate composite key
- [ ] Implement composite primary key tags
- [ ] **TEST**: FindByCompositeKey method generation
- [ ] Implement composite key finder methods

### 2.5 Repository Pattern (Optional)
- [ ] **TEST**: Generates BookRepository struct
- [ ] Implement repository struct generation
- [ ] **TEST**: Repository.Create() calls hooks
- [ ] Implement lifecycle hooks as function pointers
- [ ] **TEST**: Repository.FindByID() works
- [ ] Implement basic CRUD methods
- [ ] **TEST**: Decorator works in repository methods
- [ ] Wire decorator through repository

## Phase 3: Additional Targets

### 3.1 postgres+raw (Go + database/sql)
- [ ] **TEST**: Generates BookToPostgresRow converter
- [ ] Implement raw SQL converter (columns, values)
- [ ] **TEST**: Generates BookFromPostgresRow converter
- [ ] Implement row scanning
- [ ] Create `protoc-gen-go-dal-postgres-raw` binary

### 3.2 firestore+raw (Go)
- [ ] Update annotations with FirestoreOptions
- [ ] **TEST**: Generates BookToFirestore converter
- [ ] Implement Firestore document conversion
- [ ] **TEST**: Generates BookFromFirestore converter
- [ ] Implement reverse conversion
- [ ] Create `protoc-gen-go-dal-firestore-raw` binary

### 3.3 mongodb+raw (Go)
- [ ] Add MongoDBOptions annotation
- [ ] **TEST**: Generates BookToBSON converter
- [ ] Implement BSON conversion
- [ ] Create `protoc-gen-go-dal-mongodb-raw` binary

## Phase 4: Multi-Language Support

### 4.1 Python + postgres+raw (psycopg2)
- [ ] **TEST**: Generates Python converter functions
- [ ] Implement Python code generation
- [ ] **TEST**: Type hints are correct
- [ ] Implement Python type mapping
- [ ] Create `protoc-gen-python-dal-postgres-raw` binary

### 4.2 TypeScript + postgres (Prisma)
- [ ] **TEST**: Generates Prisma schema
- [ ] Implement Prisma schema generation
- [ ] Create `protoc-gen-ts-dal-postgres-prisma` binary

## Phase 5: Advanced Features

### 5.1 Indexes
- [ ] **TEST**: Index annotation generates GORM index tag
- [ ] Implement index tag generation
- [ ] **TEST**: Composite index on multiple fields
- [ ] Implement multi-field index support

### 5.2 Soft Deletes
- [ ] **TEST**: GORM generates DeletedAt field
- [ ] Implement soft delete support
- [ ] **TEST**: disable_soft_delete works
- [ ] Respect GORM hints

### 5.3 Timestamps
- [ ] **TEST**: auto_create_time generates GORM autoCreateTime tag
- [ ] Implement auto timestamp tags
- [ ] **TEST**: auto_update_time generates GORM autoUpdateTime tag
- [ ] Implement auto update tags

### 5.4 Relationships (Future)
- [ ] Design relationship annotations
- [ ] **TEST**: One-to-many generates association
- [ ] Implement one-to-many support
- [ ] **TEST**: Many-to-many with join table
- [ ] Implement many-to-many support
- [ ] **TEST**: Lazy loading generates LoadX() method
- [ ] Implement lazy load method generation

### 5.5 DDL Generation (Optional)
- [ ] **TEST**: Generates CREATE TABLE SQL constant
- [ ] Implement DDL SQL generation
- [ ] **TEST**: Foreign key constraints in DDL
- [ ] Implement constraint generation
- [ ] **TEST**: Index creation in DDL
- [ ] Implement index DDL

## Phase 6: Documentation & Examples

### 6.1 Documentation
- [ ] README with quick start
- [ ] Annotation reference guide
- [ ] Target-specific guides (postgres, firestore, etc.)
- [ ] Multi-language examples

### 6.2 Examples
- [ ] Simple CRUD example (postgres+gorm)
- [ ] Foreign key example
- [ ] Composite key example
- [ ] Decorator transformation example
- [ ] Repository with hooks example
- [ ] Multi-target example (same proto → postgres + firestore)

### 6.3 Testing
- [ ] Integration tests with buf
- [ ] Golden file tests for generated code
- [ ] Cross-language consistency tests
- [ ] Performance benchmarks

## Current Sprint

**Focus:** Phase 1.2 - Shared Collector Package

**Next Test to Write:**
```go
// pkg/collector/collector_test.go
func TestCollectMessages_FindsPostgresMessages(t *testing.T) {
    // Given: Multiple proto files with postgres annotations
    // When: CollectMessages called with TargetPostgres
    // Then: Returns all postgres messages with correct metadata
}
```

## Notes

- Each TODO item follows TDD: write test first, then implement
- Mark items with ✅ when test passes
- Refactor when needed before moving to next item
- Keep tests focused and atomic
- Integration tests at binary level, unit tests at package level
