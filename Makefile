.PHONY: help test lint lint-all lint-fix vulncheck mocks

GOLANGCI_LINT_VERSION := v2.15.0
GOVULNCHECK_VERSION   := v1.9.0
GOLANGCI_LINT         := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

help:
	@echo "Available targets:"
	@echo "  make test       - Run tests with the race detector"
	@echo "  make lint       - Run golangci-lint (new issues since origin/main)"
	@echo "  make lint-all   - Run golangci-lint on all code"
	@echo "  make lint-fix   - Apply golangci-lint formatters and autofixes"
	@echo "  make vulncheck  - Check dependencies for known vulnerabilities"

test:
	go test -race ./...

lint:
	$(GOLANGCI_LINT) run --new-from-merge-base=origin/main ./...

lint-all:
	$(GOLANGCI_LINT) run ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...
