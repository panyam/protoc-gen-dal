# Annotation Reference

Complete reference for protoc-gen-dal annotations defined in `dal/v1/annotations.proto`.

## Table-level Annotations

Applied at the message level using `option (dal.v1.table)`.

### TableOptions

Container for target-specific table options.

```protobuf
message UserGORM {
  option (dal.v1.table) = {
    source: "api.v1.User"
    name: "users"
    target_datastore: {...}
  };
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | Fully-qualified source API message name |
| `name` | string | No | Table/collection name (defaults to lowercase message name) |
| `target_gorm` | GormOptions | No | GORM-specific options |
| `target_datastore` | DatastoreOptions | No | Datastore-specific options |
| `target_postgres` | PostgresOptions | No | Postgres-specific options (future) |
| `target_firestore` | FirestoreOptions | No | Firestore-specific options (future) |

### GormOptions

GORM-specific table options.

```protobuf
option (dal.v1.table) = {
  target_gorm: {
    source: "api.v1.User"
    table_name: "users"
  }
};
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | Source API message name |
| `table_name` | string | No | Database table name |

**Note:** Usually specify `source` and `name` at TableOptions level rather than target_gorm.

### DatastoreOptions

Google Cloud Datastore options.

```protobuf
option (dal.v1.table) = {
  target_datastore: {
    source: "api.v1.User"
    kind: "User"
    namespace: "production"
  }
};
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | Source API message name |
| `kind` | string | No | Datastore kind (defaults to message name) |
| `namespace` | string | No | Datastore namespace |

### PostgresOptions

Postgres-specific options (planned).

```protobuf
option (dal.v1.table) = {
  target_postgres: {
    source: "api.v1.User"
    schema: "public"
    table_name: "users"
  }
};
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | Source API message name |
| `schema` | string | No | Postgres schema |
| `table_name` | string | No | Table name |

### FirestoreOptions

Google Cloud Firestore options (planned).

```protobuf
option (dal.v1.table) = {
  target_firestore: {
    source: "api.v1.User"
    collection: "users"
  }
};
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | Source API message name |
| `collection` | string | No | Firestore collection name |

## Field-level Annotations

Applied to individual fields using `[(dal.v1.column) = {...}]`.

### ColumnOptions

Container for target-specific column options.

```protobuf
uint32 id = 1 [(dal.v1.column) = {
  gorm_tags: ["primaryKey", "autoIncrement"]
  datastore_tags: ["-"]
}];
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `gorm_tags` | repeated string | GORM struct tags |
| `datastore_tags` | repeated string | Datastore struct tags |
| `postgres_tags` | repeated string | Postgres-specific tags (future) |
| `firestore_tags` | repeated string | Firestore-specific tags (future) |

### GORM Tags

Tags passed directly to GORM struct tag. See [GORM documentation](https://gorm.io/docs/models.html#Fields-Tags) for complete list.

**Common tags:**

| Tag | Example | Description |
|-----|---------|-------------|
| `primaryKey` | `gorm_tags: ["primaryKey"]` | Mark as primary key |
| `autoIncrement` | `gorm_tags: ["autoIncrement"]` | Auto-increment field |
| `unique` | `gorm_tags: ["unique"]` | Unique constraint |
| `uniqueIndex` | `gorm_tags: ["uniqueIndex:idx_email"]` | Named unique index |
| `index` | `gorm_tags: ["index:idx_name"]` | Named index |
| `not null` | `gorm_tags: ["not null"]` | NOT NULL constraint |
| `default` | `gorm_tags: ["default:active"]` | Default value |
| `type` | `gorm_tags: ["type:varchar(255)"]` | Column type |
| `size` | `gorm_tags: ["size:255"]` | Column size |
| `autoCreateTime` | `gorm_tags: ["autoCreateTime"]` | Auto-set on create |
| `autoUpdateTime` | `gorm_tags: ["autoUpdateTime"]` | Auto-update on save |
| `foreignKey` | `gorm_tags: ["foreignKey:AuthorID"]` | Foreign key field |
| `references` | `gorm_tags: ["references:ID"]` | Referenced field |
| `constraint` | `gorm_tags: ["constraint:OnDelete:CASCADE"]` | Foreign key constraint |
| `embedded` | `gorm_tags: ["embedded"]` | Embed struct |
| `embeddedPrefix` | `gorm_tags: ["embeddedPrefix:author_"]` | Prefix for embedded fields |
| `-` | `gorm_tags: ["-"]` | Ignore field |

**Multiple tags:**
```protobuf
string email = 1 [(dal.v1.column) = {
  gorm_tags: ["uniqueIndex", "not null", "type:varchar(255)"]
}];
```

**Database-specific types:**

Postgres:
```protobuf
gorm_tags: ["type:uuid"]
gorm_tags: ["type:jsonb"]
gorm_tags: ["type:text[]"]
gorm_tags: ["type:inet"]
```

MySQL:
```protobuf
gorm_tags: ["type:json"]
gorm_tags: ["type:longtext"]
```

SQLite:
```protobuf
gorm_tags: ["type:datetime"]
```

### Datastore Tags

Tags passed to Datastore struct tag. See [Datastore documentation](https://pkg.go.dev/cloud.google.com/go/datastore) for details.

**Common tags:**

| Tag | Example | Description |
|-----|---------|-------------|
| `noindex` | `datastore_tags: ["noindex"]` | Don't index field |
| `omitempty` | `datastore_tags: ["omitempty"]` | Omit if empty |
| `-` | `datastore_tags: ["-"]` | Ignore field (used for Key) |

**Examples:**

```protobuf
// Don't index large text
string bio = 1 [(dal.v1.column) = {
  datastore_tags: ["noindex"]
}];

// Exclude from storage (Key field)
string key = 2 [(dal.v1.column) = {
  datastore_tags: ["-"]
}];

// Multiple tags
string metadata = 3 [(dal.v1.column) = {
  datastore_tags: ["noindex", "omitempty"]
}];
```

## Custom Converter Annotations

Override default conversion behavior for specific fields.

### ConverterFunc

Specify custom converter function names (planned).

```protobuf
message User {
  google.protobuf.Timestamp birthday = 1 [(dal.v1.converter) = {
    to_func: "birthdayToDate"
    from_func: "dateFromBirthday"
  }];
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `to_func` | string | Function name for API → Target conversion |
| `from_func` | string | Function name for Target → API conversion |

## Usage Patterns

### Basic GORM Entity

```protobuf
message UserGORM {
  option (dal.v1.table) = {
    source: "api.v1.User"
    name: "users"
  };

  uint32 id = 1 [(dal.v1.column) = {
    gorm_tags: ["primaryKey", "autoIncrement"]
  }];

  string name = 2 [(dal.v1.column) = {
    gorm_tags: ["not null"]
  }];

  string email = 3 [(dal.v1.column) = {
    gorm_tags: ["uniqueIndex", "not null"]
  }];
}
```

### Basic Datastore Entity

```protobuf
message UserDatastore {
  option (dal.v1.table) = {
    target_datastore: {
      source: "api.v1.User"
      kind: "User"
      namespace: "production"
    }
  };

  string key = 1 [(dal.v1.column) = {
    datastore_tags: ["-"]
  }];

  string id = 2;
  string name = 3;

  string email = 4 [(dal.v1.column) = {
    datastore_tags: ["noindex"]
  }];
}
```

### Foreign Key Relationship

```protobuf
message BookGORM {
  uint32 author_id = 1 [(dal.v1.column) = {
    gorm_tags: [
      "foreignKey:AuthorID",
      "references:ID",
      "constraint:OnDelete:CASCADE,OnUpdate:CASCADE"
    ]
  }];

  AuthorGORM author = 2 [(dal.v1.column) = {
    gorm_tags: ["foreignKey:AuthorID"]
  }];
}
```

### Composite Primary Key

```protobuf
message BookEditionGORM {
  string book_id = 1 [(dal.v1.column) = {
    gorm_tags: ["primaryKey"]
  }];

  int32 edition = 2 [(dal.v1.column) = {
    gorm_tags: ["primaryKey"]
  }];
}
```

### Indexes

```protobuf
message UserGORM {
  // Single field index
  string email = 1 [(dal.v1.column) = {
    gorm_tags: ["index:idx_email"]
  }];

  // Composite index
  string first_name = 2 [(dal.v1.column) = {
    gorm_tags: ["index:idx_name,priority:1"]
  }];

  string last_name = 3 [(dal.v1.column) = {
    gorm_tags: ["index:idx_name,priority:2"]
  }];

  // Unique index
  string username = 4 [(dal.v1.column) = {
    gorm_tags: ["uniqueIndex:idx_username"]
  }];
}
```

### Timestamps

```protobuf
message UserGORM {
  int64 created_at = 1 [(dal.v1.column) = {
    gorm_tags: ["autoCreateTime"]
  }];

  int64 updated_at = 2 [(dal.v1.column) = {
    gorm_tags: ["autoUpdateTime"]
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

### JSONB Fields

```protobuf
message UserGORM {
  map<string, string> metadata = 1 [(dal.v1.column) = {
    gorm_tags: ["type:jsonb"]
  }];

  repeated string tags = 2 [(dal.v1.column) = {
    gorm_tags: ["type:jsonb"]
  }];
}
```

### Postgres Arrays

```protobuf
message ProductGORM {
  repeated string tags = 1 [(dal.v1.column) = {
    gorm_tags: ["type:text[]"]
  }];

  repeated int32 ratings = 2 [(dal.v1.column) = {
    gorm_tags: ["type:integer[]"]
  }];
}
```

## Validation

The generator validates:
- `source` field is required for all table options
- `source` references a valid proto message
- Tag syntax is not validated (passed through to target)

Invalid configurations produce warnings during code generation.

## Extension

To add support for new targets:

1. Add target-specific options to TableOptions:
```protobuf
message TableOptions {
  optional MyTargetOptions target_mytarget = 10;
}
```

2. Define target-specific message:
```protobuf
message MyTargetOptions {
  string source = 1;
  string custom_field = 2;
}
```

3. Add field-level tags to ColumnOptions:
```protobuf
message ColumnOptions {
  repeated string mytarget_tags = 10;
}
```

4. Implement generator that reads these options.

See `protos/dal/v1/annotations.proto` for complete definitions.
