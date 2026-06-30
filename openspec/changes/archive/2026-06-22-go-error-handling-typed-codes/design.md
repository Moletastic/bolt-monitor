## Locked decisions

These decisions were settled before the proposal was drafted. They are not open during implementation.

1. **Registry scope is broad.** All 21 codes currently emitted by `services/monitor-api/handler.go` become typed constants in `shared/errors`, including handler-defined codes (`IMMUTABLE_FIELD`, `INLINE_CHANNEL_CONFIG`, `NOT_FOUND`).
2. **`code → httpStatus` lives in `shared/errors` as a single map.** Handlers never pass `http.StatusXxx` and a code together. `Respond(err)` is the only function that emits a status code.
3. **Repository sentinels return `*TypedError` directly** (option B). The handler stops using `errors.Is` to switch on sentinels. Six `errors.Is` switches and three `err.Error() == "..."` comparisons disappear.
4. **Validation carries `details.field`.** The current `details.detail` carrying the raw `fmt.Errorf` string is replaced by `details.field` (dotted path) and an optional `details.reason` for context. Field paths use square-bracket indexing for collections, e.g. `businessHoursPath.steps[2].channelId`.
5. **`INTERNAL` details are stripped.** `Respond` returns `response.Err[any]("INTERNAL", nil)` for any non-typed error, regardless of cause. Operators get failure context from CloudWatch logs; the response body carries only the code.
6. **`Field[T]` lives in `shared/rules`**, owned by `code-patterns-foundation`. `shared/errors` exposes `WithField(err, field)` as the primitive that `Field[T]` calls into. No circular dependency: `shared/rules` imports `shared/errors`, never the reverse.
7. **Handlers never type-assert `*TypedError` directly.** They call `errors.Respond(err)` for every error path. `As` is for tests and for the rare handler that needs to enrich details conditionally.
8. **`details.id` is dropped from not-found responses.** The dashboard never reads it. The migration is purely removal, not replacement. Future debug aids can land later if a real consumer appears.

## Code → status map (final)

| Code | HTTP status |
|---|---|
| `NOT_FOUND` | 404 |
| `INVALID_JSON` | 400 |
| `VALIDATION_FAILED` | 400 |
| `IMMUTABLE_FIELD` | 400 |
| `INLINE_CHANNEL_CONFIG` | 400 |
| `SERVICE_NOT_FOUND` | 404 |
| `SERVICE_ALREADY_EXISTS` | 409 |
| `SERVICE_ACTIVE` | 409 |
| `SERVICE_NOT_ARCHIVED` | 409 |
| `SERVICE_HAS_NO_POLICY` | 404 |
| `MONITOR_NOT_FOUND` | 404 |
| `MONITOR_ALREADY_EXISTS` | 409 |
| `MONITOR_DISABLED` | 409 |
| `MONITOR_STATUS_NOT_FOUND` | 404 |
| `LAST_MONITOR` | 409 |
| `INCIDENT_NOT_FOUND` | 404 |
| `INCIDENT_NOT_ACTIONABLE` | 409 |
| `POLICY_NOT_FOUND` | 404 |
| `POLICY_REFERENCED` | 409 |
| `CHANNEL_NOT_FOUND` | 404 |
| `INTERNAL` | 500 |

The map is defined once in `shared/errors/code.go` as a `map[Code]codeSpec`. `StatusOf(code)` returns the registered status. `StatusOf` panics if `code` is not registered; this is the test-time enforcement that adding a constant without registering it cannot slip through.

## TypedError shape

```go
type TypedError struct {
    Code    Code
    Details map[string]any
    Cause   error
    Message string  // optional human hint, never emitted
}

func (e *TypedError) Error() string
func (e *TypedError) Unwrap() error   // returns Cause

func New(code Code, details map[string]any) *TypedError
func Wrap(code Code, cause error, details map[string]any) *TypedError
func WithField(err *TypedError, field string) *TypedError
func As(err error) (*TypedError, bool)
```

`New` produces an error with no cause. `Wrap` produces one with a cause; `Error()` formats as `"<CODE>: <cause>"`. `WithField` returns a copy with `details.field` set (overwriting any prior `field`). `As` is a thin wrapper over `errors.As` that returns the typed value when present.

`Details` is the only field that flows to `response.reason.details`. `Message` is intentionally never serialized; the spec forbids it on the envelope. If a human-readable hint is needed, it goes into `details.reason` or `details.message`.

## Validation field paths

Field paths are dotted strings. Examples drawn from existing validators:

| Validator | Field path |
|---|---|
| `Service.Validate` (tenantId) | `tenantId` |
| `Service.Validate` (name) | `name` |
| `Service.Validate` (technologyKey) | `technologyKey` |
| `Monitor.Validate` (probeLocations) | `probeLocations` |
| `Monitor.Validate` (failureThreshold) | `failureThreshold` |
| `HTTPConfiguration.Validate` (target) | `http.target` |
| `validateNotificationChannel` (config.botToken) | `config.botToken` |
| `validateEscalationPolicy` (step 3 channel) | `businessHoursPath.steps[2].channelId` |
| `validateEscalationPolicy` (step 3 offHours) | `offHoursPath.steps[2].channelId` |

The format is `parent.child` for nested fields and `parent.children[index].child` for collection elements. There is no separate `path` key; `field` carries the full path. `reason` is a free-form short human description (e.g., `"required"`, `"must be a valid absolute URL"`).

`shared/errors` does not parse field paths. It treats `field` as an opaque string. The conventions are documented in a comment on `WithField` and pinned by a unit test that exercises a representative path.

## Repository migration shape

Today:

```go
var errServiceAlreadyExists = errors.New("service already exists")
```

After:

```go
var errServiceAlreadyExists = sharederrors.New(
    sharederrors.CodeServiceAlreadyExists, nil,
)
```

Today (three sites in `repository.go`):

```go
if err.Error() == "service not found" { ... }
```

After (the repository itself returns the typed error directly):

```go
return monitorconfig.Service{}, sharederrors.New(
    sharederrors.CodeServiceNotFound, nil,
)
```

Handler sites collapse from:

```go
created, err := h.repo.CreateService(ctx, service)
if err != nil {
    if errors.Is(err, errServiceAlreadyExists) {
        return errResponse(http.StatusConflict, "SERVICE_ALREADY_EXISTS",
            map[string]any{"id": service.ServiceID})
    }
    return serverError(err)
}
```

To:

```go
created, err := h.repo.CreateService(ctx, service)
if err != nil {
    return errors.Respond(err)
}
```

The `details.id` enrichment on not-found responses goes away. The wire shape loses that key; the dashboard has no consumer of it.

## INTERNAL details stripping

`Respond` for any non-typed error:

```go
func Respond(err error) (int, response.Envelope[any]) {
    var te *TypedError
    if errors.As(err, &te) {
        return StatusOf(te.Code), response.Err[any](string(te.Code), te.Details)
    }
    return http.StatusInternalServerError,
           response.Err[any](string(CodeInternal), nil)
}
```

The current `serverError` (`handler.go:1100`) leaks `err.Error()` into `INTERNAL.details.detail` for `errMissingTableName`. That leak is gone. Operators get the underlying error message from CloudWatch logs; the response body is opaque.

This is a deliberate debuggability trade-off. The argument for it: today's `err.Error()` strings are not stable (they go through `fmt.Errorf` and `errors.New` formatting), and leaking them invites the dashboard to start depending on text that might change. The argument against: a future debugging session might want a trace ID. The locked decision is to strip; if a future change needs trace IDs, it adds a `details.traceId` key explicitly.

## Tests

`shared/errors/code_test.go`:
- Every `CodeXxx` constant has a registry entry (table-driven test against `registry`).
- `StatusOf(unknown)` panics.
- Each constant's string value equals the literal it replaces on the wire.

`shared/errors/typed_test.go`:
- `Error()` and `Unwrap()` round-trip.
- `As(err, &te)` finds a `TypedError` even when wrapped via `fmt.Errorf("...: %w", te)`.
- `WithField` returns a new error with `details.field` set without mutating the original.
- `Wrap` carries cause into `Error()` output.

`shared/errors/respond_test.go`:
- Typed error with known code → (registered status, envelope with matching code/details).
- Non-typed error → (500, envelope with code `INTERNAL` and nil details).
- `INTERNAL` details remain nil even when the cause's `Error()` is non-empty.

`services/monitor-api/main_test.go`:
- `TestCreateNotificationChannelValidation` (line 923): the `strings.Contains(detail, "botToken is required")` assertion migrates to `details.field == "config.botToken"`.
- New tests for each typed sentinel: repository returns `*TypedError` with the expected code; handler routes through `Respond` without modification.
- New test: a non-typed error from the repository (`errMissingTableName` from `repository.go:25`) reaches the wire as `INTERNAL` with nil details.

`services/monitor-api/repository_test.go`:
- Any test asserting on `err.Error()` migrates to `sharederrors.As(err, &te)` and `te.Code`.

## Compatibility with in-flight changes

**`code-patterns-foundation`**: lands `shared/rules` with `Field[T]`. `Field[T]` calls `sharederrors.WithField` from `shared/errors`. The dependency direction is `shared/rules` → `shared/errors`. This change does not introduce `Field[T]`; it provides the primitive. `code-patterns-foundation/specs/.../spec.md:25` continues to hold: rule failures surface as `*shared/errors.TypedError` with `code = VALIDATION_FAILED` and `details.field` per failing rule.

**`ts-result-and-no-any`**: mirrors the Go `Code` constants as a TypeScript enum. The Go side stays wire-compatible; the strings emitted today (`SERVICE_NOT_FOUND`, etc.) are unchanged. The TS enum becomes a one-shot translation once this change lands.

**`api-response-envelope`** (complete, stable): this change consumes `response.Err` exactly as the envelope spec defines. No envelope changes. The 21 string literals that disappear from `handler.go` reappear as typed constants in `shared/errors` with the same string values.

## Risks

- **Wire incompatibility if a code string changes.** Mitigated by a single test asserting every `CodeXxx` constant equals the literal string it replaces. Any future code change that wants to rename a code is a breaking API change and would propose a new capability.
- **Field-path convention drift.** Mitigated by a comment on `WithField` documenting the format and a unit test exercising a representative path including bracket indexing.
- **Repository test silent breakage.** T5 in the task list explicitly audits `repository_test.go` for `err.Error()` assertions before the handler collapses.
- **Operator debugging gets harder** because `INTERNAL` no longer carries `err.Error()`. Documented as a deliberate trade-off. CloudWatch logs remain the source of failure detail; no logging code changes.

## Open questions

None. All eight decisions are locked.
