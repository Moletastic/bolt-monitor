## 1. shared/aws facade

- [x] 1.1 Create `shared/aws/dynamo.go` with `DynamoDBAPI` interface covering `GetItem`, `PutItem`, `UpdateItem`, `DeleteItem`, `Query`, `Scan` — typed against the domain record shapes, not raw `map[string]types.AttributeValue`.
- [x] 1.2 Add `NewDynamoDB(client *dynamodb.Client) DynamoDBAPI` returning the concrete wrapper. The wrapper translates between domain records and `AttributeValue` once, in one place.
- [x] 1.3 Create `shared/aws/eventbridge.go` with `EventBridgeAPI` interface covering `PutEvents`. Wrapper handles `PutEventsRequestEntry` shape.
- [x] 1.4 Create `shared/aws/sqs.go` with `SQSAPI` interface covering `SendMessage`, `ReceiveMessage`, `DeleteMessage`, `ChangeMessageVisibility`. Wrapper handles `types.Message` and `SendMessageInput` shape.
- [x] 1.5 Wire `shared/aws` into `go.work`.
- [x] 1.6 Migrate `services/monitor-api/repository.go` to depend on `DynamoDBAPI`. Remove direct `*dynamodb.Client` usage from the file.
- [x] 1.7 Migrate `services/check-runtime/sqs.go` and `services/check-runtime/repository.go` to depend on `SQSAPI` and `DynamoDBAPI`.
- [x] 1.8 Update `infra/stacks/bootstrap.ts` to wire the concrete `*dynamodb.Client`, `*eventbridge.Client`, `*sqs.Client` into the Lambda constructors. Existing SST wiring injects resources/env; Go entrypoints construct concrete clients through facade constructors.
- [x] 1.9 Add `shared/aws/*_test.go` contract tests: assert the wrapper translates domain records correctly (use LocalStack or recorded fixtures; the test must not hit real AWS in CI).

## 2. shared/rules

- [x] 2.1 Create `shared/rules/rules.go` with combinator types: `Rule[T] func(T) error`, `All[T](rules ...Rule[T]) Rule[T]`, `Any[T](rules ...Rule[T]) Rule[T]`, `Not[T](r Rule[T]) Rule[T]`, `When[T](pred func(T) bool, r Rule[T]) Rule[T]`.
- [x] 2.2 Add a `Builder[T]` type with `Add(r Rule[T])` and `Build() Rule[T]`. Returns a single rule that aggregates failures via `errors.Join`.
- [x] 2.3 Add `Field[T]` helper that scopes a rule to a struct field, returning a typed error with `details.field` set.
- [x] 2.4 Wire `shared/rules` into `go.work`.
- [x] 2.5 Refactor `shared/monitorconfig/model.go:208-312` validation: express each existing check as a `Field[MonitorConfig](m, "interval", ...)` rule. Keep the exported `Validate` signature unchanged.
- [x] 2.6 Update `services/monitor-api/handler.go` to consume the new `Validate` and map rule failures to `VALIDATION_FAILED` with `details.field` (per `go-error-handling-typed-codes`). Existing `respondAPIGateway` already routes typed validation errors through `errors.Respond`.

## 3. date-fns in dashboard

- [x] 3.1 Add `date-fns` to `apps/dashboard/package.json`. Pin to a major version. Run `pnpm install --frozen-lockfile`.
- [x] 3.2 Audit `apps/dashboard/**` for `new Date(` and `Date.now(` usages. Replace with `date-fns` equivalents:
  - `new Date()` (current time) → use a small `now()` wrapper from `date-fns` or inject a clock.
  - `new Date(isoString)` → `parseISO(isoString)`.
  - `date.toISOString()` → `formatISO(date)`.
  - `date.getTime()` / `Date.now()` → `Date.now()` is allowed only for the once-per-render wall-clock snapshot; prefer `new Date()` replacement via a clock.
- [x] 3.3 Add an ESLint rule banning native `Date` constructor in dashboard code: `no-restricted-syntax` for `NewExpression[callee.name='Date']` and `CallExpression[callee.name='Date']` with the message "Use date-fns instead of native Date".
- [x] 3.4 Allow a narrow exception for the clock wrapper itself (`apps/dashboard/lib/clock.ts`) and the existing setup files. Document the exception.
- [x] 3.5 Run `pnpm lint`, fix violations, run `make check-dashboard`, `make build-dashboard`. All pass.

## 4. Tests

- [x] 4.1 Add `shared/rules/rules_test.go` covering combinator short-circuit behavior, `errors.Join` aggregation, and `Field` scoping.
- [x] 4.2 Add a test in `shared/monitorconfig/model_test.go` asserting the refactored `Validate` returns the same errors for the same inputs (golden test).
- [x] 4.3 Add a dashboard test asserting the `no-restricted-syntax` rule fires on a sample `new Date()` snippet (lint-test fixture).
- [x] 4.4 Run `make test-go-all`, `make lint-go`, `make lint-dashboard`, `make check-dashboard`. All pass.

## 5. Documentation

- [x] 5.1 Update `AGENTS.md` Go section with facade pattern example: domain code depends on `shared/aws.DynamoDBAPI`, not the SDK client.
- [x] 5.2 Update `AGENTS.md` Go section with rules pattern example: one validation block using `Field` and `Builder`.
- [x] 5.3 Update `AGENTS.md` TypeScript section with the date-fns adoption note and the clock-wrapper convention.
- [x] 5.4 Cross-reference `CONSTITUTION.md` §13 (facade), §20 (rules), §22 (no native Date) to this change.
