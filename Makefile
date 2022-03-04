.PHONY: generate
generate:
	go generate ./...

.PHONY: fmt
fmt:
	goimports -l -w ./

.PHONY: test
test:
	go test ./...
