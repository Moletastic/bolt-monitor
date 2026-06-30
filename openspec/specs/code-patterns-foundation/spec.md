## Overview

Domain code talks to internal interfaces, not vendor SDKs. Business rules are composable predicates, not inline branches. The dashboard uses `date-fns` for every time operation; the native `Date` object is banned.

## Requirements

### Requirement: AWS access goes through a facade

Domain Go code SHALL depend on interfaces in `shared/aws` (`DynamoDBAPI`, `EventBridgeAPI`, `SQSAPI`), not on `aws-sdk-go-v2` types directly. Concrete wrappers translate between domain records and SDK shapes in one place. The infrastructure stack in `infra/stacks/bootstrap.ts` is the only place that constructs the SDK clients and injects them into the wrappers.

### Requirement: AWS facade covers all AWS touchpoints

`services/monitor-api/repository.go`, `services/check-runtime/sqs.go`, and `services/check-runtime/repository.go` SHALL NOT import `github.com/aws/aws-sdk-go-v2/service/dynamodb` (or `eventbridge` / `sqs`) types directly. They SHALL import `shared/aws` interfaces.

### Requirement: Rules are composable

Business rules SHALL be expressed as `Rule[T] func(T) error` and combined with combinators `All`, `Any`, `Not`, `When`, and a `Builder[T]` that aggregates failures with `errors.Join`. Inline `if` chains in domain code SHALL be replaced with rule chains when the chain is more than one check.

### Requirement: Field-scoped rules report the field name

Rules that validate a struct field SHALL use a `Field[T]` helper that wraps the failure with `details.field` set, so the response envelope's `reason.details` can pinpoint the offending field.

### Requirement: Monitor-config validation uses the rules package

`shared/monitorconfig/model.go` `Validate` SHALL be a rule chain built via the `Builder`. The exported signature SHALL be preserved. Failures SHALL surface as `*shared/errors.TypedError` with `code = VALIDATION_FAILED` and `details.field` per failing rule.

### Requirement: No native `Date` in the dashboard

The TypeScript / JavaScript `Date` object SHALL NOT be constructed or called directly in `apps/dashboard/**` except inside the explicit clock wrapper (`apps/dashboard/lib/clock.ts`) and the test setup. ESLint SHALL enforce this via `no-restricted-syntax`. `date-fns` SHALL be the only date utility used.

### Requirement: `date-fns` covers every time operation

`new Date(isoString)` -> `parseISO`. `date.toISOString()` -> `formatISO`. `Date.now()` -> wrapped via the clock helper. Date arithmetic (`addDays`, `differenceInMinutes`, etc.) -> `date-fns` functions. Manual `getTime()` / `setTime()` math SHALL be replaced with `date-fns` equivalents.

### Requirement: The clock wrapper is the single source of "now"

A `now()` helper in `apps/dashboard/lib/clock.ts` SHALL be the only way the dashboard reads the current wall-clock time. Tests SHALL inject a frozen clock. No direct `new Date()` in components or actions.

### Requirement: Facade and rules are testable without the SDK

Tests for code that uses the AWS facade SHALL use the `shared/aws` interface, not the SDK client. Tests for code that uses rules SHALL use plain `Rule[T]` values, not mocked validators. Concrete SDK behavior is covered by `shared/aws/*_test.go` contract tests against LocalStack or recorded fixtures, not in handler tests.
