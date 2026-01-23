# protoc-gen-dal: Design Summary

## Project Vision

Generate Data Access Layer (DAL) converters that transform between clean API protobuf messages and various database/datastore representations.

## Core Principles

### 1. Targets: Database Access Layers

**Target** = The data access technology you're using

There are three types of targets:

**1. Direct Database Targets** (database-specific)
- `postgres-raw`, `mysql-raw` - Direct SQL for specific databases
- Language can vary: Go + database/sql, Python + psycopg2, TypeScript + pg
- Tied to one database type

**2. ORM Targets** (database-agnostic at codegen time)
- `gorm` - Database-agnostic ORM (supports postgres, mysql, sqlite, etc.)
- Database chosen at runtime via GORM dialects
- Allows database-specific type hints (e.g., `type: "jsonb"` for PostgreSQL)
- Generated code is Go only (GORM is Go-specific)

**3. Datastore Targets** (NoSQL/document stores)
- `firestore`, `mongodb`, `dynamodb` - NoSQL datastores
- Language can vary per target

### 2. Sidecar Pattern (API Purity)

**Keep API protos clean** - No database concerns in API definitions!

```
protos/
├── api/
│   └── library/v1/
│       └── book.proto          # Clean API - no DB pollution!
│
└── dal/
    └── library/v1/
        ├── book_gorm.proto          # GORM schema (multi-DB)
        ├── book_postgres.proto      # PostgreSQL raw SQL schema
        ├── book_firestore.proto     # Firestore schema
        └── book_mongodb.proto       # MongoDB schema
```

**Benefits:**
- API protos remain database-agnostic
- One API message can map to multiple datastores
- DB-only fields (created_at, updated_at) don't leak into API

### 3. Opt-In Tables, Opt-Out Fields

**Messages:**
- WITHOUT target annotation (e.g., `postgres`) → **SKIP**
- WITH target annotation → **INCLUDE**
- No need for `skip_dal` at message level

**Fields:**
- **Include by default** (once message is opted-in)
- Use `skip_field` to exclude specific fields

### 4. Type Safety Over Runtime Lookups

**Generated code must be compile-time safe:**
- ✅ Typed structs: `BookGORM`, `BookFirestore`
- ✅ Typed decorators: `func(*Book, *BookGORM) error`
- ❌ NO `map[string]any` dictionaries
- ❌ NO string-based transformer registries
- ❌ NO runtime map lookups

### 5. Hooks as Function Pointers

**Lifecycle hooks without interface boilerplate:**

```go
type BookRepository struct {
    db *gorm.DB

    // Optional function pointers - only set what you need!
    OnBeforeCreate func(context.Context, *Book) error
    OnAfterCreate  func(context.Context, *Book) error
    // ... etc
}
```

**Benefits:**
- No need to implement empty interface methods
- Set only the hooks you care about
- Type-safe, compile-time checked

### 6. Decorator Pattern for Transformations

**Start simple with decorators** - no declarative transformers yet:

```go
// Fully typed - IDE autocomplete works!
bookGorm, err := BookToGORM(apiBook, func(api *Book, gorm *BookGORM) error {
    // Split composite ID
    parts := strings.Split(api.Id, ":")
    gorm.Type = parts[0]
    gorm.DBID = parts[1]

    // Custom fields
    gorm.SearchText = api.Title + " " + api.Author
    return nil
})
```

**Future:** Add typed function pointer transformers when needed (compile-time checked, no runtime lookups)

## Architecture

### No Intermediate Representation (IR)

**Decision:** Skip IR layer - directly process proto annotations

**Rationale:**
- IR would just duplicate proto descriptor information
- Adds complexity without clear benefit
- Direct annotation reading is simpler and sufficient

### Two-Phase Processing

```
Phase 1: Collection (Grouper)
├─ Scan all proto files
├─ Find messages with target annotation
├─ Build message index for cross-referencing
└─ Group by target

Phase 2: Generation
├─ Receives ALL messages for target
├─ Analyze relationships, foreign keys across files
├─ Generate code with templates
└─ Return generated files
```

**Why Phase 1 is critical:**
- Foreign keys may reference messages in different files
- Need to resolve relationships across entire proto set
- One-to-many, many-to-many requires seeing all related messages

### One Binary Per Target

```
cmd/
├── protoc-gen-dal-gorm/              # GORM (multi-DB ORM)
├── protoc-gen-dal-postgres-raw/      # PostgreSQL raw SQL (Go)
├── protoc-gen-dal-firestore/         # Firestore (Go)
└── protoc-gen-python-dal-postgres/   # PostgreSQL raw SQL (Python)
```

**Each main.go is thin:**
1. Collect messages for target (using shared collector)
2. Delegate to generator
3. Done

**Benefits:**
- Focused single responsibility
- Easy to add new targets
- Independent versioning possible
- Clear separation of concerns

### Package Structure

```
protoc-gen-dal/                    # Monorepo
├── cmd/
│   ├── protoc-gen-dal-gorm/
│   │   └── main.go                   # Thin: collect → delegate
│   ├── protoc-gen-dal-postgres-raw/
│   └── protoc-gen-dal-firestore/
│
├── pkg/
│   ├── collector/
│   │   └── collector.go              # Shared: collect messages by target
│   │
│   ├── gorm/
│   │   ├── generator.go              # GORM generator (multi-DB)
│   │   └── templates/
│   │       └── go/
│   │
│   ├── postgres/
│   │   ├── raw_go.go                 # Generator for Go+raw SQL
│   │   ├── raw_python.go             # Generator for Python+psycopg2
│   │   └── templates/
│   │       ├── go/
│   │       └── python/
│   │
│   ├── firestore/
│   │   ├── generator_go.go
│   │   └── templates/
│   │
│   └── mongodb/
│       └── generator_go.go
│
└── protos/
    └── dal/v1/
        └── annotations.proto         # Shared annotations
```

## Annotations Design

### Target-Focused Annotations

```protobuf
// GORM target (database-agnostic ORM)
message GormOptions {
  string source = 1;            // "library.v1.Book" - links to API message
  string table = 2;             // "books"
  repeated string embedded = 3; // Embedded field names
}

// PostgreSQL target (raw SQL)
message PostgresOptions {
  string source = 1;        // "library.v1.Book" - links to API message
  string table = 2;         // "books"
  string schema = 3;        // "public"
}

// Firestore target
message FirestoreOptions {
  string source = 1;        // "library.v1.Book" - links to API message
  string collection = 2;    // "books"
}
```

### Field-Level Annotations

**Philosophy: No abstraction layer - use native target syntax directly**

```protobuf
message ColumnOptions {
  string name = 1;  // Column name override (optional)

  // Target-specific tags (use what you already know!)
  repeated string gorm_tags = 10;      // GORM: ["primaryKey", "type:uuid"]
  repeated string sql_tags = 11;       // Raw SQL
  repeated string firestore_tags = 12; // Firestore
  repeated string mongodb_tags = 13;   // MongoDB
}

message ForeignKeyOptions {
  string references = 1;        // "authors.id"
  ReferentialAction on_delete = 2;
  ReferentialAction on_update = 3;
}
```

**Example:**
```protobuf
int64 created_at = 5 [(dal.v1.column) = {
  gorm_tags: ["autoCreateTime", "index"]
}];
// Generates: `gorm:"autoCreateTime;index"`
```

## Generated Code Pattern

### GORM Example

**Input:**
```protobuf
message BookGorm {
  option (dal.v1.gorm) = {
    source: "library.v1.Book"
    table: "books"
  };

  // Use GORM tags directly - no abstraction!
  string id = 1 [(dal.v1.column) = {
    gorm_tags: ["primaryKey", "type:uuid"]
  }];

  string title = 2 [(dal.v1.column) = {
    gorm_tags: ["type:varchar(255)", "not null"]
  }];

  int64 published_at = 3;

  repeated string tags = 4 [(dal.v1.column) = {
    gorm_tags: ["type:jsonb"]  // PostgreSQL-specific - works with postgres dialect
  }];

  // Timestamp fields with GORM's auto-time tags
  int64 created_at = 5 [(dal.v1.column) = {
    gorm_tags: ["autoCreateTime", "index"]
  }];

  int64 updated_at = 6 [(dal.v1.column) = {
    gorm_tags: ["autoUpdateTime"]
  }];
}
```

**Output:**
```go
// Generated GORM model
type BookGORM struct {
    ID          string   `gorm:"primaryKey;type:uuid"`
    Title       string   `gorm:"type:varchar(255);not null"`
    PublishedAt int64    `gorm:"column:published_at"`
    Tags        []string `gorm:"type:jsonb"`
    CreatedAt   int64    `gorm:"autoCreateTime;index"`
    UpdatedAt   int64    `gorm:"autoUpdateTime"`
}

func (BookGORM) TableName() string { return "books" }

// Typed converter with optional decorator
func BookToGORM(api *libraryv1.Book,
    decorator func(*libraryv1.Book, *BookGORM) error) (*BookGORM, error)

// Reverse converter
func BookFromGORM(gorm *BookGORM,
    decorator func(*BookGORM, *libraryv1.Book) error) (*libraryv1.Book, error)

// Optional: Repository with hooks
type BookRepository struct {
    db *gorm.DB

    OnBeforeCreate func(context.Context, *libraryv1.Book) error
    OnAfterCreate  func(context.Context, *libraryv1.Book) error
}
```

## Complex Type Handling

### Storage Strategies for Repeated/Map Fields

```protobuf
message Book {
  // JSON storage (default for complex types)
  repeated string tags = 1 [(dal.v1.column) = {
    type: "JSONB"
  }];

  // Native array (PostgreSQL)
  repeated string categories = 2 [(dal.v1.column) = {
    type: "TEXT[]"
  }];

  // Map as JSON
  map<string, string> metadata = 3 [(dal.v1.column) = {
    type: "JSONB"
  }];
}
```

### Composite Keys

**Option 1: Multiple Primary Keys (Pure SQL)**
```protobuf
message BookEdition {
  string book_id = 1 [(dal.v1.column) = {primary_key: true}];
  int32 edition_number = 2 [(dal.v1.column) = {primary_key: true}];
}
```

**Option 2: Virtual Field with Decorator**
```go
// User provides decorator to split/combine
bookGorm, _ := BookToGORM(apiBook, func(api *Book, gorm *BookGORM) error {
    parts := strings.Split(api.Id, ":")
    gorm.Type = parts[0]
    gorm.DBID = parts[1]
    return nil
})
```

## Relationship Handling

### Foreign Keys

```protobuf
message Book {
  string author_id = 3 [(dal.v1.foreign_key) = {
    references: "authors.id"
    on_delete: CASCADE
  }];
}
```

**Generator:**
- Parses "authors.id"
- Finds `Author` message by table name in collected messages
- Generates GORM tag: `foreignKey:AuthorID;references:ID`

### Future: Lazy Loading

```protobuf
message Book {
  repeated Author authors = 4 [(dal.v1.relationship) = {
    type: ONE_TO_MANY
    foreign_key: "book_id"
    lazy_load: true
  }];
}
```

Generates `LoadAuthors(ctx, book)` method for on-demand loading.

## Migration Strategy

**Philosophy:** Don't auto-execute DDL - too dangerous!

**GORM:** User calls `db.AutoMigrate(&BookGORM{})`

**Raw SQL:** Optionally generate DDL constants
```go
const BookTableDDL = `CREATE TABLE IF NOT EXISTS books (...)`
```

**Best practice:** Use migration tools (flyway, migrate, goose) for production

## Development Workflow

### Test-Driven Development (TDD)

1. Write one breaking test
2. Run test (should fail)
3. Write minimal code to make it pass
4. Refactor if needed
5. Repeat

**Benefits:**
- Prevents over-engineering
- Ensures code does what's needed
- Built-in regression protection
- Forces thinking about API before implementation

## Open Questions / Future Work

1. **Relationship annotations** - Full spec for one-to-many, many-to-many
2. **Declarative transformers** - Typed function pointers (not runtime lookups)
3. **Nested messages** - How to handle message-type fields
4. **Indexes** - Generate index creation code
5. **Validation** - Field validation annotations
6. **Caching** - Repository-level caching hooks
7. **Transactions** - Transaction support in repositories
8. **Multi-language** - Python, TypeScript generators

## Design Decisions Log

| Decision | Rationale |
|----------|-----------|
| Sidecar pattern | Keep API protos clean, allow multiple DB representations |
| No IR layer | Direct proto reading simpler, no duplication |
| One binary per target | Focused, easy to add new combinations |
| Grouper phase | Need all messages to resolve cross-file relationships |
| Type safety | Compile-time checks, no runtime map lookups |
| Function pointer hooks | Avoid interface boilerplate |
| Decorator pattern | Simple, type-safe transformations |
| Target-focused annotations | Database is the hero, not the language |
| Opt-in messages | Explicit about which messages are DB entities |
| TDD workflow | Prevent over-engineering, ensure correctness |
| GORM as target (not vehicle) | GORM is database-agnostic; database chosen at runtime via dialects. Same generated code works for postgres/mysql/sqlite. Users specify DB-specific types if needed (e.g., `type: "jsonb"`). Simpler mental model than postgres-gorm, mysql-gorm, etc. |
| No abstraction layer for tags | Use native target syntax directly (e.g., `gorm_tags: ["primaryKey", "autoCreateTime"]`). Don't create another language on top of what users already know. Pass through tags verbatim. Reduces learning curve and maintains full feature access. |
| One file per proto file | Generate one Go file per proto file (not per message). All GORM messages in `gorm/user.proto` → `user_gorm.go`. Embedded types automatically included in same file. Users control organization via proto file structure. Simpler mental model and better for Go compilation. |
| Built-in type converters | Provide smart defaults for common conversions (`*timestamppb.Timestamp` ↔ `time.Time`, numeric types). Custom converters via annotations (`to_func`/`from_func`) override defaults. Nested message converters auto-generated when available. Decorator function handles edge cases. Priority: custom > built-in > skip (decorator). |
| Converter naming convention | Use `{Source}To{Target}` and `{Source}From{Target}` patterns (e.g., `UserToUserGORM`, `UserFromUserGORM`). Prevents naming collisions when source and target have same name. Clear directionality. |
| Converter registry for nested types | Track available converters in a registry. Generate nested converter calls only when converter exists. Skip fields requiring manual handling. Enables recursive conversions without manual wiring. Requires explicit `source` annotation - no inference for clarity. |
| Conversion type system | Categorize conversions into distinct types: `ConvertIgnore` (skip), `ConvertByAssignment` (direct copy), `ConvertByTransformer` (no error), `ConvertByTransformerWithError` (nested messages), `ConvertByTransformerWithIgnorableError` (optional errors). Clear semantics for each conversion pattern. |
| Source/Target naming convention | Use "Source" and "Target" throughout codebase instead of "Source" and "Gorm". Makes code reusable across all targets (postgres, firestore, mongodb). Only model names use target-specific suffixes (e.g., `UserGORM`). |
| In-place conversion with dest parameter | Converters accept both `src` and `dest` parameters, allowing in-place modification and avoiding unnecessary allocations. Enables pointer vs value type flexibility: `converter(src, &out.Field, nil)` for value fields, `converter(src, out.Field, nil)` for pointer fields. |
| Optional keyword determines GORM pointers | Proto message fields without `optional` keyword generate as value types in GORM structs. Only fields with explicit `optional` keyword become pointers. Matches user intent: embedded types are values, nullable fields are pointers. Uses `HasOptionalKeyword()` not `HasPresence()`. |
| Template helpers for field access | `fieldRef` template helper generates correct field access expressions (`&obj.field` for values, `obj.field` for pointers). Eliminates error-prone manual pointer logic in templates. Single source of truth for address-of semantics. |
| Explicit converter warnings | When matching message fields lack converters, emit clear warnings to stderr with actionable guidance. No silent failures or type inference. User must add `source` annotation for automatic conversion or handle in decorator. |
| Applicative conversion for collections | Maps and repeated fields apply conversion logic to their contained types recursively. `map<K, primitive>` and `[]primitive` use direct assignment (`ConvertByAssignment`). `[]MessageType` uses loop-based conversion with element converter. `map<K, MessageType>` uses loop-based conversion with value converter. Check contained type, not container, to determine conversion strategy. Early return for primitive containers avoids incorrect message converter lookup. |
| Loop-based conversion for repeated messages | `[]MessageType` fields generate loop code that applies element converter to each item. Uses `IsRepeated` flag and `TargetElementType`/`SourceElementType` fields in `FieldMappingData`. Template generates: `make([]ElementType, len(src.Field))` then `for i, item := range src.Field { converter(item, &out.Field[i], nil) }`. Proper error handling with index in error message. Reuses existing `ConversionType` enum - no new types needed. |
| Loop-based conversion for map message values | `map<K, MessageType>` fields generate loop code that applies value converter to each entry. Uses `IsMap` flag and `TargetElementType`/`SourceElementType` fields in `FieldMappingData`. Template generates: `make(map[K]ValueType, len(src.Field))` then `for key, value := range src.Field { var converted ValueType; converter(value, &converted, nil); out.Field[key] = converted }`. Proper error handling with key in error message. Fixed `protoToGoType` to use `buildStructName` for message value types instead of `protoScalarToGo` which returned `interface{}`. |
| Foreign keys via native tags | Foreign keys don't need special annotations - use native GORM tags directly: `gorm_tags: ["foreignKey:AuthorID", "references:ID", "constraint:OnDelete:CASCADE"]`. Violates our "no abstraction layer" principle to create a special `foreign_key` annotation when GORM already has perfectly good syntax. Users know GORM - just let them use it. Same applies to all database-specific features. |
| Composite keys via native tags | Composite keys don't need special implementation - multiple fields with `gorm_tags: ["primaryKey"]` already work. GORM handles composite primary keys automatically. No need for special primary_key annotation or composite key detection logic. Keep it simple - trust the target ORM/database to do its job. |
| Phase 2 simplification | Phase 2.1 (Field Transformations) already complete via built-in converters and decorators. Phase 2.3 (Foreign Keys) and 2.4 (Composite Keys) skipped - native tags handle everything. Phase 2.2 (Collections) required actual code generation for loops. Key insight: Only implement features that require code generation logic; everything else is just tag pass-through. |
| Datastore target implementation | Google Cloud Datastore added as first Phase 3 target. Mirrors GORM generator pattern: collector extracts DatastoreOptions (kind, namespace, source), generator creates entity structs with datastore tags, templates generate Kind() method. Key field added automatically (excluded via `datastore:"-"`) for manual key management. LoadKey/SaveKey PropertyLoadSaver methods deferred - not essential for MVP, users can manually manage keys which is simpler and more explicit. Will add if requested. |
| Datastore key management | Skip PropertyLoadSaver interface (LoadKey/SaveKey) for MVP. Not required for basic Datastore usage - users simply pass keys to Get/Put operations and can manually set Key field if needed. PropertyLoadSaver only needed for advanced custom property transformations. Keeps generated code simple. Can be added later as enhancement if users request automatic key population. Follows "implement only what requires code generation" principle. |
| Datastore converter generation | Reuses GORM converter infrastructure with Datastore-specific type conversions. Generates bidirectional converters (XToY, XFromY) with decorator support. Built-in type conversions: uint32↔string (IDs stored as strings in Datastore keys), Timestamp↔time.Time (native support). Helper functions generated in converter file: timestampToTime, timeToTimestamp, mustParseUint. Follows same pattern as GORM converters for consistency. Full support for scalar fields, repeated fields (direct assignment for primitives, loop-based for messages), and map fields (direct assignment for primitive values, loop-based for message values). Converter registry tracks available nested converters for automatic loop generation. |
| Shared generator utilities | Extract common code from GORM and Datastore generators into reusable packages: `pkg/generator/common` (file naming, package extraction, type mapping), `pkg/generator/registry` (converter tracking). Prevents "relearning" mistakes like map field handling. Centralizes proto→Go type mapping in `ProtoFieldToGoType()` which correctly generates `map[K]V` types instead of entry structs. ~500 lines of duplicate code consolidated. All utilities have unit tests. Future targets (postgres-raw, firestore, mongodb) can immediately leverage these utilities. |
| Datastore repeated/map fields | Mirrors GORM approach: repeated scalars ([]string, []int32) use direct assignment. Maps with scalar values (map[string]string) use direct assignment with nil checks. Repeated message types ([]Author) generate loop-based conversion calling element converter. Maps with message values (map[string]Author) generate loop-based conversion calling value converter. Proper bidirectional handling: API uses pointers ([]*api.Author), Datastore uses values ([]AuthorDatastore). Template uses $converter variable to access parent context for package name in nested conversions. |
| Converter strategy utilities | Separate user intent (ConversionType: what conversion to apply) from implementation detail (FieldRenderStrategy: how to render it). ConversionType determined by generator based on proto annotations, field types, and built-in rules. FieldRenderStrategy derived from ConversionType + field characteristics (pointer, repeated, map). Strategy types: StrategyInlineValue (struct literal), StrategySetterSimple/Transform/WithError/IgnoreError (post-construction statements), StrategyLoopRepeated/Map (loop blocks). Utilities in `pkg/generator/converter`: collection helpers (CheckMapValueType, CheckRepeatedElementType), conversion detection (IsTimestampToInt64, IsNumericConversion), template helper generators (TimestampHelperFunctions, MustParseUintHelper). Enables template simplification: templates render based on strategy, not conditional logic scattered throughout. |
| Struct literal + setters pattern | Unified converter template structure across GORM and Datastore generators. Three-section pattern: (1) Struct literal initialization for inline fields (non-pointer, no errors, direct assignments or simple transformations), (2) Post-construction setter statements for pointer fields and error handling, (3) Separate loop-based conversion sections for repeated/map message types. Fields classified into groups (ToTargetInlineFields, ToTargetSetterFields, ToTargetLoopFields) by render strategy. Template helpers (isInlineValue, isSetterSimple, etc.) make templates readable. Empty ToTargetCode/FromTargetCode handled via fallback: `{{ if .ToTargetCode }}{{ .ToTargetCode }}{{ else }}src.{{ .SourceField }}{{ end }}`. Benefits: cleaner generated code (single struct initialization vs scattered assignments), consistent pattern across all targets, easier to extend for future targets. |
| Timestamp as time.Time | google.protobuf.Timestamp fields automatically map to time.Time (not int64) in generated database structs. Both GORM and Datastore natively support time.Time, making it more idiomatic and type-safe than Unix timestamps. Converter helpers (timestampToTime, timeToTimestamp) use ts.AsTime() and timestamppb.New(t) with proper zero-value handling (time.Time{} and t.IsZero()). Proto files still use google.protobuf.Timestamp - the mapping happens during code generation via ProtoFieldToGoType(). |
| Shared converters package | All converter helper functions centralized in pkg/converters instead of being duplicated in every generated file. Generated code imports "github.com/panyam/protoc-gen-dal/pkg/converters" and calls converters.TimestampToTime(), converters.TimeToTimestamp(), converters.MustParseUint(). Cleaner generated code, easier maintenance, single source of truth for conversion logic. Users can also import and use these converters in their own code. |
| Well-known type registry | Two-level type system: (1) Well-known type registry in pkg/generator/common/types.go maps proto types to idiomatic Go types (google.protobuf.Timestamp→time.Time, google.protobuf.Any→[]byte). (2) Type conversion registry in pkg/generator/converter/type_mappings.go maps between source (proto-generated) and target (our generated) types with converter functions. Distinction crucial: source uses proto-generated types (*anypb.Any), target uses our mappings ([]byte). getSourceTypeKey() returns proto names, getTargetTypeKey() applies well-known mappings. RegisterWellKnownType() allows easy extension. Benefits: Consistent handling across all well-known types, database-friendly representations, extensible for custom types. |
| Type conversion registry | Centralized registry in pkg/generator/converter/type_mappings.go defines how to convert between type pairs. Each mapping specifies ToTargetTemplate and FromTargetTemplate (Go expression templates), ConversionType (error handling semantics), and optional TargetIsPointer override. Current mappings: google.protobuf.Timestamp→int64 (Unix epoch), google.protobuf.Timestamp→google.protobuf.Timestamp (actually time.Time in Go), google.protobuf.Any→bytes (binary serialization), uint32→string (Datastore key IDs). Templates use {{.SourceField}}/{{.TargetField}} placeholders. Converter functions in pkg/converters: AnyToBytes/BytesToAny (proto.Marshal/Unmarshal), TimestampToInt64/Int64ToTimestamp, TimestampToTime/TimeToTimestamp. Registry enables automatic conversion for complex types without manual code. |
| Template support for inline converters | Templates handle both converter functions (ToTargetConverterFunc for message conversions) and inline converter code (ToTargetCode from type mapping registry). For ConvertByTransformerWithError strategy: if ToTargetConverterFunc exists, use it (e.g., AuthorToAuthorGORM); if ToTargetCode exists, use inline conversion with error handling (e.g., out.Field, err = converters.AnyToBytes(src.Field)). Applied to both GORM and Datastore templates in ToTarget and FromTarget directions. Fixes issue where type mapping registry conversions were generating invalid tuple assignments. Enables seamless integration of type conversion registry with existing message converter infrastructure. |
| Enum type support | Proto enum fields map to their Go enum type (int32 alias) in generated structs. Detected via field.Enum != nil in ProtoFieldToGoType(). Enum types package-qualified when from different package (e.g., api.SampleEnum). sourcePkgName parameter added to ProtoFieldToGoType() for qualification. Both GORM and Datastore generators collect source package imports for enum types. Template Imports changed from []string to []ImportSpec to support aliased imports ({{ if .Alias }}{{ .Alias }} {{ end }}"{{ .Path }}"). Package aliases use GetPackageAlias() (last path component) for consistency with converter generation. Enums converted as simple value types via direct assignment in converters - no special conversion logic needed since enums are just int32 aliases in Go. |
| Message field type resolution | MessageRegistry in pkg/generator/common/message_registry.go provides source→target message type lookups based on the `source` annotation (not name matching). When BlogAsIsGORM (with `source: "api.Blog"`) inherits an "author" field via field merging, registry maps `api.Author→AuthorGORM`, ensuring correct type usage. ProtoFieldToGoType() uses registry parameter to resolve message field types. Flexible naming: users can name targets anything (MyFancyAuthor works as long as `source: "api.Author"` is set). Validation enforces explicit definitions: ValidateMissingTypes() scans all merged fields, reports error if any referenced message type lacks a target definition. No auto-generation - users must define all needed types. Single-pass algorithm: collect all referenced source message types, check registry, skip well-known types and existing target types, error on missing. Shared across GORM and Datastore generators. Example: `BlogAsIsGORM.Author` becomes `AuthorGORM` (not raw `Author` from api package), even though BlogAsIsGORM has no explicit field definitions - the type comes from registry lookup of `api.Author`. |
| Source message validation | Collector validates that all source message references actually exist, failing fast with clear errors. Previously, missing source messages were silently skipped (returned nil), leading to confusion. Now CollectMessages() returns error if any source is missing. All extract functions (extractGormInfo, extractPostgresInfo, etc.) validate source exists and return descriptive errors: "message 'gorm.BookGORM' references source 'api.Book' which does not exist. Please ensure the source proto file is imported and the message name is correct". Errors collected and reported together. Benefits: catches typos immediately, forces explicit decisions, better developer experience. Applied across all targets (GORM, Postgres, Firestore, MongoDB, Datastore). |
| Oneof-aware field merging | Oneof declarations create individual fields (not a single field with the oneof name). Example: `oneof move_type { MoveUnitAction move_unit = 4; }` creates fields named "move_unit", "attack_unit", etc. (not "move_type"). Smart detection: if target has a field matching the oneof NAME, automatically skip ALL oneof members. Zero boilerplate: `message GameMoveGORM { google.protobuf.Any move_type = 1; }` auto-skips move_unit, attack_unit, end_turn, build_unit from source. MergeSourceFields() checks target field names against source oneof names; when matched, skips all members of that oneof. Three approaches: (1) **Auto-skip via name matching** (recommended) - name target field after oneof. (2) Define target mappings for each oneof member type for proper conversion. (3) Manual skip via `skip_field` annotation on individual members. Implementation in pkg/generator/common/field_merge.go lines 55-81. Tested with GameMoveGORM in weewar.proto - cleanest syntax, zero manual field declarations. |
| MessageRegistry for converter lookup | Nested message field converters now resolve correctly via MessageRegistry instead of direct message type lookups. Bug fix: Previously both sourceMsg and targetMsg pointed to the source proto (e.g., api.IndexInfo), causing converter lookup to fail ("IndexInfo:IndexInfo" not found). Solution: Use msgRegistry.LookupTargetMessage(sourceMsg) to resolve api.IndexInfo→IndexInfoGORM, then check "IndexInfo:IndexInfoGORM" in converter registry. Applied to all applicatives: regular nested messages, repeated message fields ([]Tile), and map message values (map<int32, Player>). Eliminated 90% of false "missing converter" warnings (20→2 warnings in weewar.proto). Changes in pkg/gorm/generator.go: Updated buildFieldConversion() signature to accept msgRegistry parameter, changed lines 691-712 to call msgRegistry.LookupTargetMessage() instead of using targetField.Message directly, propagated msgRegistry through GenerateConverters()→generateConverterFileCode()→buildConverterData()→buildFieldConversion() call chain. Test coverage in pkg/gorm/converter_lookup_test.go with programmatic proto descriptor creation. |
| Automatic google.protobuf.Any serialization | Fields of type google.protobuf.Any are automatically serialized to []byte in the database using protobuf marshaling. Single Any fields: Direct conversion via converters.MessageToAnyBytes(msg) and converters.AnyBytesToMessage[T](bytes). Repeated Any fields: Loop-based conversion using wrapper functions converters.MessageToAnyBytesConverter and converters.AnyBytesToMessageConverter that match the standard converter signature (dest, src, decorator). Implementation: pkg/converters/any.go provides MessageToAnyBytes (packs message into Any via anypb.New, then marshals to bytes), AnyBytesToMessage (unmarshals bytes to Any, then unpacks via any.UnmarshalNew), and standard-signature wrappers for use in templates. Detection in pkg/gorm/generator.go lines 727-757: checks if target field is google.protobuf.Any via GetWellKnownTypeMapping, extracts source package and type name, branches on IsRepeated flag to set either direct ToTargetCode/FromTargetCode (single fields) or element types with converter functions (repeated fields). Zero configuration required - works automatically for any proto message type stored in Any fields. Eliminates all "missing converter" warnings for Any fields (reduced from 20→0 warnings). Example: GameMoveGORM.move_type (google.protobuf.Any) stores various move actions, GameMoveGORM.changes (repeated google.protobuf.Any) stores multiple world changes, both convert transparently between typed messages and []byte storage. |
| Shared test utilities | Test helper functions extracted into pkg/generator/testutil to eliminate duplication between GORM and Datastore test files. Provides TestProtoSet, TestFile, TestMessage, TestField structures for programmatic proto descriptor creation. CreateTestPlugin() builds complete protogen.Plugin from test data. BuildCodeGeneratorRequest() and BuildFileDescriptor() create FileDescriptorProto with proper package info, field types, and DAL annotations. TestMessage supports both GormOpts and DatastoreOpts for target-agnostic test data. Reduced ~400 lines of duplicate code across gorm/generator_test.go and datastore/generator_test.go. Benefits: Consistent test patterns across all generators, easier to add tests for new targets, centralized proto descriptor building logic, reduced maintenance burden when proto structures change. |
| Shared converter utilities | Field classification and converter utilities extracted to eliminate Go-specific code duplication between GORM and Datastore generators. pkg/generator/converter/classification.go provides ClassifyFields() generic function using FieldWithStrategy interface, groups field mappings by render strategy (inline/setter/loop) for both ToTarget and FromTarget directions. AddRenderStrategies() calculates render strategies from conversion types and field characteristics without generator-specific logic. pkg/generator/common/package_info.go provides ExtractPackageInfo() to cleanly extract import path and alias from proto messages, handling ;packagename suffix stripping. pkg/generator/common/custom_converters.go provides CollectCustomConverterImports() to scan field annotations for to_func/from_func and add imports, plus ExtractCustomConverters() to generate converter code from column options. All utilities fully tested. Benefits: Eliminates ~300 lines of duplication, ensures consistent behavior between GORM and Datastore, easier to add new targets (postgres-raw, firestore, mongodb), centralized logic for Go-specific but target-agnostic operations. |
| Unified FieldMapping struct | Consolidated duplicate FieldMapping structures from GORM (FieldMappingData) and Datastore (FieldMapping) into single shared converter.FieldMapping struct in pkg/generator/converter/field_mapping.go. Both generators now use []*converter.FieldMapping for field mappings, eliminating ~85 lines of duplicate struct definitions. Helper functions (BuildMapFieldMapping, BuildRepeatedFieldMapping, BuildMessageToMessageMapping, etc.) modified to accept mapping parameter for in-place modification instead of creating intermediate types. Benefits: Single source of truth for field mapping logic, simplified ClassifyFields (no slice-to-pointer conversion needed), unified pointer handling semantics (HasOptionalKeyword for target fields), easier maintenance, consistent behavior across all targets. Eliminated intermediate FieldMappingResult type that added unnecessary complexity. |
| Consolidated buildFieldMapping functions | Created shared converter.BuildFieldMapping() function that both GORM and Datastore use instead of maintaining separate 130-line implementations. Target-specific generators provide addRenderStrategies callback for customization. Eliminated ~250 lines of duplicate field mapping logic. Both buildFieldConversion (GORM) and buildFieldMapping (Datastore) are now thin 2-line wrappers calling the shared function. Refactored Datastore to use same shared helpers (BuildMapFieldMapping, BuildRepeatedFieldMapping, etc.) that GORM uses, removing 200+ lines of inline map/repeated field handling. Benefits: Single source of truth for field conversion logic, consistent behavior across targets, easier to add new generator targets, simplified maintenance, reduced testing burden. |
| Shared codegen types | Created pkg/generator/types/codegen.go with shared types used across all generators: GeneratedFile (file path + content), GenerateResult (list of generated files), ConverterData (converter function metadata). Eliminated duplicate type definitions from GORM and Datastore generators (~50 lines). Both generators now import from types package. Benefits: Consistent type definitions, easier to extend for new targets, single source of truth for codegen data structures. |
| Field ordering based on source proto | Generated struct fields maintain source proto field order, not target proto field order. When target proto overrides source fields with different field numbers, the generated struct preserves the source field's position for consistency. Implementation: MergeSourceFields() tracks source field numbers separately (sourceFieldNumbers map) and uses them for sorting instead of the merged field's numbers. New target-only fields use their own field numbers for positioning. Benefits: Deterministic field ordering across regenerations, fields appear in the same logical order as the source API definition, easier to understand the mapping between API and database representations. Implementation in pkg/generator/common/field_merge.go:86-134. Test coverage in pkg/generator/common/field_merge_ordering_test.go. Example: api.User has created_at=8, updated_at=9; gorm.UserWithPermissions overrides them as created_at=4, updated_at=5; generated struct has fields in positions 8 and 9 (not 4 and 5) to maintain source order. |
| DAL helper generation | Generated DAL helpers provide basic CRUD operations (Create, Update, Save, Get, Delete, List, BatchGet) for each GORM entity to eliminate service-layer boilerplate. Create() creates new records (fails if exists). Update() updates existing records with optimistic locking support via WHERE conditions (fails if not found). Save() implements upsert pattern with WillCreate hook for timestamp handling - checks if record exists before saving, calls hook only on create. Get/Delete accept primary key parameters (single or composite). List accepts pre-built GORM query for flexibility. BatchGet handles bulk fetches efficiently. Configuration: `generate_dal=true`, `dal_filename_suffix="_dal"`, `dal_output_dir="dal"` (optional subdirectory). Primary key detection: auto-detects `gorm_tags: ["primaryKey"]`, falls back to "id" field, skips messages without PKs gracefully. Composite keys generate separate parameters and key struct for BatchGet. Hook pattern: DAL struct has optional `WillCreate func(context.Context, *Entity) error` for lifecycle customization (e.g., setting timestamps). Hooks NOT in Create/Update (can be added at caller level). Optimistic locking: Pass db.Where("version = ?", oldVersion) to Update/Save for conditional updates. Value-based structs (`&EntityDAL{}`) not repository pattern - keeps it simple. Shared utilities in pkg/generator/common: GetColumnOptions(), GetColumnName(), ToSnakeCase(). Generated in separate file (e.g., user_gorm_dal.go) or subdirectory. Implementation in pkg/gorm/dal.go with comprehensive unit tests (11 tests covering detection, building, filename generation, code structure). Benefits: Reduces 80%+ of repetitive service code, consistent CRUD patterns across entities, type-safe primary key handling, optional but powerful hooks for custom logic, built-in optimistic locking support. |
| Cross-DB serialization tags | For cross-database compatibility (SQLite/PostgreSQL), repeated fields, maps, and complex types require `serializer:json` GORM tag instead of database-specific tags like `type:jsonb` or `type:text[]`. The `serializer:json` tag uses GORM's built-in JSON serializer which works across all database backends. Implementation: Added validation in pkg/gorm/generator.go (validateSerializerTags function) to warn at parse time when complex types lack serializer tags. Validation skips embedded fields (detected via `embedded` tag) and types with implement_scanner enabled. Registry lookup: Uses MessageRegistry.LookupTargetMessage() to resolve source message types to their GORM targets, then checks if target has implement_scanner option. Warnings format: `[WARN] Field 'MessageName.FieldName' (repeated message): missing serializer:json tag for cross-DB compatibility (SQLite/PostgreSQL)`. Applied to repeated primitives, repeated messages, maps with any value type. Benefits: Early detection of serialization issues, clear actionable warnings, prevents runtime database errors, guides users to correct tags. |
| Optional Valuer/Scanner generation | Added `implement_scanner` boolean option to GormOptions annotation for opt-in driver.Valuer and sql.Scanner implementation. When enabled, generates Value() and Scan() methods using encoding/json for automatic JSON serialization. Implementation: Added ImplementScanner field to collector.MessageInfo and gorm.StructData, extracted from GormOptions.ImplementScanner in collector. Template conditionally generates scanner_valuer.go.tmpl only when .ImplementScanner is true. Value() method marshals struct to JSON, Scan() method unmarshals from JSON or string. Validation smart suppression: When field's GORM target type has implement_scanner, validation skips warnings about missing serializer:json tags since GORM will use Valuer/Scanner methods automatically. Registry lookup resolves source messages to GORM targets (e.g., api.GamePlayer → GamePlayerGORM) before checking implement_scanner flag. Applied to both repeated fields and map value types. Benefits: Eliminates need for manual serializer:json tags on every field, cleaner proto definitions, automatic JSON handling, opt-in avoids conflicts with embedded structs, validation system understands the relationship. Example: `message GamePlayerGORM { option (dal.v1.gorm) = { source: "weewar.v1.GamePlayer", implement_scanner: true }; }` generates Valuer/Scanner, then `repeated GamePlayerGORM players = 1;` needs no serializer tag and generates no warning. |
| Late-binding table/kind support | Added `optional bool dal` field to GormOptions and DatastoreOptions annotations to control DAL generation independently of table/kind specification. Enables late-binding pattern where the same struct can be stored in different tables at runtime. Implementation: DAL struct has TableName field set via constructor `NewXDAL(tableName)`. db() helper method returns `db.Table(tableName)` when TableName is set, otherwise returns db unchanged to let GORM resolve table via struct's TableName() method. GenerateDAL flag computed from: (1) explicit `dal` annotation if set, (2) otherwise defaults to true if table is specified. Datastore Kind() method now conditional - only generated when kind is explicitly specified. Benefits: Same struct usable in multiple tables (e.g., Address in user_addresses and company_addresses), DAL generation decoupled from table specification, full backward compatibility with existing protos, runtime flexibility without code duplication. Tests: TestDALLateBinding verifies same struct in different tables, TestDALTableNameEmpty verifies fallback to struct's TableName() method. |
| Datastore tags support | Added `datastore_tags` field to ColumnOptions annotation for Google Cloud Datastore field customization. Mirrors `gorm_tags` pattern - use native Datastore tag syntax directly. Supported tags: `-` (ignore field, exclude from properties), `noindex` (don't index field), `omitempty` (omit if empty), `flatten` (flatten nested struct). Multiple tags joined with commas: `datastore_tags: ["noindex", "omitempty"]` generates `datastore:"field_name,noindex,omitempty"`. Implementation in pkg/datastore/generator.go buildFieldTags() reads tags from column options and applies them. Special handling for `-` tag returns early with `datastore:"-"`. Benefits: Consistent with gorm_tags pattern, full access to native Datastore tag features, no abstraction layer - users use familiar Datastore syntax. |
| Oneof field access | Proto oneof fields require getter method access (`src.GetFieldName()`) instead of direct field access (`src.FieldName`) in generated converter code. This is because proto oneofs don't expose member fields directly on the struct - they're accessed via generated getter methods. Implementation: Added `SourceIsOneofMember` boolean field to FieldMapping struct. BuildFieldMapping() detects real oneofs (not synthetic proto3 optional) via `field.Oneof != nil && !field.Oneof.Desc.IsSynthetic()`. Added `sourceFieldAccess()` helper and `srcField` template function to generate correct access expression. For FromTarget direction (Target→Proto), oneof fields are filtered out entirely since proto struct literals can't set oneof fields directly - users must handle via decorator function. Tests: pkg/generator/converter/oneof_test.go validates detection of real oneofs vs synthetic proto3 optional. Benefits: google.protobuf.Value and similar oneof-based well-known types now generate correct converter code. |
| Conditional source package imports | Struct files (e.g., user_gorm.go) only import the source package when actually needed for enum types. Previously, the import was added unconditionally when a source message existed, causing "imported and not used" errors when no enum fields were present. Implementation: After building StructData, generators now check if any field type string contains the source package prefix (e.g., "v1."). Import added only when a match is found. Applied to both GORM and Datastore struct generators. Benefits: Cleaner generated code without unused imports, eliminates build errors. |
| PropertyLoadSaver for Datastore maps | Google Cloud Datastore doesn't natively support Go map types (e.g., `map[string]int64`). Added `implement_property_loader` option to DatastoreOptions annotation. When enabled, generates `Save()` and `Load()` methods implementing the PropertyLoadSaver interface. Map fields are serialized to JSON bytes and stored as blob properties. Template in `property_load_saver.go.tmpl` generates: (1) `Save()` extracts non-map fields via temporary struct, then JSON-encodes map fields as properties with NoIndex=true; (2) `Load()` separates map properties, loads non-map fields via temporary struct, then JSON-decodes map properties back to maps. Follows same opt-in pattern as GORM's `implement_scanner`. Example: `implement_property_loader: true` in datastore_options generates PropertyLoadSaver for any struct with map fields. Benefits: Enables storing map fields in Datastore without manual serialization, consistent pattern with GORM's Valuer/Scanner approach. |
| Deterministic code generation | Generated code must be identical across multiple regenerations (no random ordering due to Go map iteration). Multiple levels of determinism: (1) Field merge tie-breaking uses field name as secondary sort key when field numbers are equal (target adds new fields with same number as source). (2) Embedded types sorted alphabetically by full name before generation. (3) File groups sorted by proto path before iteration. (4) Imports sorted by path in ImportMap.ToSlice(). (5) Error messages sorted for consistent validation output. Implementation spread across: pkg/generator/common/field_merge.go (sort with secondary key), pkg/generator/common/imports.go (sorted ToSlice), pkg/gorm/generator.go (sorted embedded types and file groups), pkg/datastore/generator.go (sorted file groups), pkg/gorm/dal.go and pkg/datastore/dal.go (sorted file groups), pkg/generator/common/message_registry.go (sorted error messages). Benefits: Reproducible builds, easier code review (no spurious diffs), version control friendly. |
