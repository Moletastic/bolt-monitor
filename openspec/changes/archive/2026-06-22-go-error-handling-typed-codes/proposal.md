## Why

The response envelope defines `reason.code` as the machine-readable identity of every failure, but the monitor API today treats that field as an opaque string. `services/monitor-api/handler.go` declares 21 distinct code literals inline (`SERVICE_NOT_FOUND`, `VALIDATION_FAILED`, `IMMUTABLE_FIELD`, …), maps repository sentinels to those literals via `errors.Is` switches at six sites, and uses three string-equality comparisons against `err.Error()`. The dashboard (`apps/dashboard/lib/api.ts:60`) propagates `reason.code` into `ApiError.code` but every consumer still branches on HTTP status, not on code — so today the codes are a parallel contract with no enforcement that it stays consistent.

Two changes already in flight depend on a registry that does not yet exist. `code-patterns-foundation/specs/code-patterns-foundation/spec.md:25` references `*shared/errors.TypedError` with `code = VALIDATION_FAILED` and `details.field` as the contract rule failures will produce. `ts-result-and-no-any/proposal.md:8` proposes a TypeScript enum that mirrors the codes registered in `shared/errors/*` Go subpackages. Both changes are blocked on a module that this change introduces.

`api-response-envelope` is complete and stable. The envelope shape, the `response.Err` constructor, and the `Envelope.MarshalJSON` behavior are settled. This change consumes that envelope rather than modifying it.

## What Changes

- Add a Go module `shared/errors` defining `Code` (a typed string), the full set of code constants that the monitor API currently emits, a registry mapping each code to its HTTP status, a `TypedError` type carrying `Code` + `Details` + `Cause`, and a single `Respond(err)` helper that handlers call.
- Convert every validation site (`shared/monitorconfig/model.go`, the inline validators in `services/monitor-api/handler.go`) to return `*TypedError{Code: CodeValidationFailed, Details: {"field": <path>, "reason": <text>}}` instead of `fmt.Errorf`.
- Convert every repository sentinel in `services/monitor-api/repository.go` to a `*TypedError` with the matching code, including the three sites that currently return plain `errors.New("service not found")` and are matched via `err.Error() ==`.
- Replace every `errResponse(...)` and `serverError(...)` call in `services/monitor-api/handler.go` with `errors.Respond(err)`. Drop `details.id` from not-found responses; the dashboard has no consumer of that key, and the wire stays cleaner.
- `go.work` adds `./shared/errors`.
- Tests in `shared/errors/*_test.go` pin the registry. `services/monitor-api/main_test.go` and `repository_test.go` assertions migrate from string-content checks (`details.detail` contains `"botToken is required"`) to field-path checks (`details.field == "config.botToken"`).

## Capabilities

### New Capabilities

- `go-error-handling-typed-codes`: the `shared/errors` Go module with typed error codes (`Code` type, constants, registry), the `TypedError` type, and the `Respond(err)` helper. Defines the contract that `code-patterns-foundation` and `ts-result-and-no-any` already reference.

### Modified Capabilities

- `api-response-envelope`: handler sites that previously called `errResponse` and `serverError` route through `errors.Respond`. The envelope shape is unchanged; `reason.code` is now sourced from typed constants. `Envelope.MarshalJSON` and `response.Err` are unmodified.
- `monitor-crud-api`: every CRUD, status, runs, manual-run, audit, and probe-locations endpoint routes errors through `errors.Respond`. Repository sentinels become typed. Validation surfaces `details.field`. Not-found responses no longer carry `details.id`.
- `monitor-configuration`: `Service.Validate`, `Monitor.Validate`, `HTTPConfiguration.Validate`, and `Monitor.ValidateWithCatalog` return `*TypedError` with field paths.
- `notification-channel-crud`: `validateNotificationChannel` returns `*TypedError` with field paths.
- `escalation-policy-crud`: `validateEscalationPolicy` returns `*TypedError` with field paths, including step-indexed paths for steps.
- `notification-route-channel-reference`: `INLINE_CHANNEL_CONFIG` becomes a typed constant.

## Impact

- New Go module `shared/errors` with `go.mod`, `code.go`, `typed.go`, `respond.go`, plus `code_test.go`, `typed_test.go`, `respond_test.go`.
- `go.work` gains `./shared/errors` next to the existing shared modules.
- `shared/monitorconfig/model.go`: ~9 `fmt.Errorf` sites in `Service.Validate`, `Monitor.Validate`, `HTTPConfiguration.Validate`, `Monitor.ValidateWithCatalog` become `sharederrors.Wrap(CodeValidationFailed, ..., {"field": ..., "reason": ...})`.
- `services/monitor-api/repository.go`: 6 sentinel vars become `sharederrors.New(CodeXxx, nil)`. 3 plain `errors.New("service not found")` returns in `ArchiveService`, `ReactivateService` (twice) become typed.
- `services/monitor-api/handler.go`: `validateNotificationChannel` (~9 sites) and `validateEscalationPolicy` (~4 sites) return typed errors with field paths. `requestHasInlineStepConfig` stays boolean; handler turns it into `CodeInlineChannelConfig` at the call site. ~30 `errResponse`/`serverError` call sites collapse to `errors.Respond(err)`. The `errResponse` and `serverError` helpers are deleted.
- `services/monitor-api/main_test.go`: `TestCreateNotificationChannelValidation` (line 923) and any sibling assertions migrate from `strings.Contains(detail, "...")` to `details.field` checks. New tests pin: typed error pass-through, non-typed → INTERNAL/empty, code constant values match wire identity.
- `services/monitor-api/repository_test.go`: assertions on `err.Error()` migrate to `sharederrors.As(err, &te)` plus `te.Code` equality.
- `shared/api/response`: no change. `Envelope`, `Status`, `Reason`, `Pagination`, `Ok`, `Err`, `OkPaginated` remain the wire contract.
- `apps/dashboard`: no change. Code values are wire-identical to today's strings. The dashboard parser at `lib/api.ts:60` continues to populate `ApiError.code` from `response.reason.code`.

## Out of Scope

- TypeScript enum mirroring this registry (`ts-result-and-no-any`).
- `shared/rules.Field[T]` helper and the rules package (`code-patterns-foundation`). This change exposes `sharederrors.WithField(err, field)` as the primitive that `Field[T]` will call into.
- AWS facade migration in the repository layer (`code-patterns-foundation`).
- OpenAPI schema sync. `openapi/openapi.yaml`'s `ErrorResponse` (`{error: string}`) is intentionally left as-is; it does not match the envelope, and that drift is a separate concern.
- Fan-out of `INTERNAL` into more specific codes (e.g., `STORAGE_UNAVAILABLE`, `THROTTLED`). `INTERNAL` remains the single unspecified-failure code per the locked decision.
- Any change to the channel-in-use success envelope at `services/monitor-api/handler.go:815` (the `response.Ok[channelInUseResponse]` workaround with HTTP 409). That handler quirk survives this change.
