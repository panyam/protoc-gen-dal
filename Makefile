
all: buf build test install

buf:
	cd protos && buf generate

build:
	go build -o ./bin/protoc-gen-dal ./cmd/protoc-gen-dal
	go build -o ./bin/protoc-gen-dal-gorm ./cmd/protoc-gen-dal-gorm
	go build -o ./bin/protoc-gen-dal-datastore ./cmd/protoc-gen-dal-datastore

install:
	go build -o ${GOBIN}/protoc-gen-dal ./cmd/protoc-gen-dal
	go build -o ${GOBIN}/protoc-gen-dal-gorm ./cmd/protoc-gen-dal-gorm
	go build -o ${GOBIN}/protoc-gen-dal-datastore ./cmd/protoc-gen-dal-datastore

test:
	clear
	go test ./... -v
	cd tests ; make 

