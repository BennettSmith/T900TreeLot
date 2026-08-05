.DEFAULT_GOAL := help

override COVERAGE_MIN := 85
GO_FILES = $(shell git ls-files --cached --others --exclude-standard '*.go')
NODE_MODULES_STAMP = node_modules/.package-lock.json
DOCKER ?= $(shell if docker info >/dev/null 2>&1; then echo docker; else echo sudo docker; fi)
COMPOSE = $(DOCKER) compose
IMAGE ?= treelot:local

.PHONY: help doctor acceptance-preflight install-hooks commit-messages assets assets-watch assets-check \
	showcase format format-check lint test-db test coverage traceability ci image up down migrate logs ps acceptance \
	render-setup-checklist

help: ## List available targets and explain when to use them.
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[[:alnum:]_.-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

doctor: ## Diagnose local tools, Git hooks, Docker, networking, and required ports.
	@bash ./scripts/preflight.sh doctor

acceptance-preflight: ## Check acceptance prerequisites and required ports without cleanup.
	@bash ./scripts/preflight.sh acceptance

install-hooks: ## Configure Git to use the repository's tracked local hooks.
	@git config --local core.hooksPath .githooks
	@echo "Git hooks installed from .githooks."

commit-messages: ## Validate commit subjects from BASE_SHA through HEAD_SHA.
	@test -n "$(BASE_SHA)" || { echo "BASE_SHA is required." >&2; exit 2; }
	@test -n "$(HEAD_SHA)" || { echo "HEAD_SHA is required." >&2; exit 2; }
	@bash ./scripts/check-commit-messages.sh "$(BASE_SHA)" "$(HEAD_SHA)"

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

test-db: ## Resolve disposable treelot_test Postgres (host :5432 or Compose :5433).
	@chmod +x ./scripts/ensure-test-db.sh
	@./scripts/ensure-test-db.sh >/dev/null

test: ## Run all Go unit/component tests; use during development.
	@chmod +x ./scripts/ensure-test-db.sh
	@TEST_DATABASE_URL="$$(./scripts/ensure-test-db.sh)" go test ./...

coverage: ## Run tests and require at least 85% statement coverage.
	@chmod +x ./scripts/ensure-test-db.sh
	@profile="$$(mktemp)"; \
	trap 'rm -f "$$profile"' EXIT; \
	packages="$$(go list ./internal/... ./web/... | grep -Ev '/testdb$$')"; \
	coverpkg="$$(echo "$$packages" | paste -sd, -)"; \
	TEST_DATABASE_URL="$$(./scripts/ensure-test-db.sh)" \
		go test ./... -count=1 -covermode=atomic -coverpkg="$$coverpkg" -coverprofile="$$profile"; \
	total="$$(go tool cover -func="$$profile" | awk '/^total:/ {gsub(/%/, "", $$3); print $$3}')"; \
	test -n "$$total" || { echo "Unable to determine total coverage." >&2; exit 1; }; \
	awk -v total="$$total" -v minimum="$(COVERAGE_MIN)" 'BEGIN { \
		if (total + 0 < minimum + 0) { \
			printf "Coverage %.1f%% is below the required %.1f%%.\n", total, minimum > "/dev/stderr"; \
			exit 1; \
		} \
		printf "Coverage %.1f%% meets the required %.1f%%.\n", total, minimum; \
	}'

traceability: ## Validate requirement revisions, evidence, and generated report.
	@go run ./cmd/traceability check

ci: traceability assets-check format-check lint coverage ## Run fast required checks (no Docker acceptance).

image: ## Build the immutable production image used by Compose and acceptance.
	@$(DOCKER) build -t "$(IMAGE)" .

up: image ## Start the local Docker Compose development stack.
	@$(COMPOSE) --profile dev up --build -d postgres
	@$(COMPOSE) run --rm migrate
	@$(COMPOSE) --profile dev up -d web worker

down: ## Stop the local Docker Compose stack. Set DOWN_FLAGS=-v to remove volumes.
	@$(COMPOSE) --profile dev --profile acceptance down $(DOWN_FLAGS)

migrate: ## Apply database migrations through the migrate entry point.
	@$(COMPOSE) up -d postgres
	@$(COMPOSE) run --rm migrate

logs: ## Tail logs for postgres, web, and worker.
	@$(COMPOSE) --profile dev logs -f postgres web worker

ps: ## Show Docker Compose service status.
	@$(COMPOSE) --profile dev --profile acceptance ps

acceptance: ## Build the production image and run foundation acceptance specs.
	@chmod +x ./scripts/acceptance.sh
	@./scripts/acceptance.sh

render-setup-checklist: ## Print the EW-001 Render first-time operator checklist.
	@chmod +x ./scripts/render-setup-checklist.sh
	@./scripts/render-setup-checklist.sh
