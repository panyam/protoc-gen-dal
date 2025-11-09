# Default: production mode using new split generators

all: buf

buf: ensureenv clean
	buf dep update
	buf generate
	goimports -w `find gen | grep "\.go"`
	cd gen/gorm && go build -o /dev/null .
	go test -v

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

.PHONY: run setupdev setupprod cleanall dev
