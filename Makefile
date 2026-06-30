.PHONY: \
	test-go test-go-all test-dashboard \
	lint-go lint-dashboard lint-infra lint-all \
	check-dashboard check-infra \
	format-dashboard format-infra \
	build-go build-dashboard build-all \
	deploy-infra deploy-infra-print \
	bootstrap clean

GO_SERVICES := api-health check-runtime escalation-runtime monitor-api
GO_SHARED := aws checkexecution dynamodbschema errors monitorconfig notifications probelocationcatalog resultstatus rules

bootstrap:
	go work sync

test-go: bootstrap
	$(foreach svc,$(GO_SERVICES),go test ./services/$(svc);)
	$(foreach lib,$(GO_SHARED),go test ./shared/$(lib);)

test-go-all: bootstrap
	go test ./services/api-health
	go test ./services/check-runtime
	go test ./services/escalation-runtime
	go test ./services/monitor-api
	go test ./shared/aws
	go test ./shared/checkexecution
	go test ./shared/dynamodbschema
	go test ./shared/errors
	go test ./shared/monitorconfig
	go test ./shared/notifications
	go test ./shared/probelocationcatalog
	go test ./shared/resultstatus
	go test ./shared/rules

test-dashboard:
	cd apps/dashboard && npm run test

lint-go: bootstrap
	$(foreach svc,$(GO_SERVICES),golangci-lint run ./services/$(svc);)
	$(foreach lib,$(GO_SHARED),golangci-lint run ./shared/$(lib);)

lint-dashboard:
	cd apps/dashboard && npm run lint

lint-infra:
	cd infra && npm run format:check

lint-all: lint-go lint-dashboard lint-infra

check-dashboard:
	cd apps/dashboard && npm run typecheck

check-infra:
	cd infra && npm run check

format-dashboard:
	cd apps/dashboard && npm run format

format-infra:
	cd infra && npm run format

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
	cd apps/dashboard && npm run build

build-all: build-go build-dashboard

deploy-infra:
	cd infra && AWS_PROFILE=mole npx sst deploy --stage staging

deploy-infra-print:
	cd infra && AWS_PROFILE=mole npx sst deploy --stage staging --print-logs

clean:
	rm -f services/api-health/function.zip services/api-health/handler
	rm -f services/check-runtime/function.zip services/check-runtime/handler
	rm -f services/escalation-runtime/function.zip services/escalation-runtime/handler
	rm -f services/monitor-api/function.zip services/monitor-api/handler
