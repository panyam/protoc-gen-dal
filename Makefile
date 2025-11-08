
buf:
	cd proto && buf generate

test:
	go test ./... -v
