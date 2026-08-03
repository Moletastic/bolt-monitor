.DEFAULT_GOAL := help
SHELL := /bin/bash
.SHELLFLAGS := -euo pipefail -c
.DELETE_ON_ERROR:

.PHONY: \
	help \
	setup \
	bootstrap \
	test-go test-go-all test-dashboard \
	format-go-check vet-go ci-go \
	lint-go lint-dashboard lint-infra lint-all \
	check-dashboard check-infra test-infra \
	check-bruno check-api-contract test-api-contract check-auth-routes check-auth-cutover-prerequisites check-pnpm-install-trust check-pre-cutover-gate \
	format-dashboard format-dashboard-check format-dashboard-files format-infra format-infra-check format-infra-files \
	commitlint \
	build-go build-dashboard build-all \
	infra-status deploy-infra dev-infra remove-infra invite-admin setup-readiness readiness-api rotate-auth-key clean

GO_SERVICES := api-health check-runtime escalation-runtime monitor-api
GO_TOOLS := admin-bootstrap readiness-probe
GO_SHARED := api/response auth aws checkexecution dynamodb dynamodbrecord dynamodbschema errors escalation monitorconfig notifications resultstatus rules
GO_MODULE_DIRS := $(addprefix ./services/,$(GO_SERVICES)) $(addprefix ./tools/,$(GO_TOOLS)) $(addprefix ./shared/,$(GO_SHARED))

OPS_NODE_FLAGS := --experimental-strip-types --no-warnings
OPS_SCRIPT := infra/scripts/ops.mjs

help: ## Show documented Make targets
	@awk 'BEGIN { FS = ":.*## "; } /^[[:alnum:]_-]+:.*## / { printf "  %-36s %s\n", $$1, $$2; }' $(MAKEFILE_LIST)

setup: ## Install JavaScript dependencies and synchronize Go workspace
	pnpm --dir infra install --frozen-lockfile
	pnpm --dir apps/dashboard install --frozen-lockfile
	$(MAKE) bootstrap

bootstrap: ## Synchronize Go workspace modules
	go work sync

test-go: bootstrap ## Run Go tests for every workspace module
	$(foreach module,$(GO_MODULE_DIRS),go test $(module);)

test-go-all: test-go ## Run all Go tests

format-go-check: ## Check Go source formatting
	@files="$$(gofmt -l $(GO_MODULE_DIRS))"; \
	if [ -n "$$files" ]; then \
		printf 'gofmt needed:\n%s\n' "$$files"; \
		exit 1; \
	fi

vet-go: bootstrap ## Vet every Go workspace module
	$(foreach module,$(GO_MODULE_DIRS),go vet $(module);)

ci-go: format-go-check vet-go test-go-all ## Run Go format, vet, and test checks

test-dashboard: ## Run dashboard tests
	cd apps/dashboard && pnpm run test

lint-go: bootstrap ## Lint Go services and shared packages
	$(foreach svc,$(GO_SERVICES),golangci-lint run ./services/$(svc);)
	$(foreach lib,$(GO_SHARED),golangci-lint run ./shared/$(lib);)

lint-dashboard: ## Lint dashboard code
	cd apps/dashboard && pnpm run lint

lint-infra: ## Check infrastructure formatting
	cd infra && pnpm run format:check

lint-all: lint-go lint-dashboard lint-infra ## Run all lint checks

check-dashboard: ## Type-check dashboard code
	cd apps/dashboard && pnpm run typecheck

check-infra: ## Type-check infrastructure code
	cd infra && pnpm run check

test-infra: ## Run infrastructure and repository script tests
	cd infra && pnpm run test
	node --test scripts/check-auth-cutover-prerequisites.test.mjs scripts/check-pnpm-install-trust.test.mjs scripts/check-makefile-safety.test.mjs scripts/check-gitleaks-hook.test.mjs

check-bruno: ## Validate Bruno API collection coverage
	node scripts/check-bruno.mjs

check-auth-routes: ## Validate protected API route coverage
	node scripts/check-auth-routes.mjs

check-auth-cutover-prerequisites: ## Validate auth cutover lifecycle prerequisites
	node scripts/check-auth-cutover-prerequisites.mjs

check-pnpm-install-trust: ## Validate reviewed pnpm install-script allowlists
	node --test scripts/check-pnpm-install-trust.test.mjs
	node scripts/check-pnpm-install-trust.mjs

# Local release gates required before protected-route cutover. The dashboard build runs here once.
check-pre-cutover-gate: build-dashboard check-bruno check-api-contract check-auth-cutover-prerequisites ## Run local protected-route cutover gates

test-api-contract: ## Run API contract tests
	node --test scripts/check-api-contract.test.mjs scripts/check-bruno.test.mjs scripts/check-openapi-auth.test.mjs

check-api-contract: test-api-contract ## Validate API contract and OpenAPI authentication coverage
	node scripts/check-api-contract.mjs
	node scripts/check-openapi-auth.mjs

format-dashboard: ## Format dashboard files
	cd apps/dashboard && pnpm run format

format-dashboard-check: ## Check dashboard formatting
	cd apps/dashboard && pnpm run format:check

format-dashboard-files: ## Format dashboard FILES (whitespace-delimited; no spaces or single quotes)
	@if [ -n "$(FILES)" ]; then \
		set --; \
		for file in $(foreach file,$(FILES),'$(file)'); do \
			case "$$file" in \
				/*|apps/dashboard/*) set -- "$$@" "$(CURDIR)/$$file" ;; \
				*) set -- "$$@" "$(CURDIR)/apps/dashboard/$$file" ;; \
			esac; \
		done; \
		pnpm --dir apps/dashboard exec prettier --write "$$@"; \
	fi

format-infra: ## Format infrastructure files
	cd infra && pnpm run format

format-infra-check: ## Check infrastructure formatting
	cd infra && pnpm run format:check

format-infra-files: ## Format infrastructure FILES (whitespace-delimited; no spaces or single quotes)
	@if [ -n "$(FILES)" ]; then \
		set --; \
		for file in $(foreach file,$(FILES),'$(file)'); do \
			case "$$file" in \
				/*|infra/*) set -- "$$@" "$(CURDIR)/$$file" ;; \
				*) set -- "$$@" "$(CURDIR)/infra/$$file" ;; \
			esac; \
		done; \
		pnpm --dir infra exec prettier --write "$$@"; \
	fi

commitlint: ## Validate COMMIT_MSG_FILE or current commit message
	@if [ -z "$(COMMIT_MSG_FILE)" ]; then \
		pnpm exec commitlint --edit; \
	else \
		pnpm exec commitlint --edit "$(COMMIT_MSG_FILE)"; \
	fi

build-go: bootstrap ## Build and package Go Lambda handlers
	@for service in $(GO_SERVICES); do \
		GOOS=linux GOARCH=arm64 go build -o "services/$$service/handler" "./services/$$service"; \
		(cd "services/$$service" && zip function.zip handler); \
	done

build-dashboard: ## Build dashboard production bundle
	cd apps/dashboard && pnpm run build

build-all: build-go build-dashboard ## Build Go handlers and dashboard

infra-status: ## Show selected infrastructure target and AWS identity
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) status

deploy-infra: ## Deploy selected infrastructure target
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) deploy

dev-infra: ## Start selected infrastructure target in local development mode
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) dev

remove-infra: ## Remove selected target; persistent targets require DESTROY=yes
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) remove DESTROY=$(DESTROY)

invite-admin: ## Invite administrator with EMAIL=operator@example.com
	@if [ -z "$(EMAIL)" ]; then \
		printf '%s\n' 'EMAIL is required; usage: make invite-admin EMAIL=operator@example.com'; \
		exit 1; \
	fi
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) invite-admin EMAIL=$(EMAIL)

setup-readiness: ## Configure synthetic readiness user; use ROTATE=yes to rotate
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) setup-readiness EMAIL=$(EMAIL) ROTATE=$(ROTATE)

readiness-api: ## Verify protected API readiness; optional EMAIL=operator@example.com
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) readiness-api EMAIL=$(EMAIL)

rotate-auth-key: ## Rotate selected target authentication encryption key
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) rotate-auth-key

clean: ## Remove locally built Go Lambda artifacts
	@for service in $(GO_SERVICES); do \
		rm -f "services/$$service/function.zip" "services/$$service/handler"; \
	done
