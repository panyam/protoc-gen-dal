
buf:
	cd protos && buf generate

build:
	go build -o ./bin/protoc-gen-dal ./cmd/protoc-gen-dal
	go build -o ./bin/protoc-gen-dal-gorm ./cmd/protoc-gen-dal-gorm

test:
	cd tests ; make 
	go test ./... -v

