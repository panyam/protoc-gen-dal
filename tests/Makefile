# Default: production mode using new split generators

all: buf

buf: ensureenv clean
	buf dep update
	buf generate
	goimports -w `find gen | grep "\.go"`
	# Directory structure is preserved: gorm/*.proto -> gen/gorm/gorm/*, datastore/*.proto -> gen/datastore/datastore/*
	cd gen/gorm/gorm && go build -o /dev/null .
	cd gen/gorm/dal/gorm && go build -o /dev/null .
	cd gen/datastore/datastore && go build -o /dev/null .
	cd gen/datastore/dal/datastore && go build -o /dev/null .
	go test ./...

ensureenv:
	@test -f buf.yaml && echo "buf.yaml does not exist.  Run 'make setupdev' or 'make setupprod' to setup your environment..."

cleanall: clean
	rm -f buf.yaml

clean:
	rm -Rf gen
	rm -f buf.lock

parent:
	cd ../ && make build

setupdev: cleanall symlink-dal
	ln -s buf.yaml.dev buf.yaml

setupprod: cleanall remove-symlink
	ln -s buf.yaml.prod buf.yaml

# Create symlink to dal annotations for development
symlink-dal:
	@if [ ! -L protos/dal ]; then \
		echo "Creating dal symlink for development..."; \
		ln -s ../../protos/dal protos/dal; \
	fi

# Remove symlink (for switching back to production mode)
remove-symlink:
	@if [ -L protos/dal ]; then \
		echo "Removing dal symlink..."; \
		rm protos/dal ; \
	fi

# PostgreSQL test database configuration
PG_CONTAINER_NAME := protoc-gen-dal-test-pg
PG_PORT := 5433
PG_USER := postgres
PG_PASSWORD := testpassword
PG_DB := testdb

# Start an ephemeral PostgreSQL instance using Docker for testing
updb:
	@echo "Starting PostgreSQL container..."
	@docker run --rm -d \
		--name $(PG_CONTAINER_NAME) \
		-e POSTGRES_USER=$(PG_USER) \
		-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
		-e POSTGRES_DB=$(PG_DB) \
		-p $(PG_PORT):5432 \
    arm64v8/postgres:18.1
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo ""
	@echo "PostgreSQL is running!"
	@echo "To run tests with PostgreSQL, use:"
	@echo "  export PROTOC_GEN_DAL_TEST_PGDB=$(PG_DB)"
	@echo "  export PROTOC_GEN_DAL_TEST_PGPORT=$(PG_PORT)"
	@echo "  export PROTOC_GEN_DAL_TEST_PGUSER=$(PG_USER)"
	@echo "  export PROTOC_GEN_DAL_TEST_PGPASSWORD=$(PG_PASSWORD)"
	@echo "  make buf"
	@echo ""
	@echo "Or run: make testpg"
	@echo ""
	@echo "To stop: make downdb"

# Stop the PostgreSQL test container
downdb:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(PG_CONTAINER_NAME) 2>/dev/null || echo "Container not running"

# Tail the logs of the running PostgreSQL container
dblogs:
	@docker logs -f $(PG_CONTAINER_NAME)

# Run tests with PostgreSQL (starts container if not running)
testpg:
	@if ! docker ps --format '{{.Names}}' | grep -q '^$(PG_CONTAINER_NAME)$$'; then \
		echo "Starting PostgreSQL container..."; \
		docker run --rm -d \
			--name $(PG_CONTAINER_NAME) \
			-e POSTGRES_USER=$(PG_USER) \
			-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
			-e POSTGRES_DB=$(PG_DB) \
			-p $(PG_PORT):5432 \
			postgres:15-alpine; \
		sleep 3; \
	fi
	PROTOC_GEN_DAL_TEST_PGDB=$(PG_DB) \
	PROTOC_GEN_DAL_TEST_PGPORT=$(PG_PORT) \
	PROTOC_GEN_DAL_TEST_PGUSER=$(PG_USER) \
	PROTOC_GEN_DAL_TEST_PGPASSWORD=$(PG_PASSWORD) \
	go test ./...

# Datastore emulator configuration
DS_CONTAINER_NAME := protoc-gen-dal-test-datastore
DS_PORT := 8081
DS_PROJECT := test-project

# Start a Datastore emulator using Docker for testing
upds:
	@echo "Starting Datastore emulator container..."
	@docker run --rm -d \
		--name $(DS_CONTAINER_NAME) \
		-p $(DS_PORT):8081 \
		gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators \
		gcloud beta emulators datastore start \
			--host-port=0.0.0.0:8081 \
			--project=$(DS_PROJECT) \
			--no-store-on-disk
	@echo "Waiting for Datastore emulator to be ready..."
	@sleep 3
	@echo ""
	@echo "Datastore emulator is running!"
	@echo "To run tests with Datastore emulator, use:"
	@echo "  export DATASTORE_EMULATOR_HOST=localhost:$(DS_PORT)"
	@echo "  export DATASTORE_PROJECT_ID=$(DS_PROJECT)"
	@echo "  go test ./tests/datastore/..."
	@echo ""
	@echo "Or run: make testds"
	@echo ""
	@echo "To stop: make downds"

# Stop the Datastore emulator container
downds:
	@echo "Stopping Datastore emulator container..."
	@docker stop $(DS_CONTAINER_NAME) 2>/dev/null || echo "Container not running"

# Tail the logs of the running Datastore emulator container
dslogs:
	@docker logs -f $(DS_CONTAINER_NAME)

# Run tests with Datastore emulator (starts container if not running)
testds: downds
	@if ! docker ps --format '{{.Names}}' | grep -q '^$(DS_CONTAINER_NAME)$$'; then \
		echo "Starting Datastore emulator container..."; \
		docker run --rm -d \
			--name $(DS_CONTAINER_NAME) \
			-p $(DS_PORT):8081 \
			gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators \
			gcloud beta emulators datastore start \
				--host-port=0.0.0.0:8081 \
				--project=$(DS_PROJECT) \
				--no-store-on-disk; \
		sleep 5; \
	fi
	DATASTORE_EMULATOR_HOST=localhost:$(DS_PORT) \
	DATASTORE_PROJECT_ID=$(DS_PROJECT) \
	go test -v ./tests/datastore/...

# Real Datastore configuration
# Override via command line: make testrealDS DS_REAL_PROJECT=other-project
DS_REAL_PROJECT ?= gappeng
DS_REAL_CREDENTIALS ?= ~/dev-app-data/secrets/gappeng/gappeng-7bb71377bfa2.json
DS_REAL_NAMESPACE ?= gendal

# Run tests against real Google Cloud Datastore
# This requires:
#   - DS_REAL_PROJECT: GCP project ID
#   - DS_REAL_CREDENTIALS: Path to service account JSON (optional, uses ADC if not set)
#   - DS_REAL_NAMESPACE: Namespace for test entities (default: protoc-gen-dal-test)
testrealDS:
	@if [ -z "$(DS_REAL_PROJECT)" ]; then \
		echo "Error: DS_REAL_PROJECT must be set to your GCP project ID"; \
		echo "Usage: make testrealDS DS_REAL_PROJECT=my-project DS_REAL_CREDENTIALS=~/path/to/creds.json"; \
		exit 1; \
	fi
	@echo "Running tests against real Datastore..."
	@echo "  Project: $(DS_REAL_PROJECT)"
	@echo "  Namespace: $(DS_REAL_NAMESPACE)"
	@if [ -n "$(DS_REAL_CREDENTIALS)" ]; then \
		echo "  Credentials: $(DS_REAL_CREDENTIALS)"; \
	else \
		echo "  Credentials: Application Default Credentials (ADC)"; \
	fi
	@echo ""
	DATASTORE_PROJECT_ID=$(DS_REAL_PROJECT) \
	DATASTORE_CREDENTIALS_FILE=$(DS_REAL_CREDENTIALS) \
	DATASTORE_TEST_NAMESPACE=$(DS_REAL_NAMESPACE) \
	go test -v ./tests/datastore/...

# Run tests against real Datastore with force-delete-ns flag
# This will delete any existing entities in the test namespace before running tests
testrealDS-force:
	@if [ -z "$(DS_REAL_PROJECT)" ]; then \
		echo "Error: DS_REAL_PROJECT must be set to your GCP project ID"; \
		echo "Usage: make testrealDS-force DS_REAL_PROJECT=my-project DS_REAL_CREDENTIALS=~/path/to/creds.json"; \
		exit 1; \
	fi
	@echo "Running tests against real Datastore (with force delete)..."
	@echo "  Project: $(DS_REAL_PROJECT)"
	@echo "  Namespace: $(DS_REAL_NAMESPACE)"
	@if [ -n "$(DS_REAL_CREDENTIALS)" ]; then \
		echo "  Credentials: $(DS_REAL_CREDENTIALS)"; \
	else \
		echo "  Credentials: Application Default Credentials (ADC)"; \
	fi
	@echo "  Force delete: enabled"
	@echo ""
	DATASTORE_PROJECT_ID=$(DS_REAL_PROJECT) \
	DATASTORE_CREDENTIALS_FILE=$(DS_REAL_CREDENTIALS) \
	DATASTORE_TEST_NAMESPACE=$(DS_REAL_NAMESPACE) \
	go test -v ./tests/datastore/... -args -force-delete-ns

.PHONY: run setupdev setupprod cleanall dev updb downdb dblogs testpg upds downds dslogs testds testrealDS testrealDS-force
