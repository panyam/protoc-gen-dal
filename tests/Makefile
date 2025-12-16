# Default: production mode using new split generators

all: buf

buf: ensureenv clean
	buf dep update
	buf generate
	goimports -w `find gen | grep "\.go"`
	cd gen/gorm && go build -o /dev/null .
	cd gen/gorm/dal && go build -o /dev/null .
	cd gen/datastore && go build -o /dev/null .
	cd gen/datastore/dal && go build -o /dev/null .
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
testds:
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
	go test ./tests/datastore/...

.PHONY: run setupdev setupprod cleanall dev updb downdb dblogs testpg upds downds dslogs testds
