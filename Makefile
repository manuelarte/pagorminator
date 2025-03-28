default: help

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z\._-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help

tidy: ## Run go mod tidy in all directories
	go mod tidy
.PHONY: tidy

t: test
test: ## Run unit tests, alias: t
	go test --cover -timeout=300s -parallel=16 ${TEST_DIRECTORIES}
.PHONY: t test

fmt: format-code
format-code: tidy ## Format go code and run the fixer, alias: fmt
	golangci-lint fmt
	golangci-lint run --fix ./...
	funcorder ./...
.PHONY: fmt format-code

tools:
	go install github.com/manuelarte/funcorder@latest
.PHONY: tools