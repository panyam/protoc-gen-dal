# protoc-gen-dal: Design Summary

## Project Vision

Generate Data Access Layer (DAL) converters that transform between clean API protobuf messages and various database/datastore representations.

## Core Principles

### 1. Targets vs Vehicles

**Target** = Where data lives (The Hero)
- postgres, firestore, mongodb, dynamodb, etc.
- The primary concern - what database/datastore you're using

**Vehicle** = How you access it (Implementation Detail)
- Go + database/sql, Go + GORM, Python + psycopg2, TypeScript + Prisma
- Just different ways to access the same target

### 2. Sidecar Pattern (API Purity)

**Keep API protos clean** - No database concerns in API definitions!

```
proto/
├── api/
│   └── library/v1/
│       └── book.proto          # Clean API - no DB pollution!
│
└── dal/
    └── library/v1/
        ├── book_postgres.proto      # PostgreSQL schema
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

### One Binary Per Target+Vehicle

```
cmd/
├── protoc-gen-dal-postgres-raw/
├── protoc-gen-dal-postgres-gorm/
├── protoc-gen-dal-firestore-raw/
└── protoc-gen-python-dal-postgres-raw/
```

**Each main.go is thin:**
1. Collect messages for target (using shared collector)
2. Delegate to generator
3. Done

**Benefits:**
- Focused single responsibility
- Easy to add new target+vehicle combinations
- Independent versioning possible
- Clear separation of concerns

### Package Structure

```
protoc-gen-dal/                    # Monorepo
├── cmd/
│   ├── protoc-gen-dal-postgres-gorm/
│   │   └── main.go                   # Thin: collect → delegate
│   ├── protoc-gen-dal-postgres-raw/
│   └── protoc-gen-dal-firestore-raw/
│
├── pkg/
│   ├── collector/
│   │   └── collector.go              # Shared: collect messages by target
│   │
│   ├── postgres/
│   │   ├── gorm_go.go                # Generator for Go+GORM
│   │   ├── raw_go.go                 # Generator for Go+raw SQL
│   │   ├── raw_python.go             # Generator for Python+psycopg2
│   │   └── templates/
│   │       ├── go_gorm/
│   │       ├── go_raw/
│   │       └── python_raw/
│   │
│   ├── firestore/
│   │   ├── raw_go.go
│   │   └── templates/
│   │
│   └── mongodb/
│       └── raw_go.go
│
└── proto/
    └── dal/v1/
        └── annotations.proto         # Shared annotations
```

## Annotations Design

### Target-Focused Annotations

```protobuf
// Postgres target
message PostgresOptions {
  string source = 1;        // "library.v1.Book" - links to API message
  string table = 2;         // "books"
  string schema = 3;        // "public"

  // Vehicle-specific hints (optional)
  GormHints gorm = 10;
}

message GormHints {
  bool disable_soft_delete = 1;
  repeated string embedded = 2;
}

// Firestore target
message FirestoreOptions {
  string source = 1;
  string collection = 2;
}
```

### Field-Level Annotations

```protobuf
message ColumnOptions {
  string name = 1;              // Column name
  string type = 2;              // DB-specific type ("UUID", "JSONB")
  bool primary_key = 3;
  bool nullable = 4;
  string default = 5;
  bool unique = 6;
  bool auto_increment = 7;
  bool auto_create_time = 8;
  bool auto_update_time = 9;
}

message ForeignKeyOptions {
  string references = 1;        // "authors.id"
  ReferentialAction on_delete = 2;
  ReferentialAction on_update = 3;
}
```

## Generated Code Pattern

### GORM Example

**Input:**
```protobuf
message BookPostgres {
  option (dal.v1.postgres) = {
    source: "library.v1.Book"
    table: "books"
  };

  string id = 1 [(dal.v1.column) = {primary_key: true}];
  string title = 2;
  int64 published_at = 3;  // Transformed from Timestamp
}
```

**Output:**
```go
// Generated GORM model
type BookGORM struct {
    ID          string `gorm:"primaryKey;column:id"`
    Title       string `gorm:"column:title"`
    PublishedAt int64  `gorm:"column:published_at"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
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
| One binary per target+vehicle | Focused, easy to add new combinations |
| Grouper phase | Need all messages to resolve cross-file relationships |
| Type safety | Compile-time checks, no runtime map lookups |
| Function pointer hooks | Avoid interface boilerplate |
| Decorator pattern | Simple, type-safe transformations |
| Target-focused annotations | Database is the hero, not the language |
| Opt-in messages | Explicit about which messages are DB entities |
| TDD workflow | Prevent over-engineering, ensure correctness |
