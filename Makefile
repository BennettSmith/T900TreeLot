.DEFAULT_GOAL := help

override COVERAGE_MIN := 85
GO_FILES = $(shell git ls-files --cached --others --exclude-standard '*.go')
NODE_MODULES_STAMP = node_modules/.package-lock.json

.PHONY: help assets assets-watch assets-check showcase format format-check lint test coverage ci

help: ## List available targets and explain when to use them.
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[[:alnum:]_.-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

$(NODE_MODULES_STAMP): package.json package-lock.json
	@npm ci

assets: $(NODE_MODULES_STAMP) ## Build generated CSS for embedding in the Go binary.
	@npm run css:build

assets-watch: $(NODE_MODULES_STAMP) ## Rebuild CSS while templates and styles change.
	@npm run css:watch

assets-check: $(NODE_MODULES_STAMP) ## Verify generated CSS matches its source and templates.
	@output="$$(mktemp)"; \
	trap 'rm -f "$$output"' EXIT; \
	./node_modules/.bin/tailwindcss -i ./web/styles/app.css -o "$$output" --minify; \
	cmp -s "$$output" ./web/static/app.css || { \
		echo "web/static/app.css is stale; run 'make assets'." >&2; \
		exit 1; \
	}

showcase: assets ## Build, serve, and open the development design-system gallery.
	@url="http://localhost:$${PORT:-8080}/_dev/components"; \
	if curl --fail --silent "$$url" >/dev/null 2>&1; then \
		if command -v open >/dev/null 2>&1; then open "$$url"; \
		elif command -v xdg-open >/dev/null 2>&1; then xdg-open "$$url"; \
		else echo "Open $$url in a browser."; fi; \
		exit 0; \
	fi; \
	( deadline="$$(($$(date +%s) + 15))"; \
	  until curl --fail --silent "$$url" >/dev/null 2>&1; do \
		if test "$$(date +%s)" -ge "$$deadline"; then \
			echo "Showcase did not become ready at $$url." >&2; \
			exit 1; \
		fi; \
		sleep 0.1; \
	  done; \
	  if command -v open >/dev/null 2>&1; then open "$$url"; \
	  elif command -v xdg-open >/dev/null 2>&1; then xdg-open "$$url"; \
	  else echo "Open $$url in a browser."; fi \
	) & \
	APP_ENV=development go run ./cmd/web

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

ci: assets-check format-check lint coverage ## Run all required checks when evaluating whether work is done.
