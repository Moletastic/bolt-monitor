## Why

The current Lambda response shape is flat and inconsistent. `services/monitor-api/types.go` returns `{"error": "..."}` while `services/api-health` returns an ad-hoc success body. The dashboard parser in `apps/dashboard/lib/api.ts` matches the flat shape, so any new field is a breaking change. The constitution mandates a uniform envelope (`{ status, data?, reason?, message?, pagination? }`) for every entry point. This change introduces the envelope, the typed status enum, the pagination object, and retrofits existing Lambdas to use it. It is the foundation for the centralized error codes and the frontend Result refactor that follow.

## What Changes

- Add a shared Go module `shared/api/response` that defines the response envelope as a struct plus constructor helpers.
- Add a TypeScript class `ApiResponse<T, E>` in `apps/dashboard/lib/api-response.ts` mirroring the Go shape.
- Define `status` as a typed enum (`success | error`) in both runtimes.
- Define the pagination object as `{ page, size, total, items }` in both runtimes.
- Retrofit `services/api-health` and `services/monitor-api` handlers to return the envelope.
- Update `apps/dashboard/lib/api.ts` to parse the envelope shape.
- Add tests for the envelope constructors and the retrofit return paths.

## Capabilities

### New Capabilities

- `api-response-envelope`: shared Go module and TypeScript class that defines and constructs the uniform response envelope, including `status` enum, pagination object, and the success/failure variants.

### Modified Capabilities

- `api-health-endpoint`: response body changes from an ad-hoc shape to the envelope. `status: success`, `data: { status: "ok" }`. No `pagination`. No `reason`.
- `monitor-crud-api`: all CRUD, status, runs, and probe-locations endpoints return the envelope. Error responses carry `status: error`, `reason: { code, details }`. Paginated endpoints carry `pagination: { page, size, total, items }`.

## Impact

- New Go module `shared/api/response` with `envelope.go`, `status.go`, `pagination.go`, `errors.go`.
- New TypeScript file `apps/dashboard/lib/api-response.ts` with `ApiResponse`, `Status`, `Pagination`, `Reason` types.
- Modified `services/api-health/main.go` to return the envelope.
- Modified `services/monitor-api/handler.go` and `services/monitor-api/types.go` to remove `errorResponse` flat struct and emit envelope.
- Modified `apps/dashboard/lib/api.ts` to parse `status`, `data`, `reason`, `message`, `pagination`.
- Modified `apps/dashboard/lib/actions.ts` callers to read the new error shape (`reason.code` instead of `error`).
- New tests in `shared/api/response/*_test.go` and `apps/dashboard/lib/api-response.test.ts`.
- Dashboard `ApiError` consumers updated to surface `reason.code` and `reason.details` rather than a flat message.

## Out of Scope

- Centralized error code registry (covered by `go-error-handling-typed-codes` and `ts-result-and-no-any`).
- TypeScript `Result<T, E>` utility (covered by `ts-result-and-no-any`).
- Facade pattern, Rules pattern, date-fns (covered by `code-patterns-foundation`).
- FinOps tagging and right-sizing (deferred).
