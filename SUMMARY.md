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
| Built-in type converters | Provide smart defaults for common conversions (`*timestamppb.Timestamp` ↔ `int64`, numeric types). Custom converters via annotations (`to_func`/`from_func`) override defaults. Nested message converters auto-generated when available. Decorator function handles edge cases. Priority: custom > built-in > skip (decorator). |
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
| Datastore converter generation | Reuses GORM converter infrastructure with Datastore-specific type conversions. Generates bidirectional converters (XToY, XFromY) with decorator support. Built-in type conversions: uint32↔string (IDs stored as strings in Datastore keys), Timestamp↔int64 (Unix seconds). Helper functions generated in converter file: timestampToInt64, int64ToTimestamp, mustParseUint. Follows same pattern as GORM converters for consistency. Full support for scalar fields, repeated fields (direct assignment for primitives, loop-based for messages), and map fields (direct assignment for primitive values, loop-based for message values). Converter registry tracks available nested converters for automatic loop generation. |
| Shared generator utilities | Extract common code from GORM and Datastore generators into reusable packages: `pkg/generator/common` (file naming, package extraction, type mapping), `pkg/generator/registry` (converter tracking). Prevents "relearning" mistakes like map field handling. Centralizes proto→Go type mapping in `ProtoFieldToGoType()` which correctly generates `map[K]V` types instead of entry structs. ~500 lines of duplicate code consolidated. All utilities have unit tests. Future targets (postgres-raw, firestore, mongodb) can immediately leverage these utilities. |
| Datastore repeated/map fields | Mirrors GORM approach: repeated scalars ([]string, []int32) use direct assignment. Maps with scalar values (map[string]string) use direct assignment with nil checks. Repeated message types ([]Author) generate loop-based conversion calling element converter. Maps with message values (map[string]Author) generate loop-based conversion calling value converter. Proper bidirectional handling: API uses pointers ([]*api.Author), Datastore uses values ([]AuthorDatastore). Template uses $converter variable to access parent context for package name in nested conversions. |
| Converter strategy utilities | Separate user intent (ConversionType: what conversion to apply) from implementation detail (FieldRenderStrategy: how to render it). ConversionType determined by generator based on proto annotations, field types, and built-in rules. FieldRenderStrategy derived from ConversionType + field characteristics (pointer, repeated, map). Strategy types: StrategyInlineValue (struct literal), StrategySetterSimple/Transform/WithError/IgnoreError (post-construction statements), StrategyLoopRepeated/Map (loop blocks). Utilities in `pkg/generator/converter`: collection helpers (CheckMapValueType, CheckRepeatedElementType), conversion detection (IsTimestampToInt64, IsNumericConversion), template helper generators (TimestampHelperFunctions, MustParseUintHelper). Enables template simplification: templates render based on strategy, not conditional logic scattered throughout. |
| Struct literal + setters pattern | Unified converter template structure across GORM and Datastore generators. Three-section pattern: (1) Struct literal initialization for inline fields (non-pointer, no errors, direct assignments or simple transformations), (2) Post-construction setter statements for pointer fields and error handling, (3) Separate loop-based conversion sections for repeated/map message types. Fields classified into groups (ToTargetInlineFields, ToTargetSetterFields, ToTargetLoopFields) by render strategy. Template helpers (isInlineValue, isSetterSimple, etc.) make templates readable. Empty ToTargetCode/FromTargetCode handled via fallback: `{{ if .ToTargetCode }}{{ .ToTargetCode }}{{ else }}src.{{ .SourceField }}{{ end }}`. Benefits: cleaner generated code (single struct initialization vs scattered assignments), consistent pattern across all targets, easier to extend for future targets. |
