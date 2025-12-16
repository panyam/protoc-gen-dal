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

### 2.1 Field Transformations ✅ (Already Complete)
**SKIPPED - Already implemented in Phase 1.3**
- ✅ Built-in converters handle Timestamp ↔ int64 and numeric transformations
- ✅ Decorator pattern allows custom transformations
- ✅ JSONB storage for complex types via gorm_tags: ["type:jsonb"]
- No additional work needed - existing system covers all requirements

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

### 2.3 Foreign Keys ✅ (Already Supported)
**SKIPPED - Use native GORM tags directly, no abstraction needed**
- ✅ Foreign keys work via gorm_tags: `["foreignKey:AuthorID", "references:ID"]`
- ✅ Constraints via gorm_tags: `["constraint:OnDelete:CASCADE,OnUpdate:CASCADE"]`
- No special `foreign_key` annotation needed - violates "no abstraction layer" principle
- Users already know GORM syntax - just pass it through
- Example:
  ```protobuf
  uint32 author_id = 1 [(dal.v1.column) = {
    gorm_tags: ["foreignKey:AuthorID", "references:ID", "constraint:OnDelete:CASCADE"]
  }];
  ```

### 2.4 Composite Keys ✅ (Already Supported)
**SKIPPED - Use native GORM tags directly, no abstraction needed**
- ✅ Composite keys work via multiple gorm_tags: `["primaryKey"]` on multiple fields
- ✅ GORM handles composite keys automatically when multiple fields have primaryKey tag
- No special implementation needed - already works with existing tag pass-through
- Example:
  ```protobuf
  string book_id = 1 [(dal.v1.column) = { gorm_tags: ["primaryKey"] }];
  int32 edition_number = 2 [(dal.v1.column) = { gorm_tags: ["primaryKey"] }];
  ```
- FindByCompositeKey method generation deferred to Phase 2.5 (Repository Pattern) if needed

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

### 3.1 Google Cloud Datastore (Go) ✅ COMPLETE
- [x] **TEST**: Collector finds Datastore messages
- [x] Add TargetDatastore constant and extractDatastoreInfo
- [x] **TEST**: Generates UserDatastore entity struct
- [x] Implement basic struct generation with datastore tags
- [x] **TEST**: Kind() method returns correct kind name
- [x] Implement Kind() method template
- [x] Create `protoc-gen-dal-datastore` binary
- [x] Update Makefile build targets
- [x] **TEST**: UserToUserDatastore converter (scalar fields)
- [x] Implement converter generation with type conversions (uint32↔string, Timestamp↔int64)
- [x] **TEST**: Integration test with buf generate (basic scalar fields)
- [x] Generated code compiles and passes all tests
- [x] Add test protos with repeated/map fields (mirror GORM test coverage)
- [x] **TEST**: Repeated scalar fields ([]string, []int32)
- [x] **TEST**: Repeated message fields ([]Author)
- [x] **TEST**: Map with scalar values (map[string]string)
- [x] **TEST**: Map with message values (map[string]Author)
- [x] Implement repeated/map field support in buildFieldMapping
- [x] Enhance fieldType() to handle repeated/map/message types
- [x] Add converterRegistry for nested message lookups
- [x] Template support for loop-based conversions
- ⏸️ LoadKey/SaveKey methods (deferred - not essential for MVP, can add later)

**Design Decision**: Skipped LoadKey/SaveKey PropertyLoadSaver implementation for MVP
- Not required for basic Datastore usage
- Users can manually manage keys (simple and explicit)
- PropertyLoadSaver only needed for advanced custom property transformations
- Can be added later if users request it

### 3.2 postgres-raw (Go + database/sql)
- [ ] **TEST**: Generates BookToPostgresRow converter
- [ ] Implement raw SQL converter (columns, values)
- [ ] **TEST**: Generates BookFromPostgresRow converter
- [ ] Implement row scanning
- [ ] Create `protoc-gen-dal-postgres-raw` binary

### 3.3 firestore (Go)
- [ ] Add collector support for TargetFirestore (annotation already exists)
- [ ] **TEST**: Generates BookToFirestore converter
- [ ] Implement Firestore document conversion
- [ ] **TEST**: Generates BookFromFirestore converter
- [ ] Implement reverse conversion
- [ ] Create `protoc-gen-dal-firestore` binary

### 3.4 mongodb (Go)
- [ ] Add collector support for TargetMongoDB (annotation already exists)
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

**Phase 2 Status:**
- ✅ Phase 2.1 - Already complete via built-in converters and decorators
- ✅ Phase 2.2 - Complete (applicative conversion for collections)
- ✅ Phase 2.3 - Already supported via native gorm_tags (no abstraction needed)
- ✅ Phase 2.4 - Already supported via native gorm_tags (no abstraction needed)
- ⏸️ Phase 2.5 - Repository Pattern (Optional - deferred)

**Phase 3 Status:**
- ✅ Phase 3.0 - Refactor Common Generator Utilities (COMPLETE)
  - ✅ Create `pkg/generator/common` package
    - ✅ File organization utilities (grouping, naming)
    - ✅ Package name extraction
    - ✅ Proto→Go type mapping (centralized map field handling)
    - ✅ Import management with deduplication
  - ✅ Create `pkg/generator/registry` package
    - ✅ Converter registry for tracking available converters
    - ✅ Prevents generating calls to non-existent converters
  - ✅ Unit tests for all shared utilities (100% coverage)
  - ✅ Migration of GORM generator to use shared utilities (removed 214 lines)
  - ✅ Migration of Datastore generator to use shared utilities (removed similar duplicate code)

- ✅ Phase 3.0b - Extract Converter Strategy Utilities (COMPLETE)
  - ✅ Create `pkg/generator/converter` package
    - ✅ Collection utilities (CheckMapValueType, CheckRepeatedElementType, BuildNestedConverterName)
    - ✅ Conversion detection (IsTimestampToInt64, IsNumericConversion, BuildNumericCast)
    - ✅ Template helper generators (TimestampHelperFunctions, MustParseUintHelper)
    - ✅ Render strategy system (ConversionType, FieldRenderStrategy, DetermineRenderStrategy)
  - ✅ Unit tests for all converter utilities
  - ✅ GORM generator calculates render strategies for all field mappings
  - ✅ Datastore generator calculates render strategies for all field mappings
  - ✅ All existing tests passing (no behavioral changes)

- ✅ Phase 3.0c - Template Unification: Struct Literal + Setters Pattern (COMPLETE)
  - ✅ GORM template rewrite
    - ✅ Updated ConverterData with field groups (ToTarget/FromTarget Inline/Setter/Loop)
    - ✅ Added template helper functions (isInlineValue, isSetterSimple, etc.)
    - ✅ Struct literal initialization for inline fields
    - ✅ Post-construction setters for pointer/error handling
    - ✅ Separate loop sections for repeated/map conversions
    - ✅ Handle empty ToTargetCode/FromTargetCode (fallback to src.Field)
  - ✅ Datastore template rewrite
    - ✅ Updated FieldMapping with ConversionType/RenderStrategy/SourcePkgName
    - ✅ Updated ConverterData with field groups
    - ✅ Modified buildFieldMapping to calculate render strategies
    - ✅ Added addRenderStrategies helper function
    - ✅ Modified buildConverterData to classify fields
    - ✅ Rewrote converters.go.tmpl with same pattern as GORM
    - ✅ All tests passing

- ✅ Phase 3.1 - Google Cloud Datastore (COMPLETE)
  - ✅ Collector support (TargetDatastore, extractDatastoreInfo)
  - ✅ Basic struct generation with datastore tags
  - ✅ Kind() method generation
  - ✅ Binary created (protoc-gen-dal-datastore)
  - ✅ Converter generation with type conversions (uint32↔string, Timestamp↔time.Time)
  - ✅ Integration test with buf generate (all tests pass)
  - ✅ Full repeated/map field support (primitives and messages)
  - ✅ Converter registry for nested message conversions
  - ✅ Loop-based conversion for repeated/map message types
  - ⏸️ LoadKey/SaveKey deferred (not essential for MVP)

- ✅ Phase 3.1b - Timestamp Migration to time.Time (COMPLETE)
  - ✅ Updated ProtoFieldToGoType() to map google.protobuf.Timestamp → time.Time
  - ✅ Renamed helper functions: timestampToInt64 → timestampToTime, int64ToTimestamp → timeToTimestamp
  - ✅ Updated GORM and Datastore generators to detect Timestamp→Timestamp conversion
  - ✅ Fixed pointer status handling for time.Time fields (value type, not pointer)
  - ✅ Added "time" package import to generated struct files
  - ✅ Updated all templates to use new helper function names
  - ✅ Updated proto test definitions to use google.protobuf.Timestamp
  - ✅ Updated unit tests in pkg/generator/converter
  - ✅ All tests passing (GORM, Datastore, converter utilities)

- ✅ Phase 3.1c - Shared Converters Package (COMPLETE)
  - ✅ Created pkg/converters with TimestampToTime, TimeToTimestamp, MustParseUint
  - ✅ Updated GORM template to import converters package instead of declaring helpers
  - ✅ Updated Datastore template to import converters package
  - ✅ Updated GORM generator to call converters.TimestampToTime(), converters.TimeToTimestamp()
  - ✅ Updated Datastore generator to call converters.MustParseUint()
  - ✅ Removed duplicate helper function declarations from generated files
  - ✅ All tests passing

- ✅ Phase 3.1d - Field Merging (Opt-Out Model) (COMPLETE)
  - ✅ Created pkg/generator/common/field_merge.go with MergeSourceFields, HasSkipField, ValidateFieldMerge
  - ✅ Field matching by NAME not NUMBER (allows proto field renumbering)
  - ✅ Empty target messages inherit all source fields automatically
  - ✅ Target fields override source fields with same name
  - ✅ skip_field annotation excludes fields from struct and converters
  - ✅ Validation ensures skip_field references exist in source
  - ✅ Updated GORM buildStructData to merge fields before generation
  - ✅ Updated GORM buildConverterData to use merged fields (respects skip_field)
  - ✅ Updated Datastore buildStructData to merge fields before generation
  - ✅ Updated Datastore buildConverterData to use merged fields (respects skip_field)
  - ✅ Embedded types generated once in _embedded_gorm.go (prevents duplicates)
  - ✅ Test protos created: DocumentGormEmpty, DocumentGormPartial, DocumentGormSkip, DocumentGormExtra
  - ✅ All tests passing

- ✅ Phase 3.1e - Generic Type Mapping Registry (COMPLETE)
  - ✅ Created pkg/generator/converter/type_mappings.go
  - ✅ Centralized type conversion registry for all generators
  - ✅ Registered conversions:
    - google.protobuf.Timestamp → int64 (Unix epoch) - NEW for weewar
    - google.protobuf.Timestamp → time.Time (native time type)
    - uint32 → string (ID conversions for Datastore)
  - ✅ Added converters.TimestampToInt64() and converters.Int64ToTimestamp()
  - ✅ Updated GORM buildFieldConversion to use generic registry
  - ✅ Updated Datastore buildFieldMapping to use generic registry
  - ✅ 100% backward compatible - no migration needed
  - ✅ Easy to extend with RegisterTypeMapping()
  - ✅ All tests passing (protoc-gen-dal and weewar)

- ✅ Phase 3.1f - Well-Known Type System & google.protobuf.Any Support (COMPLETE)
  - ✅ Created well-known type registry in pkg/generator/common/types.go
  - ✅ Added WellKnownTypeMapping structure (ProtoFullName, GoType, GoImport)
  - ✅ Registered mappings:
    - google.protobuf.Timestamp → time.Time
    - google.protobuf.Any → []byte
  - ✅ Provided RegisterWellKnownType() for easy extension
  - ✅ Updated ProtoFieldToGoType() to use registry for all well-known types
  - ✅ Implemented two-level type system:
    - Level 1: Well-known type registry (proto → Go type in generated structs)
    - Level 2: Type conversion registry (source proto-gen → target generated)
  - ✅ Created getSourceTypeKey() - returns proto type names (e.g., google.protobuf.Any)
  - ✅ Created getTargetTypeKey() - applies well-known mappings (e.g., google.protobuf.Any → bytes)
  - ✅ Added google.protobuf.Any → bytes mapping with converters.AnyToBytes/BytesToAny
  - ✅ Implemented pkg/converters/any.go with proto.Marshal/Unmarshal
  - ✅ Fixed GORM template to support inline converter code (ToTargetCode) for ConvertByTransformerWithError
  - ✅ Fixed Datastore template to support inline converter code for ConvertByTransformerWithError
  - ✅ Templates now check ToTargetConverterFunc OR ToTargetCode (both directions)
  - ✅ Added comprehensive tests for well-known type registry
  - ✅ All tests passing (testany.proto and weewar)

- ✅ Phase 3.1g - Enum Type Support (COMPLETE)
  - ✅ Added enum type detection in ProtoFieldToGoType (field.Enum != nil)
  - ✅ Enum types properly qualified with source package (e.g., api.SampleEnum)
  - ✅ Added sourcePkgName parameter to ProtoFieldToGoType for package qualification
  - ✅ Updated GORM generator to collect source package imports for enum types
  - ✅ Updated Datastore generator to collect source package imports for enum types
  - ✅ Changed template Imports from []string to []common.ImportSpec for aliased imports
  - ✅ Updated templates to render imports with aliases ({{ if .Alias }}{{ .Alias }} {{ end }}"{{ .Path }}")
  - ✅ Fixed package alias extraction - use GetPackageAlias() instead of ExtractPackageName()
  - ✅ Consistent package aliases across struct generation and converters
  - ✅ Enums converted as simple value types (direct assignment in converters)
  - ✅ Added SampleEnum to testany.proto for testing
  - ✅ All tests passing - generated code compiles and enums work correctly

**Generated Code Features:**
From `tests/protos/datastore/user.proto`:
- `user.go`: 8 Datastore entity structs (UserDatastore, UserWithNamespace, UserWithLargeText, UserSimple, AuthorDatastore, ProductDatastore, LibraryDatastore, OrganizationDatastore)
- `user_converters.go`: 16 converter functions (To/From for each type except embedded AuthorDatastore)
- Repeated scalars: Direct assignment ([]string, []int32)
- Map with scalars: Direct assignment with nil check (map[string]string)
- Repeated messages: Loop-based conversion ([]AuthorDatastore)
- Map with messages: Loop-based conversion (map[string]AuthorDatastore)
- Proper bidirectional handling of pointer types

- ✅ Phase 3.1h - Source Message Validation (COMPLETE)
  - ✅ Added error reporting when source message references don't exist
  - ✅ Collector now returns error instead of silently skipping invalid messages
  - ✅ All extract functions (GORM, Postgres, Firestore, MongoDB, Datastore) validate source exists
  - ✅ Clear error messages with guidance: "define a message with 'source: \"...\""
  - ✅ Generator fails fast with all errors reported at once
  - ✅ Updated all main.go files to handle CollectMessages error
  - ✅ Updated all tests to handle error return value
  - ✅ Tested with weewar proto - correctly identifies missing action types
  - ✅ Documented workaround for oneof fields: use skip_field on individual oneof members

**Design Decision:** Source validation prevents silent failures
- Previously: Missing source messages were silently skipped (returned nil)
- Now: Missing source messages cause immediate build failure with clear error
- Benefits: Catches typos early, forces explicit field skipping, better DX
- Example error: `message 'gorm.BookGORM' references source 'api.Bookk' which does not exist`

**Oneof Field Handling:**
- Oneof creates individual fields (not a single field with oneof name)
- Example: `oneof move_type { MoveUnitAction move_unit = 4; }` creates field named "move_unit"
- Two approaches to handle oneof in GORM:
  1. **Define GORM mappings** for each oneof member type (recommended for complex types)
     ```protobuf
     message MoveUnitActionGORM {
       option (dal.v1.gorm) = { source: "api.MoveUnitAction" };
     }
     ```
  2. **Skip individual oneof fields** and use google.protobuf.Any instead
     ```protobuf
     message GameMoveGORM {
       google.protobuf.Any move_type = 1;
       bool move_unit = 4 [(dal.v1.skip_field) = true];
       bool attack_unit = 5 [(dal.v1.skip_field) = true];
     }
     ```
- Future enhancement: Message-level `skip_fields: ["field1", "field2"]` annotation

- [x] **BUG FIX: Map Fields with Non-String Keys** (Issue #001) - COMPLETE
  - [x] **Symptom**: Compilation error for maps with int32, int64, uint32, bool keys containing message values
  - [x] **Root Cause**: Templates hardcode `map[string]` in converter generation
    - Line 102 in pkg/gorm/templates/converters.go.tmpl: `make(map[string]{{ .TargetElementType }}, ...)`
    - Line 212 in pkg/gorm/templates/converters.go.tmpl: `make(map[string]*{{ .SourcePkgName }}.{{ .SourceElementType }}, ...)`
    - Same issue in pkg/datastore/templates/converters.go.tmpl lines 112, 232
  - [x] **Fix Applied**:
    1. Added `MapKeyType` field to FieldMapping struct in pkg/generator/converter/field_mapping.go
    2. Extract map key type in BuildMapFieldMapping() using ProtoKindToGoType()
    3. Updated GORM template (lines 102, 212) to use `{{ .MapKeyType }}`
    4. Updated Datastore template (lines 112, 232) to use `{{ .MapKeyType }}`
    5. Changed `%s` format specifier to `%v` in error messages for non-string keys
  - [x] **Test Case**: tests/protos/api/testany.proto TestRecord2 with map<int32, MapValueMessage>, etc.
  - [x] **Tests Added**: 7 new tests in testany_converters_test.go covering int32, int64, uint32, bool keys
  - [x] **GitHub Issue**: https://github.com/panyam/protoc-gen-dal/issues/1

- ✅ Phase 3.1i - Shared Test Utilities (COMPLETE)
  - ✅ Created pkg/generator/testutil package for test helper functions
  - ✅ Extracted duplicate test utilities from GORM and Datastore test files
  - ✅ Defined TestProtoSet, TestFile, TestMessage, TestField structures
  - ✅ Implemented CreateTestPlugin() for building protogen.Plugin from test data
  - ✅ Implemented BuildCodeGeneratorRequest() and BuildFileDescriptor()
  - ✅ Support for both GormOpts and DatastoreOpts in TestMessage
  - ✅ Updated pkg/gorm/generator_test.go to use testutil package
  - ✅ Updated pkg/datastore/generator_test.go to use testutil package
  - ✅ Removed ~400 lines of duplicate code
  - ✅ All tests passing

**Design Decision:** Test utilities in shared package
- Test helper functions were duplicated across gorm/generator_test.go and datastore/generator_test.go
- Extracting to pkg/generator/testutil enables consistent test patterns across all generators
- TestMessage structure supports both GormOpts and DatastoreOpts for target-agnostic test data
- Future generators (postgres-raw, firestore, mongodb) can immediately use these utilities
- Centralized proto descriptor building logic reduces maintenance when proto structures change

- ✅ Phase 3.1j - Shared Converter Utilities (COMPLETE)
  - ✅ Created pkg/generator/converter/classification.go for field classification
    - ✅ FieldWithStrategy interface for generic field classification
    - ✅ ClassifyFields() generic function using type parameters
    - ✅ ClassifiedFields structure with ToTarget/FromTarget groups (inline/setter/loop)
    - ✅ AddRenderStrategies() function for calculating render strategies
  - ✅ Created pkg/generator/common/package_info.go for package extraction
    - ✅ PackageInfo structure (not SourcePackageInfo - generic naming)
    - ✅ ExtractPackageInfo() extracts clean import path and alias
    - ✅ Handles ;packagename suffix stripping automatically
  - ✅ Created pkg/generator/common/custom_converters.go for converter imports
    - ✅ CollectCustomConverterImports() scans to_func/from_func annotations
    - ✅ ExtractCustomConverters() generates converter code from ColumnOptions
    - ✅ Works with both GORM and Datastore (target-agnostic)
  - ✅ Added FieldWithStrategy interface implementation to GORM FieldMappingData
  - ✅ Comprehensive test coverage for all utilities
  - ✅ All tests passing

**Design Decision:** Shared converter utilities
- Identified ~300 lines of Go-specific but target-agnostic code duplicated between GORM and Datastore
- Field classification logic was identical - both needed inline/setter/loop grouping
- Package extraction pattern repeated 3-4 times per generator
- Custom converter import collection only in GORM but should work for Datastore too
- Generic approach using interfaces and type parameters enables reuse across all Go-based targets
- Benefits: Consistent behavior, easier maintenance, foundation for future targets
- Custom converters now supported by both GORM and Datastore (previously GORM-only)

- ✅ Phase 3.1k - Comprehensive Converter Tests & Critical Bug Fix (COMPLETE)
  - ✅ Created tests/tests/gorm/ folder structure mirroring source protos
  - ✅ Wrote user_converters_test.go (27KB)
    - ✅ User: Basic fields, timestamps (Birthday), repeated primitives (Roles)
    - ✅ Author: Name and Email fields
    - ✅ Blog: Nested message conversion (Blog.Author field)
    - ✅ Product: Repeated primitives (Tags), maps (Metadata)
    - ✅ Library: Repeated messages (Contributors []Author)
    - ✅ Organization: Maps with messages (Departments map[string]Author)
    - ✅ All with round-trip verification and nil handling
  - ✅ Wrote testany_converters_test.go (10KB)
    - ✅ google.protobuf.Any field serialization/deserialization
    - ✅ Any field with different message types (Author, User, Product)
    - ✅ Enum conversion (SampleEnum with all values)
    - ✅ Timestamp field conversion (time.Time)
    - ✅ Nil Any and nil Timestamp handling
  - ✅ Wrote document_converters_test.go (18KB)
    - ✅ DocumentGormEmpty: Inherits all fields from source
    - ✅ DocumentGormPartial: Overrides specific fields
    - ✅ DocumentGormSkip: Uses skip_field to exclude fields
    - ✅ DocumentGormExtra: Adds new fields not in source
    - ✅ Field merging opt-out model verification
  - ✅ Wrote weewar_converters_test.go (21KB)
    - ✅ World: 3-level nested structures (World → WorldData → Tiles/Units)
    - ✅ WorldChange: Oneof field handling (UnitMovedChange, UnitDamagedChange)
    - ✅ Complex nested message round-trip conversion
    - ✅ Repeated nested messages with multiple levels
  - ✅ **CRITICAL BUG FIX**: Nested message converter templates
    - ✅ Bug: FromTarget converters discarded return value of nested conversions
    - ✅ Line 171 in pkg/gorm/templates/converters.go.tmpl
    - ✅ Line 191 in pkg/datastore/templates/converters.go.tmpl
    - ✅ Changed: `_, err = ConverterFunc(out.Field, ...)` → `out.Field, err = ConverterFunc(nil, ...)`
    - ✅ Impact: ALL nested message conversions now work correctly
    - ✅ Regenerated all converters with `make buf`
  - ✅ Test Results: 45+ test cases, 100% pass rate
  - ✅ Coverage: timestamps, Any fields, enums, nested messages (2-3 levels), repeated fields, maps, field merging, oneof

**Bug Impact:** Before the fix, Blog.Author and World.WorldData were always nil after round-trip conversion. The template was calling the nested converter but discarding the return value, leaving the destination field nil. This affected every FromTarget converter with nested message fields in both GORM and Datastore generators.

- ✅ **Field Ordering Fix** (COMPLETE)
  - ✅ Added TestMergeSourceFields_FieldOrdering test to verify field ordering issue
  - ✅ Fixed MergeSourceFields() to preserve source field order
  - ✅ Implementation: Track source field numbers separately in sourceFieldNumbers map
  - ✅ Sort merged fields by source field numbers (not target field numbers)
  - ✅ New target-only fields use their own field numbers for positioning
  - ✅ Added uint32, uint64, bool support to testutil GetFieldType and GetTypeName
  - ✅ Benefits: Deterministic field ordering across regenerations, fields appear in same logical order as source API
  - ✅ Example: api.User has created_at=8, updated_at=9; gorm.UserWithPermissions overrides as created_at=4, updated_at=5; generated struct has fields in positions 8 and 9
  - ✅ Implementation in pkg/generator/common/field_merge.go:86-134
  - ✅ Test coverage in pkg/generator/common/field_merge_ordering_test.go
  - ✅ All tests passing, generated code verified

**Recently Completed:**
- ✅ **DAL Helper Generation** (Phase 2.6) - COMPLETE
  - ✅ Added `generate_dal`, `dal_filename_suffix`, `dal_filename_prefix`, `dal_output_dir` options
  - ✅ Primary key detection (gorm_tags or fallback to "id")
  - ✅ Composite key support (multiple PK parameters + key struct for BatchGet)
  - ✅ Generated methods: Create, Update, Save (with WillCreate hook), Get, Delete, List, BatchGet
  - ✅ Optimistic locking support via conditional WHERE clauses on Update/Save
  - ✅ Create/Update methods without hooks (caller-level control)
  - ✅ Fixed Save() to check record existence BEFORE calling WillCreate hook
  - ✅ Added snakeCase template function for column name conversion
  - ✅ Shared utilities: GetColumnOptions(), GetColumnName(), ToSnakeCase()
  - ✅ Skips messages without primary keys gracefully
  - ✅ Optional subdirectory output (`dal_output_dir`)
  - ✅ Comprehensive unit tests including optimistic locking scenarios
  - ✅ Templates for single and composite keys
  - ✅ Hook-based lifecycle customization
  - ✅ SQLite integration tests verifying AutoMigrate and CRUD operations

- ✅ **Cross-DB Compatibility** (Phase 2.7) - COMPLETE
  - ✅ Updated proto files with `serializer:json` tags for cross-DB compatibility
  - ✅ Added validation warnings for complex types missing serializer tags
  - ✅ Validation in pkg/gorm/generator.go (validateSerializerTags function)
  - ✅ Skips warnings for embedded fields and types with implement_scanner
  - ✅ MessageRegistry.LookupTargetMessage() resolves source to GORM target types
  - ✅ Smart detection handles repeated fields, maps, and nested messages
  - ✅ Clear warning format: `[WARN] Field 'X.Y' (type): missing serializer:json tag`

- ✅ **Optional Valuer/Scanner Generation** (Phase 2.8) - COMPLETE
  - ✅ Added `implement_scanner` boolean option to GormOptions annotation
  - ✅ Opt-in driver.Valuer and sql.Scanner implementation using encoding/json
  - ✅ Template scanner_valuer.go.tmpl conditionally included
  - ✅ Value() marshals struct to JSON
  - ✅ Scan() unmarshals from JSON or string with nil handling
  - ✅ Validation smart suppression when target type has implement_scanner
  - ✅ Registry lookup ensures warnings respect Valuer/Scanner implementations
  - ✅ Works for both repeated fields and map value types
  - ✅ Benefits: Cleaner protos, automatic JSON serialization, no tag duplication

- ✅ **Late-Binding Table/Kind Support** (Phase 2.9) - COMPLETE
  - ✅ Added `optional bool dal` field to GormOptions and DatastoreOptions
  - ✅ DAL generation now conditional based on GenerateDAL flag
  - ✅ DAL struct has TableName field for runtime table override
  - ✅ Constructor pattern: `NewUserGORMDAL("custom_table")`
  - ✅ db() helper returns `db.Table(tableName)` when set, otherwise unchanged
  - ✅ Datastore Kind() method now conditional (only when kind specified)
  - ✅ Full backward compatibility - existing protos work unchanged
  - ✅ Tests: TestDALLateBinding, TestDALTableNameEmpty verify late-binding behavior
  - ✅ Use case: Same Address struct in user_addresses and company_addresses tables

**Next:**
1. **Phase 3.2**: postgres-raw (Go + database/sql)
2. **Phase 3.3**: firestore (Go)
3. **Phase 3.4**: mongodb (Go)
4. **Phase 4**: Multi-language support (Python, TypeScript)
5. **Phase 5**: Advanced features (if needed after real-world usage)
6. **Phase 6**: Service layer generation (optional - deferred based on DAL helper usage)

## Notes

- Each TODO item follows TDD: write test first, then implement
- Mark items with ✅ when test passes
- Refactor when needed before moving to next item
- Keep tests focused and atomic
- Integration tests at binary level, unit tests at package level
