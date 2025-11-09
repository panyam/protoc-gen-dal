# TODO: Implementation Roadmap

## Phase 1: Foundation (Current)

### 1.1 Update Proto Annotations ✅
- [x] Add `source` field to `TableOptions`
- [x] Add `source` field to `GormOptions`
- [x] Add `source` field to `DatastoreOptions`
- [x] Add `PostgresOptions` annotation
- [x] Add `FirestoreOptions` annotation
- [x] Regenerate proto Go code

### 1.2 Shared Collector Package ✅
- [x] **TEST**: Collector finds postgres messages across multiple files
- [x] Implement `collector.CollectMessages()`
- [x] **TEST**: Collector builds message index correctly
- [x] Implement `buildMessageIndex()`
- [x] **TEST**: Collector extracts PostgresOptions correctly
- [x] Implement `extractMessageInfo()` for postgres
- [x] **TEST**: Collector links source message via `source` field
- [x] Implement source message lookup

### 1.3 First Generator: GORM (Go) ✅
- [x] **TEST**: Simple message generates BookGORM struct
- [x] Implement basic GORM struct generation (template-based)
- [x] **TEST**: GORM tags pass through from gorm_tags annotation
- [x] Implement GORM tag extraction and rendering
- [x] **TEST**: Generates TableName() method with pointer receiver
- [x] Implement TableName() method generation
- [x] Add TargetGorm to collector
- [x] Implement extractGormInfo() function
- [x] **TEST**: Generates converter functions (BookToBookGORM, BookFromBookGORM)
- [x] Implement typed converter generation with decorator support
- [x] Implement built-in type converters (Timestamp ↔ int64, numeric types)
- [x] Implement custom converter annotations (to_func/from_func with ConverterFunc)
- [x] Implement converter registry for nested message conversions
- [x] Generate mustConvert helper functions for recursive conversions

### 1.4 First Binary: protoc-gen-dal-gorm ✅
- [x] **TEST**: Binary collects and generates from test protos
- [x] Create thin main.go that delegates to collector + generator
- [x] **TEST**: Integration test with buf generate
- [x] Wire up end-to-end flow
- [x] Generate one file per proto file (not per message)
- [x] Automatic collection of embedded message types
- [x] Package name extraction from buf-managed go_package

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

### 3.1 postgres-raw (Go + database/sql)
- [ ] **TEST**: Generates BookToPostgresRow converter
- [ ] Implement raw SQL converter (columns, values)
- [ ] **TEST**: Generates BookFromPostgresRow converter
- [ ] Implement row scanning
- [ ] Create `protoc-gen-dal-postgres-raw` binary

### 3.2 firestore (Go)
- [ ] Add collector support for TargetFirestore
- [ ] **TEST**: Generates BookToFirestore converter
- [ ] Implement Firestore document conversion
- [ ] **TEST**: Generates BookFromFirestore converter
- [ ] Implement reverse conversion
- [ ] Create `protoc-gen-dal-firestore` binary

### 3.3 mongodb (Go)
- [ ] Add collector support for TargetMongoDB
- [ ] **TEST**: Generates BookToBSON converter
- [ ] Implement BSON conversion
- [ ] Create `protoc-gen-dal-mongodb` binary

## Phase 4: Multi-Language Support

### 4.1 Python + postgres (psycopg2)
- [ ] **TEST**: Generates Python converter functions
- [ ] Implement Python code generation
- [ ] **TEST**: Type hints are correct
- [ ] Implement Python type mapping
- [ ] Create `protoc-gen-python-dal-postgres` binary

### 4.2 TypeScript + Prisma (ORM)
- [ ] **TEST**: Generates Prisma schema
- [ ] Implement Prisma schema generation
- [ ] Create `protoc-gen-ts-dal-prisma` binary

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

**Focus:** Phase 2 - Core Features or Additional Targets

**Recently Completed:**
- ✅ Phase 1.3 & 1.4 - Complete GORM generator with converters
- ✅ Struct generation with pointer receiver TableName() methods
- ✅ Converter functions with decorator pattern
- ✅ Built-in type converters (Timestamp ↔ int64, numeric types)
- ✅ Custom converter annotations (to_func/from_func with ConverterFunc message)
- ✅ Converter registry for nested message conversions
- ✅ Automatic recursive conversion via mustConvert helpers
- ✅ One file per proto file: `user_gorm.go` (structs) + `user_converters.go` (converters)

**Generated Code:**
From `tests/protos/gorm/user.proto`:
- `user_gorm.go`: 6 GORM structs + 1 embedded type (AuthorGORM)
- `user_converters.go`: 12 converter functions (To/From for each type) + mustConvert helpers

**Converter Features:**
- Smart field mapping (only converts compatible types)
- Priority: custom converters > built-in > skip (decorator handles)
- Nested conversions: `Blog.Author` automatically converts via `mustConvertAuthorToAuthorGORM`
- Skips incompatible fields (e.g., `DeletedAt` only in GORM) - decorator handles them

**Next:**
1. **Phase 2.1**: Field transformations for complex types (repeated fields, maps)
2. **Phase 2.2**: Foreign key support
3. **Phase 3**: Additional targets (postgres-raw, firestore)

## Notes

- Each TODO item follows TDD: write test first, then implement
- Mark items with ✅ when test passes
- Refactor when needed before moving to next item
- Keep tests focused and atomic
- Integration tests at binary level, unit tests at package level
