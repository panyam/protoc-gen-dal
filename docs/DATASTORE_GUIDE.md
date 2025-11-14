# Google Cloud Datastore Generator Guide

Complete guide for using protoc-gen-dal-datastore to generate Datastore entities and converters.

## Installation

```bash
go install github.com/panyam/protoc-gen-dal/cmd/protoc-gen-dal-datastore@latest
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
}
```

### 2. Define Datastore Sidecar

Datastore-optimized schema:

```protobuf
// datastore/user.proto
syntax = "proto3";
package datastore;

import "dal/v1/annotations.proto";
import "api/user.proto";

message UserDatastore {
  option (dal.v1.table) = {
    target_datastore: {
      source: "api.v1.User"
      kind: "User"
      namespace: "production"
    }
  };

  string key = 1 [(dal.v1.column) = {
    datastore_tags: ["-"]  // Managed manually via Key field
  }];

  string id = 2;
  string name = 3;
  string email = 4 [(dal.v1.column) = {
    datastore_tags: ["noindex"]  // Don't index email
  }];

  int64 created_at = 5;
}
```

### 3. Configure buf.gen.yaml

```yaml
version: v2
plugins:
  - local: protoc-gen-dal-datastore
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
    "context"
    "cloud.google.com/go/datastore"

    "yourproject/gen/datastore"
    "yourproject/gen/api"
)

func main() {
    ctx := context.Background()
    client, _ := datastore.NewClient(ctx, "project-id")

    // API to Datastore
    apiUser := &api.User{
        Id: 123,
        Name: "Alice",
        Email: "alice@example.com",
    }

    dsUser, err := datastore.UserToUserDatastore(apiUser, nil, nil)

    // Create with explicit key
    key := datastore.NameKey(dsUser.Kind(), "alice", nil)
    key.Namespace = "production"
    _, err = client.Put(ctx, key, dsUser)

    // Datastore to API
    var fetchedUser datastore.UserDatastore
    key = datastore.NameKey("User", "alice", nil)
    key.Namespace = "production"
    err = client.Get(ctx, key, &fetchedUser)

    apiUser, err = datastore.UserFromUserDatastore(nil, &fetchedUser, nil)
}
```

## Key Management

Datastore entities include a `Key` field excluded from storage:

```go
type UserDatastore struct {
    Key       *datastore.Key `datastore:"-"`  // Not stored
    Id        string
    Name      string
    Email     string `datastore:",noindex"`
    CreatedAt int64
}

func (UserDatastore) Kind() string {
    return "User"
}
```

You manage keys manually:

```go
// Name key
key := datastore.NameKey(dsUser.Kind(), "alice", nil)

// Numeric ID key
key := datastore.IDKey(dsUser.Kind(), 123, nil)

// With namespace
key := datastore.NameKey(dsUser.Kind(), "alice", nil)
key.Namespace = "production"

// With parent
parentKey := datastore.NameKey("Account", "acct-1", nil)
key := datastore.NameKey(dsUser.Kind(), "alice", parentKey)
```

## Type Conversions

Automatic conversions between proto and Datastore types:

| Proto Type | Datastore Type | Notes |
|------------|----------------|-------|
| `uint32` | `string` | IDs as strings for key compatibility |
| `int32` | `int32` | Direct mapping |
| `int64` | `int64` | Direct mapping |
| `string` | `string` | Direct mapping |
| `bool` | `bool` | Direct mapping |
| `bytes` | `[]byte` | Direct mapping |
| `google.protobuf.Timestamp` | `int64` | Unix seconds |
| `repeated string` | `[]string` | Slice |
| `map<string, string>` | `map[string]string` | Map |

Helper functions generated:

```go
// Timestamp conversion
func timestampToInt64(ts *timestamppb.Timestamp) int64
func int64ToTimestamp(seconds int64) *timestamppb.Timestamp

// ID conversion
func mustParseUint(s string) uint64  // Panics on error (for generated code)
```

## Nested Messages

Define converters for nested types:

```protobuf
// API proto
message Library {
  uint32 id = 1;
  string name = 2;
  repeated Author contributors = 3;
}

message Author {
  uint32 id = 1;
  string name = 2;
}

// Datastore proto
message LibraryDatastore {
  option (dal.v1.table) = {
    target_datastore: {source: "api.v1.Library"}
  };

  string id = 1 [(dal.v1.column) = {datastore_tags: ["-"]}];
  string library_id = 2;
  string name = 3;
  repeated AuthorDatastore contributors = 4;
}

message AuthorDatastore {
  option (dal.v1.table) = {
    target_datastore: {source: "api.v1.Author"}
  };

  string id = 1 [(dal.v1.column) = {datastore_tags: ["-"]}];
  string author_id = 2;
  string name = 3;
}
```

Generated converter automatically calls `AuthorToAuthorDatastore` for each contributor.

## Collections

**Repeated Primitives:**
```protobuf
message Product {
  repeated string tags = 1;
}

message ProductDatastore {
  option (dal.v1.table) = {
    target_datastore: {source: "api.v1.Product"}
  };

  repeated string tags = 1;  // Direct assignment
}
```

**Repeated Messages:**
```protobuf
message Library {
  repeated Author contributors = 1;
}

message LibraryDatastore {
  repeated AuthorDatastore contributors = 1;  // Loop-based conversion
}
```

**Maps with Primitives:**
```protobuf
message Product {
  map<string, string> metadata = 1;
}

message ProductDatastore {
  map<string, string> metadata = 1;  // Direct assignment with nil check
}
```

**Maps with Messages:**
```protobuf
message Organization {
  map<string, Author> departments = 1;
}

message OrganizationDatastore {
  map<string, AuthorDatastore> departments = 1;  // Loop-based conversion
}
```

## Indexing

Control which fields are indexed:

```protobuf
message UserDatastore {
  // Indexed by default
  string email = 1;

  // Not indexed (for large text fields)
  string bio = 2 [(dal.v1.column) = {
    datastore_tags: ["noindex"]
  }];

  // Multiple tags
  string profile = 3 [(dal.v1.column) = {
    datastore_tags: ["noindex", "omitempty"]
  }];
}
```

## Namespaces

Specify namespace in table options:

```protobuf
message UserDatastore {
  option (dal.v1.table) = {
    target_datastore: {
      source: "api.v1.User"
      kind: "User"
      namespace: "production"  // Optional namespace
    }
  };
}
```

Access namespace at runtime:

```go
key := datastore.NameKey(dsUser.Kind(), "alice", nil)
key.Namespace = "production"
```

## Queries

Use generated entities with Datastore queries:

```go
// Query entities
var users []datastore.UserDatastore
query := datastore.NewQuery("User").
    Filter("Email =", "alice@example.com").
    Limit(10)

keys, err := client.GetAll(ctx, query, &users)

// Convert to API messages
apiUsers := make([]*api.User, len(users))
for i, dsUser := range users {
    // Optionally set Key field
    dsUser.Key = keys[i]

    apiUsers[i], err = datastore.UserFromUserDatastore(nil, &dsUser, nil)
}
```

## Transactions

```go
// Create in transaction
_, err := client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
    apiUser := &api.User{Name: "Alice"}
    dsUser, _ := datastore.UserToUserDatastore(apiUser, nil, nil)

    key := datastore.NameKey(dsUser.Kind(), "alice", nil)
    _, err := tx.Put(key, dsUser)
    return err
})
```

## Batch Operations

```go
// Batch put
apiUsers := []*api.User{user1, user2, user3}
dsUsers := make([]*datastore.UserDatastore, len(apiUsers))
keys := make([]*datastore.Key, len(apiUsers))

for i, apiUser := range apiUsers {
    dsUsers[i], _ = datastore.UserToUserDatastore(apiUser, nil, nil)
    keys[i] = datastore.NameKey(dsUsers[i].Kind(), fmt.Sprintf("user-%d", i), nil)
}

_, err := client.PutMulti(ctx, keys, dsUsers)

// Batch get
var fetchedUsers []datastore.UserDatastore
err = client.GetMulti(ctx, keys, &fetchedUsers)

for _, dsUser := range fetchedUsers {
    apiUser, _ := datastore.UserFromUserDatastore(nil, &dsUser, nil)
}
```

## Custom Transformations

Use decorators for custom logic:

```go
// Add metadata on save
saveDecorator := func(src *api.User, dest *datastore.UserDatastore) error {
    // Custom transformations
    dest.Email = strings.ToLower(src.Email)
    return nil
}

dsUser, err := datastore.UserToUserDatastore(apiUser, nil, saveDecorator)

// Populate Key field on load
loadDecorator := func(dest *api.User, src *datastore.UserDatastore) error {
    // Access Key field if needed
    if src.Key != nil {
        dest.Id = uint32(src.Key.ID)
    }
    return nil
}

apiUser, err := datastore.UserFromUserDatastore(nil, &dsUser, loadDecorator)
```

## In-place Conversion

Reuse structs to avoid allocations:

```go
var dsUser datastore.UserDatastore

for _, apiUser := range apiUsers {
    datastore.UserToUserDatastore(apiUser, &dsUser, nil)
    key := datastore.NameKey(dsUser.Kind(), apiUser.Email, nil)
    client.Put(ctx, key, &dsUser)
}
```

## Performance Tips

1. **Batch operations** instead of individual puts/gets
2. **Reuse destination structs** in loops
3. **Limit indexed fields** - use `noindex` for large text
4. **Use query projections** for partial entities:
```go
type UserEmail struct {
    Email string
}

var emails []UserEmail
query := datastore.NewQuery("User").Project("Email")
client.GetAll(ctx, query, &emails)
```

5. **Cache query results** when appropriate
6. **Use eventual consistency** for better performance:
```go
query := datastore.NewQuery("User").EventualConsistency()
```

## Limitations

- **No PropertyLoadSaver**: Keys managed manually, not via LoadKey/SaveKey interface
- **String IDs only for uint32**: Proto uint32 fields become string in Datastore
- **No automatic key generation**: Create keys explicitly before Put operations
- **No embedded entities**: All nested messages stored as entity properties

## Common Patterns

### Ancestor Queries

```go
parentKey := datastore.NameKey("Account", "acct-1", nil)

// Create child entity
key := datastore.NameKey("User", "alice", parentKey)
client.Put(ctx, key, dsUser)

// Query by ancestor
query := datastore.NewQuery("User").Ancestor(parentKey)
var users []datastore.UserDatastore
client.GetAll(ctx, query, &users)
```

### Composite Indexes

Define in `index.yaml`:

```yaml
indexes:
- kind: User
  properties:
  - name: Email
  - name: CreatedAt
    direction: desc
```

Query:
```go
query := datastore.NewQuery("User").
    Filter("Email =", "alice@example.com").
    Order("-CreatedAt")
```

### Large Text Fields

```protobuf
message ArticleDatastore {
  string content = 1 [(dal.v1.column) = {
    datastore_tags: ["noindex"]  // Don't index large text
  }];
}
```

## Troubleshooting

**Missing converter for nested type:**

Add `source` annotation:
```protobuf
message AuthorDatastore {
  option (dal.v1.table) = {
    target_datastore: {source: "api.v1.Author"}
  };
}
```

**String conversion panic:**

Check that uint32 fields are properly converted to strings in proto definition.

**Index too large error:**

Add `noindex` tag to large fields:
```protobuf
string large_field = 1 [(dal.v1.column) = {datastore_tags: ["noindex"]}];
```

**Namespace issues:**

Ensure namespace is set on keys:
```go
key.Namespace = "production"
```

## Examples

See `tests/protos/datastore/user.proto` for comprehensive examples including:
- Basic CRUD operations
- Nested messages
- Collections (repeated and map fields)
- Custom namespaces
- Large text fields with noindex
