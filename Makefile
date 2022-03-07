.PHONY: ci
ci: tools test lint

.PHONY: generate
generate:
	go generate ./...

.PHONY: fmt
fmt:
	goimports -l -w ./

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run ./...

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.2
