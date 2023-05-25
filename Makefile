.PHONY: tidy
tidy:
	go mod tidy
.PHONY: vendor
vendor:
	go mod vendor

.PHONY: build
build:
	# Omit the symbol table and debug information to reduce the
	# size of binary.
	go build -ldflags "-s" -o paralus .

.PHONY: clean-proto
clean-proto:
	rm -rf ./gen
	find . -name "*.pb*" -type f -delete

.PHONY: build-proto
build-proto: clean-proto
	buf build
	buf generate

.PHONY: test
test:
	go test ./...

.PHONY: check
check:
	go fmt ./...

	go vet ./...

.PHONY: clean
clean:
	rm paralus

## changelog: generate changelog (latest release)
.PHONY: changelog
changelog:
	conventional-changelog -i CHANGELOG.md -s -p conventionalcommits
