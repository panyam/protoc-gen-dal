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
- [x] Refactor to ConversionType system (Assignment, Transformer, TransformerWithError, etc.)
- [x] Add Source/Target naming convention throughout codebase
- [x] Implement in-place conversion with dest parameter
- [x] Add fieldRef template helper for clean pointer/value handling
- [x] Implement optional keyword detection for GORM pointer generation
- [x] Add warnings for missing nested message converters

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
- [x] **TEST**: Repeated primitive as JSONB (Tags: []string with jsonb tag)
- [x] Implement applicative conversion for repeated primitives (direct assignment)
- [x] **TEST**: Repeated primitive as array (Categories: []string with text[] tag)
- [x] Implement applicative conversion for repeated primitives (direct assignment)
- [x] **TEST**: Map field as JSONB (Metadata: map[string]string with jsonb tag)
- [x] Implement applicative conversion for primitive maps (direct assignment)
- [x] **TEST**: Repeated message type with loop-based conversion (Library.Contributors)
- [x] Implement loop-based converter application for []MessageType
- [x] **TEST**: Map with message value type with loop-based conversion (Organization.Departments)
- [x] Implement loop-based converter application for map<K, MessageType>

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

**Focus:** Phase 2 - Core Features (Complex Types, Foreign Keys)

**Recently Completed:**
- ✅ Phase 1.3 & 1.4 - Complete GORM generator with advanced converter system
- ✅ Struct generation with pointer receiver TableName() methods
- ✅ Converter functions with decorator pattern and in-place conversion
- ✅ Built-in type converters (Timestamp ↔ int64, numeric types)
- ✅ Custom converter annotations (to_func/from_func with ConverterFunc message)
- ✅ Converter registry for nested message conversions
- ✅ ConversionType system with clear categorization (Assignment, Transformer, TransformerWithError)
- ✅ Source/Target naming convention for cross-target consistency
- ✅ Template helpers (fieldRef) for clean pointer/value handling
- ✅ Optional keyword detection for proper GORM pointer generation
- ✅ Warning system for missing nested converters
- ✅ One file per proto file: `user_gorm.go` (structs) + `user_converters.go` (converters)
- ✅ **Phase 2.2 (Complete)**: Applicative conversion for maps and repeated fields
  - ✅ `map<K, primitive>` uses direct assignment (e.g., `map[string]string`)
  - ✅ `[]primitive` uses direct assignment (e.g., `[]string`, `[]int32`)
  - ✅ `[]MessageType` uses loop-based conversion (e.g., `[]Author`)
  - ✅ `map<K, MessageType>` uses loop-based conversion (e.g., `map[string]Author`)
  - ✅ Added Product proto with Tags, Categories, Metadata, Ratings fields
  - ✅ Added Library proto with Contributors ([]Author) field
  - ✅ Added Organization proto with Departments (map<string, Author>) field
  - ✅ Tests verify round-trip conversion, nil handling, empty collections for all types
  - ✅ Early return prevents incorrect message converter lookup for primitives
  - ✅ Loop-based conversion reuses existing ConversionType enum - no new types needed
  - ✅ Fixed protoToGoType to handle map<K, MessageType> properly (use buildStructName for message values)

**Generated Code:**
From `tests/protos/gorm/user.proto`:
- `user_gorm.go`: 10 GORM structs (UserGORM, UserWithPermissions, UserWithCustomTimestamps, UserWithIndexes, UserWithDefaults, AuthorGORM, BlogGORM, ProductGORM, LibraryGORM, OrganizationGORM)
- `user_converters.go`: 20 converter functions (To/From for each type)

**Converter Features:**
- Smart field mapping with ConversionType categorization
- Priority: custom converters > built-in > skip (decorator handles)
- Nested conversions: Automatic via registry, requires explicit `source` annotation
- In-place conversion via dest parameter (avoids allocations)
- Pointer vs value handling: `optional` keyword → pointer, otherwise value
- Helpful warnings when converters missing for nested types
- **NEW**: Applicative conversion - check contained types in maps/repeated fields
  - Primitive values → direct assignment (copy entire map/slice)
  - Message values in slices → loop-based conversion with element converter
  - Message values in maps → loop-based conversion with value converter

**Next:**
1. **Phase 2.3**: Foreign key support with cross-file resolution
2. **Phase 2.4**: Composite keys
3. **Phase 3**: Additional targets (postgres-raw, firestore)

## Notes

- Each TODO item follows TDD: write test first, then implement
- Mark items with ✅ when test passes
- Refactor when needed before moving to next item
- Keep tests focused and atomic
- Integration tests at binary level, unit tests at package level
