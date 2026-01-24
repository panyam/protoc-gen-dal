# protoc-gen-dal

Protocol Buffer compiler plugins for generating Data Access Layer code. Converts between API messages (clean, transport-focused) and database entities (storage-optimized).

## Overview

protoc-gen-dal uses a sidecar pattern: keep your API protos clean, define database schemas in separate proto files. Generate type-safe converters between API messages and database entities.

Supported targets:
- **GORM** - Go ORM (works with postgres, mysql, sqlite via GORM dialects)
- **Google Cloud Datastore** - Go NoSQL database

## Quick Start

### Installation

```bash
go install github.com/panyam/protoc-gen-dal/cmd/protoc-gen-dal-gorm@latest
go install github.com/panyam/protoc-gen-dal/cmd/protoc-gen-dal-datastore@latest
```

### Example: GORM

**API proto** (`api/user.proto`):
```protobuf
syntax = "proto3";
package api.v1;

message User {
  uint32 id = 1;
  string name = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;
}
```

**GORM sidecar** (`dal/user_gorm.proto`):
```protobuf
syntax = "proto3";
package dal;

import "dal/v1/annotations.proto";
import "api/user.proto";

message UserGORM {
  option (dal.v1.table) = {
    source: "api.v1.User"
    name: "users"
  };

  uint32 id = 1 [(dal.v1.column) = {
    gorm_tags: ["primaryKey", "autoIncrement"]
  }];
  string name = 2;
  string email = 3 [(dal.v1.column) = {
    gorm_tags: ["uniqueIndex"]
  }];
  int64 created_at = 4;  // Timestamp stored as Unix seconds
}
```

**buf.gen.yaml**:
```yaml
version: v2
plugins:
  - local: protoc-gen-dal-gorm
    out: gen/gorm
    opt: paths=source_relative
```

**Generated code** (`gen/gorm/dal/user_gorm_gorm.go`):
```go
type UserGORM struct {
    ID        uint32 `gorm:"primaryKey;autoIncrement"`
    Name      string
    Email     string `gorm:"uniqueIndex"`
    CreatedAt int64
}

func (UserGORM) TableName() string {
    return "users"
}
```

**Generated converters** (`gen/gorm/dal/user_gorm_converters.go`):
```go
func UserToUserGORM(
    src *api.User,
    dest *UserGORM,
    decorator func(*api.User, *UserGORM) error,
) (out *UserGORM, err error)

func UserFromUserGORM(
    dest *api.User,
    src *UserGORM,
    decorator func(*api.User, *UserGORM) error,
) (out *api.User, err error)
```

**Usage**:
```go
// API to database
apiUser := &api.User{Name: "Alice", Email: "alice@example.com"}
dbUser, err := UserToUserGORM(apiUser, nil, nil)
db.Create(dbUser)

// Database to API
var dbUser UserGORM
db.First(&dbUser, 1)
apiUser, err := UserFromUserGORM(nil, &dbUser, nil)
```

## Features

### Multi-Service Support

The plugin preserves directory structure to avoid filename collisions. Proto files in different directories generate to separate subdirectories:

```
# Input protos
likes/v1/gorm.proto
tags/v1/gorm.proto

# Output (with out: gen/gorm)
gen/gorm/likes/v1/gorm_gorm.go
gen/gorm/tags/v1/gorm_gorm.go
```

This enables monorepos where multiple services have their own `gorm.proto` files without collision.

### DAL Helper Methods (Optional)

Generate basic CRUD helper methods to eliminate service-layer boilerplate. Enable with `generate_dal=true`:

**buf.gen.yaml**:
```yaml
plugins:
  - local: protoc-gen-dal-gorm
    out: gen
    opt:
      - paths=source_relative
      - generate_dal=true
      - dal_output_dir=dal  # Optional: put helpers in subdirectory
      - entity_import_path=github.com/example/gen  # Required when using subdirectory with buf managed mode
```

**Generated helpers** (`gen/gorm/dal/user_gorm_dal.go`):
```go
type UserGORMDAL struct {
    // Hook called when Save creates new records
    WillCreate func(context.Context, *UserGORM) error
}

func (d *UserGORMDAL) Create(ctx context.Context, db *gorm.DB, obj *UserGORM) error
func (d *UserGORMDAL) Update(ctx context.Context, db *gorm.DB, obj *UserGORM) error
func (d *UserGORMDAL) Save(ctx context.Context, db *gorm.DB, obj *UserGORM) error
func (d *UserGORMDAL) Get(ctx context.Context, db *gorm.DB, id uint32) (*UserGORM, error)
func (d *UserGORMDAL) Delete(ctx context.Context, db *gorm.DB, id uint32) error
func (d *UserGORMDAL) List(ctx context.Context, query *gorm.DB) ([]*UserGORM, error)
func (d *UserGORMDAL) BatchGet(ctx context.Context, db *gorm.DB, ids []uint32) ([]*UserGORM, error)
```

**Usage**:
```go
dal := &UserGORMDAL{
    WillCreate: func(ctx context.Context, user *UserGORM) error {
        user.CreatedAt = time.Now()
        user.UpdatedAt = time.Now()
        return nil
    },
}

// Create (fails if ID already exists)
err := dal.Create(ctx, db, userGorm)

// Update (fails if record doesn't exist)
err := dal.Update(ctx, db, userGorm)

// Update with optimistic locking (conditional update)
err := dal.Update(ctx, db.Where("version = ?", oldVersion), userGorm)
if errors.Is(err, gorm.ErrRecordNotFound) {
    // Record not found or version mismatch
}

// Save (upsert - create or update)
err := dal.Save(ctx, db, userGorm)

// Save with conditional update (optimistic locking)
err := dal.Save(ctx, db.Where("version = ?", oldVersion), userGorm)

// Get by ID (returns nil, nil if not found)
user, err := dal.Get(ctx, db, 123)

// List with custom query
users, err := dal.List(ctx, db.Where("active = ?", true).Order("name asc"))

// Batch get
users, err := dal.BatchGet(ctx, db, []uint32{1, 2, 3})

// Delete
err := dal.Delete(ctx, db, 123)
```

**Composite primary keys**:
```go
// Get by composite key
edition, err := dal.Get(ctx, db, "book-123", 2)  // book_id, edition_number

// BatchGet with key struct
type BookEditionKey struct {
    BookId        string
    EditionNumber int32
}
keys := []BookEditionKey{{"book-123", 1}, {"book-456", 2}}
editions, err := dal.BatchGet(ctx, db, keys)
```

**Configuration options**:
- `generate_dal=true` - Enable DAL generation
- `dal_filename_suffix="_dal"` - Filename suffix (default: `_dal`)
- `dal_filename_prefix=""` - Optional filename prefix
- `dal_output_dir=""` - Optional subdirectory (e.g., `dal`)
- `entity_import_path=""` - Entity package import path (auto-detected if not specified, required when using buf managed mode with subdirectory)

Primary keys are auto-detected from `gorm_tags: ["primaryKey"]` or fallback to `id` field. Messages without primary keys are skipped.

### Type Conversions

Built-in conversions handle common type mismatches:

- `google.protobuf.Timestamp` ↔ `time.Time` (native time type for database storage)
- `uint32` ↔ `string` (Datastore keys)
- Numeric types with casting (`int32` → `int64`, etc.)

### Well-Known Types

Converters handle protobuf well-known types including those with `oneof` fields:

**`google.protobuf.Struct`** - For arbitrary JSON data:
```protobuf
message UserGORM {
  // Store arbitrary extras as JSON
  StructGORM extras = 20 [(dal.v1.column) = {
    gorm_tags: ["serializer:json"]
  }];
}

// Define GORM wrapper with Valuer/Scanner
message StructGORM {
  option (dal.v1.gorm) = {
    source: "google.protobuf.Struct",
    implement_scanner: true
  };
}
```

**`google.protobuf.Value`** - For variant types (oneof):
```protobuf
message ValueGORM {
  option (dal.v1.gorm) = {
    source: "google.protobuf.Value",
    implement_scanner: true
  };
}
```

Converters correctly use getter methods (`src.GetNullValue()`, `src.GetStringValue()`) for oneof fields since they're not directly accessible as struct fields.

### Nested Messages

Converters auto-generate for nested message types when both sides have converters defined:

```protobuf
message BookGORM {
  option (dal.v1.table) = {source: "api.v1.Book"};

  uint32 id = 1;
  AuthorGORM author = 2;  // Nested message
}

message AuthorGORM {
  option (dal.v1.table) = {source: "api.v1.Author"};

  uint32 id = 1;
  string name = 2;
}
```

Generated converter calls `AuthorToAuthorGORM` automatically for the nested field.

### Collections

**Repeated primitives** - direct assignment:
```protobuf
repeated string tags = 1;  // API
repeated string tags = 1 [(dal.v1.column) = {
  gorm_tags: ["type:text[]"]
}];  // Postgres array
```

**Repeated messages** - loop-based conversion:
```protobuf
repeated Author contributors = 1;  // API
repeated AuthorGORM contributors = 1;  // GORM
```

**Maps with primitives** - direct assignment:
```protobuf
map<string, string> metadata = 1;  // API
map<string, string> metadata = 1 [(dal.v1.column) = {
  gorm_tags: ["type:jsonb"]
}];  // GORM
```

**Maps with messages** - loop-based conversion:
```protobuf
map<string, Author> authors_by_id = 1;  // API
map<string, AuthorGORM> authors_by_id = 1;  // GORM
```

### Custom Transformations

Use decorator functions for custom field transformations:

```go
decorator := func(src *api.User, dest *UserGORM) error {
    // Custom logic here
    dest.NormalizedEmail = strings.ToLower(src.Email)
    return nil
}

dbUser, err := UserToUserGORM(apiUser, nil, decorator)
```

### In-place Conversion

Converters accept destination parameter for in-place modification:

```go
var dbUser UserGORM
UserToUserGORM(apiUser, &dbUser, nil)  // Modifies dbUser in place
```

## Annotations Reference

### Table-level

**TableOptions** - Applied to GORM messages:
```protobuf
option (dal.v1.table) = {
  source: "api.v1.User"  // Required: source API message
  name: "users"          // Optional: table name (default: lowercase message name)
};
```

**DatastoreOptions** - Applied to Datastore messages:
```protobuf
option (dal.v1.table) = {
  target_datastore: {
    source: "api.v1.User"
    kind: "User"           // Optional: Datastore kind (default: message name)
    namespace: "prod"      // Optional: Datastore namespace
  }
};
```

### Field-level

**ColumnOptions** - Applied to individual fields:
```protobuf
uint32 id = 1 [(dal.v1.column) = {
  gorm_tags: ["primaryKey", "autoIncrement"]
}];

string email = 2 [(dal.v1.column) = {
  gorm_tags: ["uniqueIndex", "type:varchar(255)"]
}];
```

**Datastore tags**:
```protobuf
string id = 1 [(dal.v1.column) = {
  datastore_tags: ["-"]  // Exclude from Datastore (used for Key field)
}];

string large_text = 2 [(dal.v1.column) = {
  datastore_tags: ["noindex"]  // Don't index this field
}];
```

## Target-specific Guides

### GORM

GORM is database-agnostic. The same generated code works with postgres, mysql, sqlite via GORM dialects. Specify database-specific types in tags:

```protobuf
// Postgres JSONB
map<string, string> metadata = 1 [(dal.v1.column) = {
  gorm_tags: ["type:jsonb"]
}];

// MySQL JSON
map<string, string> metadata = 1 [(dal.v1.column) = {
  gorm_tags: ["type:json"]
}];

// UUID primary key
string id = 1 [(dal.v1.column) = {
  gorm_tags: ["primaryKey", "type:uuid", "default:gen_random_uuid()"]
}];
```

**Foreign keys**:
```protobuf
uint32 author_id = 1 [(dal.v1.column) = {
  gorm_tags: ["foreignKey:AuthorID", "references:ID", "constraint:OnDelete:CASCADE"]
}];
```

**Composite primary keys**:
```protobuf
string book_id = 1 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
int32 edition = 2 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
```

### Google Cloud Datastore

Datastore entities are generated with `Kind()` methods:

```go
type UserDatastore struct {
    Key       *datastore.Key `datastore:"-"`  // Manually managed
    Id        string
    Name      string
    Email     string
    CreatedAt int64
}

func (UserDatastore) Kind() string {
    return "User"
}
```

**Type conversions**:
- `uint32` → `string` (Datastore keys are strings)
- `google.protobuf.Timestamp` → `int64` (Unix seconds)

**Usage**:
```go
// Create
apiUser := &api.User{Name: "Alice"}
dsUser, err := UserToUserDatastore(apiUser, nil, nil)
key := datastore.NameKey("User", "alice", nil)
client.Put(ctx, key, dsUser)

// Retrieve
var dsUser UserDatastore
key := datastore.NameKey("User", "alice", nil)
client.Get(ctx, key, &dsUser)
apiUser, err := UserFromUserDatastore(nil, &dsUser, nil)
```

## Project Structure

```
protoc-gen-dal/
├── cmd/
│   ├── protoc-gen-dal-gorm/       # GORM plugin binary
│   └── protoc-gen-dal-datastore/  # Datastore plugin binary
├── pkg/
│   ├── collector/                 # Collects messages from proto files
│   ├── gorm/                      # GORM code generator
│   ├── datastore/                 # Datastore code generator
│   └── generator/
│       ├── common/                # Shared utilities (file naming, types, imports)
│       ├── converter/             # Converter strategy utilities
│       └── registry/              # Converter registry
├── protos/
│   └── dal/v1/
│       └── annotations.proto      # Annotation definitions
└── tests/
    └── protos/
        ├── gorm/                  # Test proto files for GORM
        └── datastore/             # Test proto files for Datastore
```

## Design Principles

1. **Sidecar pattern** - Keep API protos clean, DB schemas separate
2. **No abstraction layer** - Use native target syntax directly (GORM tags, Datastore tags)
3. **Type safety** - Compile-time checks, no runtime map lookups
4. **Opt-in** - Only messages with annotations generate code
5. **Explicit over implicit** - Require `source` annotation for nested conversions

## Development

```bash
# Build all binaries
make build

# Run all tests
make test

# Generate code from test protos
make buf
```

## Roadmap

**Completed:**
- ✅ GORM generator (Go)
- ✅ Google Cloud Datastore generator (Go)
- ✅ Nested message converters
- ✅ Repeated/map field support
- ✅ Shared generator utilities
- ✅ DAL helper methods (Save, Get, Delete, List, BatchGet)
- ✅ Composite primary key support
- ✅ Hook-based lifecycle customization

**Planned:**
- Firestore (Go)
- postgres-raw (Go + database/sql)
- MongoDB (Go)
- Python generators
- TypeScript generators

## License

Apache License 2.0
