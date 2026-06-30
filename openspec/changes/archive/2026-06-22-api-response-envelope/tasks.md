## 1. Shared Go response module

- [x] 1.1 Create `shared/api/response/envelope.go` with `Envelope[T any]` struct: `Status Status`, `Data *T`, `Reason *Reason`, `Message *string`, `Pagination *Pagination`.
- [x] 1.2 Add constructor `Ok[T any](data T, message ...string) Envelope[T]` and `Err[T any](code string, details map[string]any) Envelope[T]`.
- [x] 1.3 Add constructor `OkPaginated[T any](data T, page, size, total int) Envelope[T]` that builds the pagination object.
- [x] 1.4 Create `shared/api/response/status.go` with typed `Status` enum (`StatusSuccess`, `StatusError`) and string serialization (`"success"`, `"error"`).
- [x] 1.5 Create `shared/api/response/pagination.go` with `Pagination{ Page, Size, Total, Items any }` and JSON tags.
- [x] 1.6 Create `shared/api/response/reason.go` with `Reason{ Code string, Details map[string]any }`.
- [x] 1.7 Add `MarshalJSON` to `Envelope[T]` that omits nil optional fields and serializes `Status` as a string.
- [x] 1.8 Wire the new module into `go.work` (run `make bootstrap`).

## 2. TypeScript response class

- [x] 2.1 Create `apps/dashboard/lib/api-response.ts` with `ApiResponse<T>` interface matching the Go envelope: `status: 'success' | 'error'`, `data?: T`, `reason?: { code: string; details: Record<string, unknown> }`, `message?: string`, `pagination?: { page: number; size: number; total: number; items: unknown[] }`.
- [x] 2.2 Export `Status` enum: `Status.Success = 'success'`, `Status.Error = 'error'`.
- [x] 2.3 Add helper `isSuccess<T>(r: ApiResponse<T>): r is ApiResponse<T> & { data: T }` and `isError<T>(r: ApiResponse<T>): r is ApiResponse<T> & { reason: { code: string; details: Record<string, unknown> } }`.
- [x] 2.4 Add `ok`, `err`, `okPaginated` factory functions.

## 3. Retrofit api-health

- [x] 3.1 Modify `services/api-health/main.go` to return `response.Ok(map[string]string{"status": "ok"})` wrapped in `events.APIGatewayV2HTTPResponse`.
- [x] 3.2 Update the test in `services/api-health/main_test.go` to assert `status: "success"` and `data.status: "ok"`.
- [x] 3.3 Add a negative-path test returning `response.Err[any]("HEALTH_UNAVAILABLE", nil)` if a dependency is down.

## 4. Retrofit monitor-api

- [x] 4.1 Remove `errorResponse` struct from `services/monitor-api/types.go`.
- [x] 4.2 Update every handler in `services/monitor-api/handler.go` to return `response.Envelope[...]`.
- [x] 4.3 Map existing sentinel checks (`errors.Is(err, errServiceAlreadyExists)`) to `response.Err[any]("SERVICE_ALREADY_EXISTS", map[string]any{"id": id})`.
- [x] 4.4 Map validation failures to `response.Err[any]("VALIDATION_FAILED", map[string]any{"field": "..."})`.
- [x] 4.5 Map paginated list endpoints (services, monitors, runs) to `response.OkPaginated(...)`.
- [x] 4.6 Update `services/monitor-api/main_test.go` to assert envelope fields and drop flat `errorResponse` assertions.
- [x] 4.7 Update `apps/dashboard/lib/api.ts` to parse `status`, branch on `isSuccess`/`isError`, surface `reason.code` via the `ApiError` class.

## 5. Dashboard consumer updates

- [x] 5.1 Update `apps/dashboard/lib/actions.ts` callers (22+ sites) to read `result.reason?.code` instead of `result.error`.
- [x] 5.2 Update toast/alert components to display `result.reason?.details` when present.
- [x] 5.3 Update `ApiError` class in `apps/dashboard/lib/api.ts` to accept `{ code, details, message? }` instead of a flat string.

## 6. Tests

- [x] 6.1 Add `shared/api/response/envelope_test.go` covering: success omits `reason`, error omits `data`, pagination appears only when set, JSON shape matches the constitution.
- [x] 6.2 Add `apps/dashboard/lib/api-response.test.ts` covering the same invariants and the type guards.
- [x] 6.3 Run `make test-go-all`, `make lint-go`, `make check-dashboard`, `make lint-dashboard`. All pass.

## 7. Documentation

- [x] 7.1 Update `AGENTS.md` with a short subsection on the envelope shape and the helper imports.
- [x] 7.2 Note the retirement of the flat `errorResponse` shape in `services/monitor-api/types.go` doc comment.
