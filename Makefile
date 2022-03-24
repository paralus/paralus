.PHONY: tidy
tidy:
	GOPRIVATE=github.com/RafayLabs/* go mod tidy
.PHONY: vendor
vendor:
	go mod vendor

.PHONY: build-proto
build-proto:
	buf build

.PHONY: gen-proto
gen-proto:
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
	rm -rf ./**/gen
	find . -name "*.pb*" -type f -delete
