.PHONY: tidy
tidy:
	GOPRIVATE=github.com/RafaySystems/* go mod tidy
.PHONY: vendor
vendor:
	go mod vendor

.PHONY: build-proto
build-proto:
	cd components; buf build

.PHONY: gen-proto
gen-proto:
	cd components/common; buf generate
	cd components/adminsrv; buf generate

.PHONY: check
check:
	go fmt ./...
	go vet ./...

.PHONY: clean
clean:
	rm -rf components/**/gen
	find . -name "*.pb*" -type f -delete