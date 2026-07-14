.PHONY: \
	test-go test-go-all test-dashboard \
	format-go-check vet-go ci-go \
	lint-go lint-dashboard lint-infra lint-all \
	check-dashboard check-infra \
	format-dashboard format-dashboard-files format-infra format-infra-files \
	commitlint \
	build-go build-dashboard build-all \
	deploy-infra deploy-infra-print \
	bootstrap clean

GO_SERVICES := api-health check-runtime escalation-runtime monitor-api
GO_SHARED := api/response aws checkexecution dynamodb dynamodbrecord dynamodbschema errors escalation monitorconfig notifications resultstatus rules
GO_MODULE_DIRS := $(addprefix ./services/,$(GO_SERVICES)) $(addprefix ./shared/,$(GO_SHARED))

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

format-dashboard:
	cd apps/dashboard && pnpm run format

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

deploy-infra:
	cd infra && AWS_PROFILE=bolt-monitor pnpm exec sst deploy --stage staging

deploy-infra-print:
	cd infra && AWS_PROFILE=bolt-monitor pnpm exec sst deploy --stage staging --print-logs

clean:
	rm -f services/api-health/function.zip services/api-health/handler
	rm -f services/check-runtime/function.zip services/check-runtime/handler
	rm -f services/escalation-runtime/function.zip services/escalation-runtime/handler
	rm -f services/monitor-api/function.zip services/monitor-api/handler
