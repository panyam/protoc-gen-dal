# protoc-gen-dal Tests

This directory contains integration tests for the `protoc-gen-dal` code generators. The tests verify that the generated GORM models and Google Cloud Datastore entities work correctly with their respective backends.

## Directory Structure

```
tests/
├── protos/              # Proto definitions for testing
│   ├── api/             # Base proto definitions
│   ├── datastore/       # Datastore-specific proto configs
│   ├── gorm/            # GORM-specific proto configs
│   └── weewar/          # Example complex proto definitions
├── gen/                 # Generated code (created by buf generate)
│   ├── datastore/       # Generated Datastore entities
│   │   └── dal/         # Generated Datastore DAL helpers
│   ├── gorm/            # Generated GORM models
│   │   └── dal/         # Generated GORM DAL helpers (if enabled)
│   └── go/              # Generated Go protobuf code
├── tests/               # Test implementations
│   ├── datastore/       # Datastore integration tests
│   └── gorm/            # GORM integration tests
├── Makefile             # Test runner and development commands
└── buf.gen.yaml         # buf code generation configuration
```

## Prerequisites

- Go 1.21+
- Docker (for database emulators)
- buf CLI (`brew install bufbuild/buf/buf` or see https://buf.build/docs/installation)

## Quick Start

### Development Setup

For development (uses local proto definitions via symlink):

```bash
make setupdev
```

### Production Setup

For production mode (uses published proto definitions):

```bash
make setupprod
```

### Running Tests

Run all tests (requires no external services for basic tests):

```bash
make buf
```

This command:
1. Updates buf dependencies
2. Generates code from protos
3. Formats generated Go code
4. Builds the generated packages
5. Runs all tests

## Database-Specific Tests

### SQLite (Default)

GORM tests use SQLite by default, requiring no additional setup:

```bash
go test ./tests/gorm/...
```

### PostgreSQL

To run tests with PostgreSQL:

1. Start the PostgreSQL container:
   ```bash
   make updb
   ```

2. Run tests with PostgreSQL:
   ```bash
   make testpg
   ```

3. Stop the container when done:
   ```bash
   make downdb
   ```

View PostgreSQL logs:
```bash
make dblogs
```

### Google Cloud Datastore Emulator

To run tests with the Datastore emulator:

1. Start the Datastore emulator:
   ```bash
   make upds
   ```

2. Run Datastore tests:
   ```bash
   make testds
   ```

3. Stop the emulator when done:
   ```bash
   make downds
   ```

View emulator logs:
```bash
make dslogs
```

### Real Google Cloud Datastore

To run tests against a real Google Cloud Datastore (not the emulator):

1. Configure your GCP project and credentials:
   ```bash
   # Required: GCP project ID
   export DS_REAL_PROJECT=my-gcp-project

   # Optional: Path to service account JSON (uses ADC if not set)
   export DS_REAL_CREDENTIALS=~/path/to/service-account.json

   # Optional: Test namespace (default: protoc-gen-dal-test)
   export DS_REAL_NAMESPACE=my-test-namespace
   ```

2. Run the tests:
   ```bash
   make testrealDS
   ```

   Or pass variables directly:
   ```bash
   make testrealDS DS_REAL_PROJECT=my-project DS_REAL_CREDENTIALS=~/creds.json
   ```

**Safety Check**: Tests will fail if the namespace already contains entities for the test kinds. This prevents accidental data loss. To override:

```bash
# Force delete existing entities in the namespace
make testrealDS-force DS_REAL_PROJECT=my-project
```

Or run directly with the flag:
```bash
DATASTORE_PROJECT_ID=my-project \
DATASTORE_TEST_NAMESPACE=my-namespace \
go test ./tests/datastore/... -args -force-delete-ns
```

**Cleanup**: All entities in the test namespace are automatically deleted after tests complete.

## Test Categories

### GORM Tests (`tests/gorm/`)

Tests for GORM model generation:

- **sqlite_test.go**: Basic CRUD operations with SQLite
- **user_converters_test.go**: User model conversion (proto <-> GORM)
- **document_converters_test.go**: Document model with nested types
- **weewar_converters_test.go**: Complex game models with embedded types
- **testany_converters_test.go**: google.protobuf.Any field handling

### Datastore Tests (`tests/datastore/`)

Tests for Datastore entity generation:

- **datastore_test.go**: Test helpers and client setup
- **user_dal_test.go**: DAL pattern tests (Put, Get, Delete, Query, etc.)

Note: Datastore tests use a custom `TestUser` struct that only uses Datastore-compatible types (signed integers, strings, etc.). The generated entities may have `uint32` fields which are not supported by Datastore.

## Environment Variables

### PostgreSQL

| Variable | Description | Default |
|----------|-------------|---------|
| `PROTOC_GEN_DAL_TEST_PGDB` | Database name | `testdb` |
| `PROTOC_GEN_DAL_TEST_PGPORT` | PostgreSQL port | `5433` |
| `PROTOC_GEN_DAL_TEST_PGUSER` | Username | `postgres` |
| `PROTOC_GEN_DAL_TEST_PGPASSWORD` | Password | `testpassword` |

### Datastore (Emulator)

| Variable | Description | Default |
|----------|-------------|---------|
| `DATASTORE_EMULATOR_HOST` | Emulator host (e.g., `localhost:8081`) | - |
| `DATASTORE_PROJECT_ID` | GCP project ID | `test-project` |

### Datastore (Real GCP)

| Variable | Description | Default |
|----------|-------------|---------|
| `DATASTORE_PROJECT_ID` | GCP project ID | Required |
| `DATASTORE_CREDENTIALS_FILE` | Path to service account JSON | Uses ADC |
| `DATASTORE_TEST_NAMESPACE` | Namespace for test entities | Required |

Test flags:
| Flag | Description |
|------|-------------|
| `-force-delete-ns` | Force delete existing entities in namespace |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make buf` | Full build and test cycle |
| `make clean` | Remove generated code and lock file |
| `make cleanall` | Clean and remove buf.yaml |
| `make parent` | Build parent project |
| `make setupdev` | Setup for development with symlinks |
| `make setupprod` | Setup for production mode |
| `make updb` | Start PostgreSQL container |
| `make downdb` | Stop PostgreSQL container |
| `make dblogs` | Tail PostgreSQL logs |
| `make testpg` | Run tests with PostgreSQL |
| `make upds` | Start Datastore emulator |
| `make downds` | Stop Datastore emulator |
| `make dslogs` | Tail Datastore emulator logs |
| `make testds` | Run tests with Datastore emulator |
| `make testrealDS` | Run tests with real GCP Datastore |
| `make testrealDS-force` | Run tests with real GCP Datastore (force delete existing) |

## Adding New Tests

### Adding a New Proto

1. Add your `.proto` file to `protos/api/` (base definition)
2. Add GORM-specific annotations in `protos/gorm/` (optional)
3. Add Datastore-specific annotations in `protos/datastore/` (optional)
4. Run `make buf` to generate code
5. Add tests in `tests/gorm/` or `tests/datastore/`

### Test Patterns

GORM tests typically follow this pattern:
```go
func TestXxx(t *testing.T) {
    db := setupTestDB(t) // SQLite in-memory
    // ... test CRUD operations
}
```

Datastore tests typically follow this pattern:
```go
func TestXxx(t *testing.T) {
    client := setupTestClient(t) // Requires emulator
    ctx := context.Background()
    cleanupKind(ctx, client, "YourKind")
    // ... test operations
}
```

## Troubleshooting

### "buf.yaml does not exist"

Run `make setupdev` or `make setupprod` to create the buf.yaml symlink.

### Tests skipped with "DATASTORE_EMULATOR_HOST not set"

Start the Datastore emulator with `make upds` or use `make testds` which handles this automatically.

### "datastore: unsupported struct field type: uint32"

Datastore only supports signed integers. If you encounter this error, your entity struct may have unsigned integer fields. Create a test-specific struct using only supported types:
- Signed integers: `int`, `int8`, `int16`, `int32`, `int64`
- `bool`, `string`, `float32`, `float64`
- `[]byte`, `*datastore.Key`, `time.Time`
- Structs and slices of the above

### "datastore: unsupported struct field type: map[string]int64"

Datastore doesn't natively support Go map types. To store maps, enable the `implement_property_loader` option in your proto:

```protobuf
message MyEntityDatastore {
  option (dal.v1.datastore_options) = {
    source: "api.MyEntity"
    kind: "MyEntity"
    implement_property_loader: true  // Enables PropertyLoadSaver interface
  };

  map<string, int64> my_map = 1;
}
```

This generates `Save()` and `Load()` methods that serialize map fields to JSON. See `docs/DATASTORE_GUIDE.md` for details.

### Docker container fails to start

Ensure Docker is running and no conflicting containers exist:
```bash
docker ps -a | grep protoc-gen-dal
docker rm -f <container_id>
```
