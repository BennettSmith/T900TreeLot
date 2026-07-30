.DEFAULT_GOAL := help

override COVERAGE_MIN := 85
GO_FILES = $(shell git ls-files --cached --others --exclude-standard '*.go')

.PHONY: help format format-check lint test coverage ci

help: ## List available targets and explain when to use them.
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[[:alnum:]_.-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

format: ## Format all Go files; use before committing code.
	@test -n "$(GO_FILES)" || { echo "No Go files found." >&2; exit 1; }
	@gofmt -w $(GO_FILES)

format-check: ## Check Go formatting without changing files; use in CI.
	@test -n "$(GO_FILES)" || { echo "No Go files found." >&2; exit 1; }
	@unformatted="$$(gofmt -l $(GO_FILES))"; \
	if test -n "$$unformatted"; then \
		echo "The following Go files need formatting:" >&2; \
		echo "$$unformatted" >&2; \
		exit 1; \
	fi

lint: ## Run Go static analysis; use before submitting changes.
	@go vet ./...

test: ## Run all Go tests; use during development.
	@go test ./...

coverage: ## Run tests and require at least 85% statement coverage.
	@profile="$$(mktemp)"; \
	trap 'rm -f "$$profile"' EXIT; \
	go test ./... -covermode=atomic -coverprofile="$$profile"; \
	total="$$(go tool cover -func="$$profile" | awk '/^total:/ {gsub(/%/, "", $$3); print $$3}')"; \
	test -n "$$total" || { echo "Unable to determine total coverage." >&2; exit 1; }; \
	awk -v total="$$total" -v minimum="$(COVERAGE_MIN)" 'BEGIN { \
		if (total + 0 < minimum + 0) { \
			printf "Coverage %.1f%% is below the required %.1f%%.\n", total, minimum > "/dev/stderr"; \
			exit 1; \
		} \
		printf "Coverage %.1f%% meets the required %.1f%%.\n", total, minimum; \
	}'

ci: format-check lint coverage ## Run all required checks when evaluating whether work is done.
