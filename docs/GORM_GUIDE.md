# GORM Generator Guide

Complete guide for using protoc-gen-dal-gorm to generate GORM models and converters.

## Installation

```bash
go install github.com/panyam/protoc-gen-dal/cmd/protoc-gen-dal-gorm@latest
```

## Basic Usage

### 1. Define API Proto

Clean API message focused on transport:

```protobuf
// api/user.proto
syntax = "proto3";
package api.v1;

import "google/protobuf/timestamp.proto";

message User {
  uint32 id = 1;
  string name = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}
```

### 2. Define GORM Sidecar

Database-optimized schema with GORM annotations:

```protobuf
// gorm/user.proto
syntax = "proto3";
package gorm;

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

  string name = 2 [(dal.v1.column) = {
    gorm_tags: ["not null", "size:255"]
  }];

  string email = 3 [(dal.v1.column) = {
    gorm_tags: ["uniqueIndex", "not null"]
  }];

  int64 created_at = 4 [(dal.v1.column) = {
    gorm_tags: ["autoCreateTime"]
  }];

  int64 updated_at = 5 [(dal.v1.column) = {
    gorm_tags: ["autoUpdateTime"]
  }];
}
```

### 3. Configure buf.gen.yaml

```yaml
version: v2
plugins:
  - local: protoc-gen-dal-gorm
    out: gen
    opt: paths=source_relative
```

### 4. Generate Code

```bash
buf generate
```

### 5. Use Generated Code

```go
package main

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "yourproject/gen/gorm"
    "yourproject/gen/api"
)

func main() {
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

    // Migrate schema
    db.AutoMigrate(&gorm.UserGORM{})

    // API to database
    apiUser := &api.User{
        Name: "Alice",
        Email: "alice@example.com",
    }
    dbUser, err := gorm.UserToUserGORM(apiUser, nil, nil)
    db.Create(dbUser)

    // Database to API
    var fetchedUser gorm.UserGORM
    db.First(&fetchedUser, 1)
    apiUser, err = gorm.UserFromUserGORM(nil, &fetchedUser, nil)
}
```

## Advanced Features

### Nested Messages

Define converters for nested message types:

```protobuf
// API proto
message Book {
  uint32 id = 1;
  string title = 2;
  Author author = 3;
}

message Author {
  uint32 id = 1;
  string name = 2;
}

// GORM proto
message BookGORM {
  option (dal.v1.table) = {source: "api.v1.Book"};

  uint32 id = 1 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
  string title = 2;
  uint32 author_id = 3;  // Foreign key

  // Embedded association
  AuthorGORM author = 4 [(dal.v1.column) = {
    gorm_tags: ["foreignKey:AuthorID"]
  }];
}

message AuthorGORM {
  option (dal.v1.table) = {source: "api.v1.Author"};

  uint32 id = 1 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
  string name = 2;
}
```

Generated converter automatically calls `AuthorToAuthorGORM` for nested field.

### Optional Fields

Use proto3 `optional` keyword for nullable fields:

```protobuf
// API proto
message User {
  uint32 id = 1;
  string name = 2;
  optional string nickname = 3;  // Nullable
}

// GORM proto
message UserGORM {
  option (dal.v1.table) = {source: "api.v1.User"};

  uint32 id = 1 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
  string name = 2;
  optional string nickname = 3;  // Generates as *string in Go
}
```

Generated Go code:
```go
type UserGORM struct {
    ID       uint32
    Name     string
    Nickname *string  // Pointer type from optional keyword
}
```

### Collections

**Repeated Primitives - Database Array:**
```protobuf
message Product {
  repeated string tags = 1;
}

message ProductGORM {
  option (dal.v1.table) = {source: "api.v1.Product"};

  repeated string tags = 1 [(dal.v1.column) = {
    gorm_tags: ["type:text[]"]  // Postgres array
  }];
}
```

**Repeated Primitives - JSONB:**
```protobuf
message ProductGORM {
  repeated string tags = 1 [(dal.v1.column) = {
    gorm_tags: ["type:jsonb"]  // Postgres JSONB
  }];
}
```

**Repeated Messages:**
```protobuf
message Library {
  repeated Author contributors = 1;
}

message LibraryGORM {
  option (dal.v1.table) = {source: "api.v1.Library"};

  repeated AuthorGORM contributors = 1 [(dal.v1.column) = {
    gorm_tags: ["foreignKey:LibraryID"]
  }];
}
```

**Maps with Primitives:**
```protobuf
message Product {
  map<string, string> metadata = 1;
}

message ProductGORM {
  option (dal.v1.table) = {source: "api.v1.Product"};

  map<string, string> metadata = 1 [(dal.v1.column) = {
    gorm_tags: ["type:jsonb"]
  }];
}
```

**Maps with Messages:**
```protobuf
message Organization {
  map<string, Author> departments = 1;
}

message OrganizationGORM {
  option (dal.v1.table) = {source: "api.v1.Organization"};

  map<string, AuthorGORM> departments = 1 [(dal.v1.column) = {
    gorm_tags: ["type:jsonb"]
  }];
}
```

### Custom Transformations with Decorators

Apply custom logic during conversion:

```go
// Normalize email on save
saveDecorator := func(src *api.User, dest *gorm.UserGORM) error {
    dest.Email = strings.ToLower(src.Email)
    return nil
}

dbUser, err := gorm.UserToUserGORM(apiUser, nil, saveDecorator)

// Mask sensitive data on load
loadDecorator := func(dest *api.User, src *gorm.UserGORM) error {
    dest.Email = maskEmail(src.Email)
    return nil
}

apiUser, err := gorm.UserFromUserGORM(nil, &dbUser, loadDecorator)
```

### In-place Conversion

Reuse existing struct to avoid allocations:

```go
var dbUser gorm.UserGORM

// First conversion
gorm.UserToUserGORM(apiUser1, &dbUser, nil)
db.Create(&dbUser)

// Reuse same struct
gorm.UserToUserGORM(apiUser2, &dbUser, nil)
db.Create(&dbUser)
```

## Common Patterns

### Foreign Keys

```protobuf
message BookGORM {
  uint32 author_id = 1 [(dal.v1.column) = {
    gorm_tags: [
      "foreignKey:AuthorID",
      "references:ID",
      "constraint:OnDelete:CASCADE,OnUpdate:CASCADE"
    ]
  }];
}
```

### Composite Primary Keys

```protobuf
message BookEditionGORM {
  string book_id = 1 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
  int32 edition = 2 [(dal.v1.column) = {gorm_tags: ["primaryKey"]}];
  string title = 3;
}
```

### Indexes

```protobuf
message UserGORM {
  string email = 1 [(dal.v1.column) = {
    gorm_tags: ["uniqueIndex:idx_email"]
  }];

  string first_name = 2 [(dal.v1.column) = {
    gorm_tags: ["index:idx_name,priority:1"]
  }];

  string last_name = 3 [(dal.v1.column) = {
    gorm_tags: ["index:idx_name,priority:2"]
  }];
}
```

### Soft Deletes

```protobuf
message UserGORM {
  int64 deleted_at = 1 [(dal.v1.column) = {
    gorm_tags: ["index"]
  }];
}
```

Generated code automatically uses GORM's soft delete feature.

### Default Values

```protobuf
message UserGORM {
  string status = 1 [(dal.v1.column) = {
    gorm_tags: ["default:active"]
  }];

  string id = 2 [(dal.v1.column) = {
    gorm_tags: ["type:uuid", "default:gen_random_uuid()"]
  }];
}
```

### Database-specific Types

**Postgres:**
```protobuf
string id = 1 [(dal.v1.column) = {
  gorm_tags: ["type:uuid"]
}];

map<string, string> metadata = 2 [(dal.v1.column) = {
  gorm_tags: ["type:jsonb"]
}];

repeated string tags = 3 [(dal.v1.column) = {
  gorm_tags: ["type:text[]"]
}];
```

**MySQL:**
```protobuf
map<string, string> metadata = 1 [(dal.v1.column) = {
  gorm_tags: ["type:json"]
}];

string content = 2 [(dal.v1.column) = {
  gorm_tags: ["type:text"]
}];
```

**SQLite:**
```protobuf
int64 created_at = 1 [(dal.v1.column) = {
  gorm_tags: ["type:datetime"]
}];
```

## Type Conversions

Automatic conversions between proto and Go types:

| Proto Type | GORM Type | Notes |
|------------|-----------|-------|
| `uint32` | `uint32` | Direct mapping |
| `int32` | `int32` | Direct mapping |
| `int64` | `int64` | Direct mapping |
| `string` | `string` | Direct mapping |
| `bool` | `bool` | Direct mapping |
| `bytes` | `[]byte` | Direct mapping |
| `google.protobuf.Timestamp` | `int64` | Unix seconds |
| `optional string` | `*string` | Pointer for nullability |
| `repeated string` | `[]string` | Slice |
| `map<string, string>` | `map[string]string` | Map |

## Error Handling

All converters return errors for nested message conversions:

```go
dbUser, err := gorm.UserToUserGORM(apiUser, nil, nil)
if err != nil {
    // Handle conversion error
    // Error wraps nested converter errors with field context
    log.Printf("conversion failed: %v", err)
}
```

Error messages include field context:
```
converting Author: converting Books[0]: field validation failed
```

## Performance Tips

1. **Reuse destination structs** to avoid allocations:
```go
var dbUser gorm.UserGORM
for _, apiUser := range apiUsers {
    gorm.UserToUserGORM(apiUser, &dbUser, nil)
    db.Create(&dbUser)
}
```

2. **Batch operations** with GORM:
```go
dbUsers := make([]gorm.UserGORM, len(apiUsers))
for i, apiUser := range apiUsers {
    gorm.UserToUserGORM(apiUser, &dbUsers[i], nil)
}
db.CreateInBatches(dbUsers, 100)
```

3. **Preload associations** to avoid N+1 queries:
```go
var dbBooks []gorm.BookGORM
db.Preload("Author").Find(&dbBooks)

apiBooks := make([]*api.Book, len(dbBooks))
for i, dbBook := range dbBooks {
    apiBooks[i], _ = gorm.BookFromBookGORM(nil, &dbBook, nil)
}
```

## Troubleshooting

**Missing converter for nested type:**
```
Warning: No converter found for field 'Author'. Skipping automatic conversion.
Add source annotation.
```

Solution: Add `source` annotation to nested message:
```protobuf
message AuthorGORM {
  option (dal.v1.table) = {source: "api.v1.Author"};
  // ...
}
```

**Compilation error on generated code:**

Check that GORM tags are valid. The generator passes tags through without validation.

**Nil pointer dereference:**

Use `optional` keyword in proto for nullable fields:
```protobuf
optional string nickname = 1;  // Generates *string
```

## Examples

See `tests/protos/gorm/user.proto` for comprehensive examples including:
- Basic CRUD operations
- Nested messages
- Collections (repeated and map fields)
- Custom timestamps
- Indexes
- Default values
