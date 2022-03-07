.PHONY: ci
ci: tools test lint check_generate check_fmt

.PHONY: check_generate
check_generate: generate fmt check_repository_unchanged

.PHONY: check_fmt
check_fmt: fmt check_repository_unchanged

.PHONY: check_repository_unchanged
check_repository_unchanged: 
	_tools/check_repository_unchanged.sh

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
	golangci-lint --skip-dirs "network/local" run ./... # todo: enable for all directories

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.2
