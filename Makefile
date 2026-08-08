.PHONY: help tidy t test fmt lint tools

GO_PKGS   := $(shell go list -f {{.Dir}} ./...)

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z\._-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

tidy: ## Run go mod tidy in all directories
	go mod tidy

t: test
test: ## Run unit tests, alias: t
	go test --cover -timeout=300s -parallel=16 $(GO_PKGS)
	cd tests && go test --cover -timeout=300s -parallel=16 ./...

tools:
	go install golang.org/x/vuln/cmd/govulncheck@latest

fmt: tidy
	golangci-lint fmt

lint: fmt
	golangci-lint custom -v
	./custom-gcl  run --fix ./...
