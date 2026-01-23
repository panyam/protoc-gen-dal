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
import "google/protobuf/timestamp.proto";

message UserDatastore {
  option (dal.v1.datastore_options) = {
    source: "api.v1.User"
    kind: "User"
    namespace: "production"
  };

  // ID stored in Datastore Key, excluded from properties
  string id = 1 [(dal.v1.column) = {
    datastore_tags: ["-"]
  }];

  string name = 2;

  // Large text fields should not be indexed
  string email = 3 [(dal.v1.column) = {
    datastore_tags: ["noindex"]
  }];

  // Timestamps map to time.Time automatically
  google.protobuf.Timestamp created_at = 4;
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

Specify namespace in datastore_options:

```protobuf
message UserDatastore {
  option (dal.v1.datastore_options) = {
    source: "api.v1.User"
    kind: "User"
    namespace: "production"  // Optional namespace
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

## Map Fields (PropertyLoadSaver)

Google Cloud Datastore doesn't natively support Go map types like `map[string]int64`. To store map fields, use the `implement_property_loader` option which generates the PropertyLoadSaver interface:

```protobuf
message ReactionCountsDatastore {
  option (dal.v1.datastore_options) = {
    source: "api.v1.ReactionCounts"
    kind: "ReactionCounts"
    implement_property_loader: true  // Enable PropertyLoadSaver generation
  };

  string entity_id = 1;
  int64 total_count = 2;

  // Map field - will be serialized to JSON
  map<string, int64> counts_by_type = 3 [(dal.v1.column) = {
    datastore_tags: ["noindex"]  // Maps are typically not indexed
  }];
}
```

This generates `Save()` and `Load()` methods:

```go
type ReactionCountsDatastore struct {
    Key          *datastore.Key   `datastore:"-"`
    EntityId     string           `datastore:"entity_id"`
    TotalCount   int64            `datastore:"total_count"`
    CountsByType map[string]int64 `datastore:"counts_by_type,noindex"`
}

// Save implements PropertyLoadSaver - serializes maps to JSON
func (m *ReactionCountsDatastore) Save() ([]datastore.Property, error) { ... }

// Load implements PropertyLoadSaver - deserializes JSON back to maps
func (m *ReactionCountsDatastore) Load(props []datastore.Property) error { ... }
```

**How it works:**
1. `Save()` extracts non-map fields using a temporary struct, then JSON-encodes map fields as `[]byte` properties
2. `Load()` separates map properties, loads non-map fields, then JSON-decodes map properties back

**Usage:**

```go
counts := &dsgen.ReactionCountsDatastore{
    EntityId:   "post-123",
    TotalCount: 5,
    CountsByType: map[string]int64{
        "like": 3,
        "love": 2,
    },
}

key := datastore.NameKey(counts.Kind(), counts.EntityId, nil)
_, err := client.Put(ctx, key, counts)  // PropertyLoadSaver.Save() called automatically

var fetched dsgen.ReactionCountsDatastore
err = client.Get(ctx, key, &fetched)     // PropertyLoadSaver.Load() called automatically
// fetched.CountsByType == map[string]int64{"like": 3, "love": 2}
```

**Supported map types:**
- `map[string]int64`, `map[string]int32`, `map[string]string`, etc.
- Maps with any scalar value type
- Maps with message value types (nested messages also JSON-serialized)

## Limitations

- **String IDs only for uint32**: Proto uint32 fields become string in Datastore
- **No automatic key generation**: Create keys explicitly before Put operations
- **No embedded entities**: All nested messages stored as entity properties
- **Map fields require PropertyLoadSaver**: Use `implement_property_loader: true` for structs with map fields

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

## DAL Helpers (Data Access Layer)

The generator can optionally produce DAL helper structs that simplify common database operations. Enable DAL generation with the `dal=true` option:

### Configuration

```yaml
version: v2
plugins:
  - local: protoc-gen-dal-datastore
    out: gen/datastore
    opt:
      - paths=source_relative
      - dal=true
      - dal_out=gen/datastore/dal  # Optional: separate directory for DAL files
```

### Generated DAL Structure

For each entity, a DAL struct is generated:

```go
type UserDatastoreDAL struct {
    // Kind overrides the Datastore kind for all operations.
    // If empty, uses the struct's Kind() method.
    Kind string

    // Namespace overrides the Datastore namespace for all operations.
    Namespace string

    // WillPut hook is called before Put operations.
    // Return an error to prevent the put.
    WillPut func(context.Context, *datastore.UserDatastore) error
}

func NewUserDatastoreDAL(kind string) *UserDatastoreDAL
```

### DAL Methods

Each DAL provides these methods:

| Method | Description |
|--------|-------------|
| `Put(ctx, client, obj)` | Save entity, returns key |
| `Get(ctx, client, key)` | Get by key, returns (nil, nil) if not found |
| `GetByID(ctx, client, id)` | Get by string ID |
| `Delete(ctx, client, key)` | Delete by key |
| `DeleteByID(ctx, client, id)` | Delete by string ID |
| `PutMulti(ctx, client, objs)` | Batch save |
| `GetMulti(ctx, client, keys)` | Batch get |
| `GetMultiByIDs(ctx, client, ids)` | Batch get by IDs |
| `DeleteMulti(ctx, client, keys)` | Batch delete |
| `Query(ctx, client, q)` | Execute query |
| `Count(ctx, client, q)` | Count query results |

### Basic Usage

```go
import (
    "context"
    "cloud.google.com/go/datastore"

    dsgen "yourproject/gen/datastore"
    "yourproject/gen/datastore/dal"
)

func main() {
    ctx := context.Background()
    client, _ := datastore.NewClient(ctx, "project-id")

    // Create DAL instance
    userDAL := dal.NewUserDatastoreDAL("User")

    // Create entity
    user := &dsgen.UserDatastore{
        Id:    "alice",
        Name:  "Alice",
        Email: "alice@example.com",
    }

    // Save - key is determined automatically from Id field
    key, err := userDAL.Put(ctx, client, user)

    // Get by ID
    fetched, err := userDAL.GetByID(ctx, client, "alice")

    // Delete by ID
    err = userDAL.DeleteByID(ctx, client, "alice")
}
```

### Key Determination

The DAL determines keys in this order:
1. If `entity.Key` is set, uses that key
2. If `entity.Id` is set, creates a name key from the ID
3. Otherwise, creates an incomplete key (auto-generated numeric ID)

```go
// Explicit key
user.Key = datastore.NameKey("User", "alice", nil)
userDAL.Put(ctx, client, user)

// Key from Id field
user.Id = "alice"
userDAL.Put(ctx, client, user)  // Creates NameKey("User", "alice", nil)

// Auto-generated key
user.Id = ""
user.Key = nil
userDAL.Put(ctx, client, user)  // Creates IncompleteKey("User", nil)
```

### Namespace Override

Set a namespace on the DAL to apply it to all operations:

```go
userDAL := dal.NewUserDatastoreDAL("User")
userDAL.Namespace = "production"

// All operations now use the "production" namespace
key, _ := userDAL.Put(ctx, client, user)
// key.Namespace == "production"
```

### WillPut Hook

The WillPut hook allows pre-save validation or transformation:

```go
userDAL.WillPut = func(ctx context.Context, u *dsgen.UserDatastore) error {
    // Validate
    if u.Name == "" {
        return errors.New("name is required")
    }

    // Transform
    u.Email = strings.ToLower(u.Email)
    u.UpdatedAt = time.Now().Unix()

    return nil
}

// Hook is called before Put and PutMulti
userDAL.Put(ctx, client, user)
```

### Batch Operations

```go
// Batch save
users := []*dsgen.UserDatastore{
    {Id: "user-1", Name: "Alice"},
    {Id: "user-2", Name: "Bob"},
    {Id: "user-3", Name: "Charlie"},
}
keys, err := userDAL.PutMulti(ctx, client, users)

// Batch get by IDs
ids := []string{"user-1", "user-2", "user-3"}
fetched, err := userDAL.GetMultiByIDs(ctx, client, ids)
// fetched[i] is nil if not found

// Batch delete
err = userDAL.DeleteMulti(ctx, client, keys)
```

### Query Support

```go
// Query all users
q := datastore.NewQuery("User")
users, err := userDAL.Query(ctx, client, q)

// Query with filter
q := datastore.NewQuery("User").FilterField("name", ">=", "B")
users, err := userDAL.Query(ctx, client, q)

// Count
q := datastore.NewQuery("User")
count, err := userDAL.Count(ctx, client, q)
```

### Handling Missing Entities

`Get` and `GetByID` return `(nil, nil)` when an entity is not found:

```go
user, err := userDAL.GetByID(ctx, client, "does-not-exist")
if err != nil {
    // Handle actual error
    return err
}
if user == nil {
    // Entity not found
    return ErrNotFound
}
```

`GetMulti` and `GetMultiByIDs` return `nil` for missing entities at their index:

```go
ids := []string{"exists", "does-not-exist", "exists-2"}
users, err := userDAL.GetMultiByIDs(ctx, client, ids)
// users[0] != nil
// users[1] == nil (not found)
// users[2] != nil
```

## Type Limitations

Datastore only supports these Go types:
- Signed integers: `int`, `int8`, `int16`, `int32`, `int64`
- `bool`, `string`, `float32`, `float64`
- `[]byte`, `*datastore.Key`, `time.Time`
- Structs and slices of the above

**Unsigned integers (`uint32`, `uint64`) are NOT supported.** If your proto uses `uint32`, convert to `int32` or `string` in your Datastore sidecar proto.

## Examples

See `tests/protos/datastore/user.proto` for comprehensive examples including:
- Basic CRUD operations
- Nested messages
- Collections (repeated and map fields)
- Custom namespaces
- Large text fields with noindex
- DAL helper integration
