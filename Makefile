
buf:
	cd proto && buf generate

build:
	go build -o ./bin/protoc-gen-dal ./cmd/protoc-gen-dal

test:
	go test ./... -v

