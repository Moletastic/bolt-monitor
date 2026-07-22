.PHONY: \
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
	infra-status deploy-infra dev-infra remove-infra invite-admin rotate-auth-key clean

GO_SERVICES := api-health check-runtime escalation-runtime monitor-api
GO_TOOLS := admin-bootstrap
GO_SHARED := api/response auth aws checkexecution dynamodb dynamodbrecord dynamodbschema errors escalation monitorconfig notifications resultstatus rules
GO_MODULE_DIRS := $(addprefix ./services/,$(GO_SERVICES)) $(addprefix ./tools/,$(GO_TOOLS)) $(addprefix ./shared/,$(GO_SHARED))

OPS_NODE_FLAGS := --experimental-strip-types --no-warnings
OPS_SCRIPT := infra/scripts/ops.mjs

setup:
	pnpm --dir infra install --frozen-lockfile
	pnpm --dir apps/dashboard install --frozen-lockfile
	$(MAKE) bootstrap

bootstrap:
	go work sync

test-go: bootstrap
	$(foreach module,$(GO_MODULE_DIRS),go test $(module);)

test-go-all: test-go

format-go-check:
	@files="$$(gofmt -l $(GO_MODULE_DIRS))"; \
	if [ -n "$$files" ]; then \
		printf 'gofmt needed:\n%s\n' "$$files"; \
		exit 1; \
	fi

vet-go: bootstrap
	$(foreach module,$(GO_MODULE_DIRS),go vet $(module);)

ci-go: format-go-check vet-go test-go-all

test-dashboard:
	cd apps/dashboard && pnpm run test

lint-go: bootstrap
	$(foreach svc,$(GO_SERVICES),golangci-lint run ./services/$(svc);)
	$(foreach lib,$(GO_SHARED),golangci-lint run ./shared/$(lib);)

lint-dashboard:
	cd apps/dashboard && pnpm run lint

lint-infra:
	cd infra && pnpm run format:check

lint-all: lint-go lint-dashboard lint-infra

check-dashboard:
	cd apps/dashboard && pnpm run typecheck

check-infra:
	cd infra && pnpm run check

test-infra:
	cd infra && pnpm run test
	node --test scripts/check-auth-cutover-prerequisites.test.mjs scripts/check-pnpm-install-trust.test.mjs

check-bruno:
	node scripts/check-bruno.mjs

check-auth-routes:
	node scripts/check-auth-routes.mjs

check-auth-cutover-prerequisites:
	node scripts/check-auth-cutover-prerequisites.mjs

check-pnpm-install-trust:
	node --test scripts/check-pnpm-install-trust.test.mjs
	node scripts/check-pnpm-install-trust.mjs

# Local release gates required before protected-route cutover. The dashboard build runs here once.
check-pre-cutover-gate: build-dashboard check-bruno check-api-contract check-auth-cutover-prerequisites

test-api-contract:
	node --test scripts/check-api-contract.test.mjs scripts/check-bruno.test.mjs scripts/check-openapi-auth.test.mjs

check-api-contract: test-api-contract
	node scripts/check-api-contract.mjs
	node scripts/check-openapi-auth.mjs

format-dashboard:
	cd apps/dashboard && pnpm run format

format-dashboard-check:
	cd apps/dashboard && pnpm run format:check

format-dashboard-files:
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

format-infra:
	cd infra && pnpm run format

format-infra-check:
	cd infra && pnpm run format:check

format-infra-files:
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

commitlint:
	@if [ -z "$(COMMIT_MSG_FILE)" ]; then \
		pnpm exec commitlint --edit; \
	else \
		pnpm exec commitlint --edit "$(COMMIT_MSG_FILE)"; \
	fi

build-go: bootstrap
	GOOS=linux GOARCH=arm64 go build -o services/api-health/handler ./services/api-health
	GOOS=linux GOARCH=arm64 go build -o services/check-runtime/handler ./services/check-runtime
	GOOS=linux GOARCH=arm64 go build -o services/escalation-runtime/handler ./services/escalation-runtime
	GOOS=linux GOARCH=arm64 go build -o services/monitor-api/handler ./services/monitor-api
	cd services/api-health && zip function.zip handler
	cd services/check-runtime && zip function.zip handler
	cd services/escalation-runtime && zip function.zip handler
	cd services/monitor-api && zip function.zip handler

build-dashboard:
	cd apps/dashboard && pnpm run build

build-all: build-go build-dashboard

infra-status:
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) status

deploy-infra:
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) deploy

dev-infra:
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) dev

remove-infra:
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) remove DESTROY=yes

invite-admin:
	@if [ -z "$(EMAIL)" ]; then \
		printf '%s\n' 'EMAIL is required; usage: make invite-admin EMAIL=operator@example.com'; \
		exit 1; \
	fi
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) invite-admin EMAIL=$(EMAIL)

rotate-auth-key:
	node $(OPS_NODE_FLAGS) $(OPS_SCRIPT) rotate-auth-key

clean:
	rm -f services/api-health/function.zip services/api-health/handler
	rm -f services/check-runtime/function.zip services/check-runtime/handler
	rm -f services/escalation-runtime/function.zip services/escalation-runtime/handler
	rm -f services/monitor-api/function.zip services/monitor-api/handler
