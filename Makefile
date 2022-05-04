.PHONY: ci
ci: tools test lint check_generate check_fmt check_tidy

.PHONY: check_generate
check_generate: generate fmt check_repository_unchanged

.PHONY: check_fmt
check_fmt: fmt check_repository_unchanged

.PHONY: check_tidy
check_tidy: tidy check_repository_unchanged

.PHONY: check_repository_unchanged
check_repository_unchanged: 
	_tools/check_repository_unchanged.sh

.PHONY: generate
generate:
	go generate ./...

.PHONY: fmt
fmt:
	gosimports -l -w ./

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run --skip-files "cmd/ssb-test/main.go" ./...

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
	go install github.com/rinchsan/gosimports/cmd/gosimports@latest # https://github.com/golang/go/issues/20818
