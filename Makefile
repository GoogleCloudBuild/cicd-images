TIMEOUT_UNIT = 20m

GOFLAGS_TEST ?= -v -cover

all: help

.PHONY: install-tools
install-tools: ## install dev tools
	@echo "Installing dev tools"
	@cd tools && cat tools.go | grep '_' | awk -F '"' '{print $$2}' | xargs -t go install

.PHONY: lint
lint: ## runs go linter on all go files
	@echo "Linting go files..."
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## runs gofmt on all go files
	@echo "Formatting go files..."
	@go fmt ./...

.PHONY: test-unit
test-unit: ## runs unit tests
	@echo "Running go unit tests"
	@go test $(GOFLAGS_TEST) ./...

.PHONY: pre-commit
pre-commit: ## run pre-commit hooks script manually
	@pre-commit run --verbose --all-files --hook-stage pre-push

.PHONY: check-licenses
check-licenses: ## run licensing check and create licensings if needed
	@./license_check.sh

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
