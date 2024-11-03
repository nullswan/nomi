.PHONY: all
all: build-dev

.PHONY: fmt
fmt:
	golines . --write-output --max-len=80 --base-formatter="gofmt" --tab-len=2
	golangci-lint run --fix

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: build-dev
build-dev:
	@echo "Building..."
	@go build -ldflags="-X main.buildDate=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.buildVersion=$$(git rev-parse --short HEAD)-dev" -o dist/ ./...
	@echo "Done!"

.PHONY: dev
dev: build-dev
	@echo "Deploying..."
	@cp ./dist/cli ~/.local/bin/nomi
	@chmod +x ~/.local/bin/nomi
	@echo "Done!"
