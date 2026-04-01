module github.com/panyam/protoc-gen-dal

go 1.24.6

require (
	github.com/panyam/protokit v0.0.1
	google.golang.org/protobuf v1.36.11
)

// Local development
replace github.com/panyam/protokit => ../../protokit/main
