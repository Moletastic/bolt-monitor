## Why

Three constitution principles remain unimplemented. (1) Facade pattern: domain code in `shared/*` and `services/*` imports AWS SDK types directly, which couples business logic to vendor shapes and makes mocking painful. (2) Rules pattern in Go: business rules are scattered as inline `if` branches across handlers and validators, not composable. (3) No native `Date`: the dashboard uses `new Date()` and `Date.now()` in forms, toasts, and timeline components, making time handling non-deterministic across timezones and refactor-hostile. This change introduces a thin AWS facade in Go, a composable rules package, and adopts `date-fns` (and a lint rule) in the dashboard to retire the native `Date` object.

## What Changes

- Add a Go facade package `shared/aws` defining interfaces for DynamoDB, EventBridge, and SQS that the rest of the domain talks to. Concrete implementations wrap `aws-sdk-go-v2` clients.
- Migrate `services/monitor-api/repository.go` and `services/check-runtime/sqs.go` to depend on the facade interfaces, not the SDK.
- Add a Go `shared/rules` package with composable predicates: `All`, `Any`, `Not`, `When`, plus a per-domain rule builder.
- Refactor monitor-config validation in `shared/monitorconfig/model.go:208-312` to use the rules package.
- Add `date-fns` to `apps/dashboard/package.json`.
- Replace `new Date()` / `Date.now()` usage in `apps/dashboard/**` with `date-fns` equivalents.
- Add an ESLint rule (or convention + grep guard in CI) banning native `Date` constructor in dashboard code.

## Capabilities

### New Capabilities

- `code-patterns-foundation`: AWS facade interfaces in Go, the rules pattern package in Go, and `date-fns` adoption + no-native-Date enforcement in the dashboard.

### Modified Capabilities

- `monitor-crud-api`: repository depends on `shared/aws.DynamoDBAPI` interface, not the concrete `dynamodb.Client`. Easier to mock in tests; tests no longer need `dynamodb.Client` doubles.
- `check-execution-pipeline`: SQS and EventBridge dispatches go through `shared/aws` facade interfaces.
- `monitor-configuration`: validation rules expressed as composable predicates in `shared/rules`.

## Impact

- New Go module `shared/aws` with `dynamo.go`, `eventbridge.go`, `sqs.go`, each defining an interface and a constructor that returns the concrete implementation backed by `aws-sdk-go-v2`.
- Modified `services/monitor-api/repository.go` to depend on the facade interface.
- Modified `services/check-runtime/sqs.go` and `services/check-runtime/repository.go` similarly.
- New Go module `shared/rules` with `rules.go` (combinators) and a `Builder` type for declarative rule chains.
- Refactored `shared/monitorconfig/model.go` validation to use the rules package. The exported `Validate` function keeps its current signature; the body is a rule chain.
- Modified `apps/dashboard/package.json` to add `date-fns` and remove direct native `Date` calls in `apps/dashboard/**`.
- New ESLint config / `no-restricted-syntax` rule banning `NewExpression[callee.object.name='Date']` and `CallExpression[callee.object.name='Date']` in dashboard code.
- Tests in `shared/aws/*_test.go` (contract tests against the real SDK via a recorded fixture or LocalStack), `shared/rules/rules_test.go` (combinator behavior), and dashboard tests covering `date-fns` usage.

## Out of Scope

- Response envelope (covered by `api-response-envelope`).
- Go error-handling typed codes (covered by `go-error-handling-typed-codes`).
- TypeScript `Result` and no-`any` (covered by `ts-result-and-no-any`).
- Backend FinOps tagging and right-sizing (deferred).
