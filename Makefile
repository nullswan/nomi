.PHONY: build
build:
	@echo "Building..."
	@goreleaser build --snapshot --clean

.PHONY: build-dev
build-dev:
	@echo "Building..."
	@go build -o dist/ ./...

.PHONY: fmt
fmt:
	golines . --write-output --max-len=80 --base-formatter="gofmt" --tab-len=2
	golangci-lint run --fix

.PHONY: test
test:
	go test -v -cover ./...