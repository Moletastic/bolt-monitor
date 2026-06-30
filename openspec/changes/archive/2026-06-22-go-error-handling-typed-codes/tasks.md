# Tasks

## T1. Scaffold `shared/errors` [x]

Create the module skeleton. No consumers yet.

- New directory `shared/errors/` with `go.mod` (matching `shared/api/response/go.mod`'s style: `module bolt-monitor/shared/errors`, `go 1.26.0`).
- Add `./shared/errors` to `go.work` next to the existing shared modules.
- `code.go`: define `type Code string`; declare the 21 constants from the locked map; declare `var registry = map[Code]codeSpec{...}` with one entry per constant; declare `func StatusOf(code Code) int` that returns `registry[code].status` and panics on unknown.
- `code_test.go`: table-driven test that every constant has a registry entry; test that `StatusOf(unknown)` panics; test that each constant's `string(...)` value equals the literal wire identity (the same string currently emitted by `handler.go`).

Verify: `cd shared/errors && go test ./...` is green. The package builds clean with zero imports outside `net/http` and the standard library.

## T2. Implement `TypedError` [x]

Build on T1. No consumers yet beyond tests.

- `typed.go`: define `TypedError` with `Code`, `Details`, `Cause`, `Message` fields. Implement `Error()` (formatted as `"<code>: <cause>"` when cause is non-nil, else just the code), `Unwrap()` returning `Cause`. Implement `New`, `Wrap`, `WithField`, `As`. `WithField` returns a copy (does not mutate the receiver). `As` is a thin wrapper over `errors.As` that returns the typed value.
- `typed_test.go`: test `Error()` round-trip; test `Unwrap()`; test that `fmt.Errorf("wrap: %w", te)` followed by `As` finds the inner `TypedError`; test `WithField` immutability (mutating the returned error does not mutate the original); test `WithField` overwriting prior `details.field`.

Verify: `go test ./...` green.

## T3. Implement `Respond` [x]

Build on T1+T2. Still no service-side consumers.

- `respond.go`: declare `func Respond(err error) (int, response.Envelope[any])`. Behavior: typed → `(StatusOf(te.Code), response.Err[any](string(te.Code), te.Details))`. Non-typed → `(http.StatusInternalServerError, response.Err[any](string(CodeInternal), nil))`.
- `respond_test.go`: typed pass-through preserves details; non-typed yields `INTERNAL` with nil details; `INTERNAL` is stripped even when `err.Error()` is non-empty.

Verify: `go test ./...` green. The package now depends on `shared/api/response`.

## T4. Rewrite `shared/monitorconfig/model.go` validators [x]

Build on T1+T2.

- `model.go`: convert each `fmt.Errorf` in `Service.Validate`, `Monitor.Validate`, `HTTPConfiguration.Validate`, and `Monitor.ValidateWithCatalog` to `sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": <path>, "reason": <text>})`. Field paths follow the locked convention (e.g., `tenantId`, `name`, `http.target`, `probeLocations`).
- `model_test.go`: any assertion that does `if err.Error() != "..."` migrates to `sharederrors.As(err, &te)` plus `te.Code == sharederrors.CodeValidationFailed` plus `te.Details["field"] == "..."`.

Verify: `make test-go-all` green.

## T5. Convert repository sentinels to `*TypedError` [x]

Build on T1+T2.

- `services/monitor-api/repository.go`: replace `errServiceAlreadyExists`, `errMonitorAlreadyExists`, `errCannotDeleteActiveService`, `errCannotDeleteLastMonitorFromActiveService`, `errIncidentNotActionable`, `errMissingTableName` with `sharederrors.New(sharederrors.CodeXxx, nil)`. The three sites that return `errors.New("service not found")` (in `ArchiveService` and `ReactivateService`) become `sharederrors.New(sharederrors.CodeServiceNotFound, nil)`.
- `services/monitor-api/repository_test.go`: audit for any `err.Error()` string assertions; migrate to `sharederrors.As(err, &te)` plus `te.Code` equality.

Verify: `make test-go-all` green. Repository layer is fully typed.

## T6. Rewrite handler inline validators [x]

Build on T1+T2.

- `services/monitor-api/handler.go`: convert `validateNotificationChannel` (line 852) — 9 `fmt.Errorf` sites become typed with field paths. Field paths follow the locked convention: `name`, `type`, `target`, `config.botToken`, `config.apiKey`, `config.fromEmail`, etc. Switch cases on `channel.Type` keep their structure; each `require(key)` failure sets `details.field = "config.<key>"`.
- `validateEscalationPolicy` (line 1025): 4 `fmt.Errorf` sites become typed. The loop produces `details.field = "businessHoursPath.steps[<idx>].channelId"` (1-indexed in the message, 0-indexed in the path) and `offHoursPath.steps[<idx>].channelId` for the off-hours path.
- `requestHasInlineStepConfig` stays boolean; the handler turn it into at the call site turns it into `sharederrors.New(sharederrors.CodeInlineChannelConfig, map[string]any{"detail": "steps must reference channels by channelId; remove target and config"})`.

Verify: `make test-go-all` green. Validators emit typed errors only.

## T7. Collapse handler call sites to `errors.Respond` [x]

Build on T4+T5+T6. This is the largest mechanical step.

- `services/monitor-api/handler.go`: replace every `errResponse(statusCode, code, details)` call with `errors.Respond(err)`. Replace every `serverError(err)` call with `errors.Respond(err)`. Drop the `details.id` enrichment on not-found responses (per locked decision #8).
- Delete the `errResponse` and `serverError` helpers at the bottom of `handler.go` (lines 1096–1104).
- The `errors` package import alias becomes `sharederrors "bolt-monitor/shared/errors"`.
- `INVALID_JSON` sites that previously built `errResponse(http.StatusBadRequest, "INVALID_JSON", nil)` from `json.Unmarshal` failures now build a typed error first, then respond: `return sharederrors.Respond(sharederrors.New(sharederrors.CodeInvalidJSON, nil))`. (Or simpler: keep the `errResponse` pattern via `Respond` directly — but the typed-then-respond form is more consistent.)

Verify: `make test-go-all` green. `handler.go` is shorter. The compiler catches any missed call site because `errResponse` and `serverError` no longer exist.

## T8. Migrate test assertions [x]

Build on T7.

- `services/monitor-api/main_test.go`:
  - `TestCreateNotificationChannelValidation` (line 923): `strings.Contains(detail, "botToken is required")` migrates to `details.field == "config.botToken"`.
  - Add tests that pin the wire identity of every `CodeXxx` constant used in the handler (table-driven).
  - Add a test that `errMissingTableName` (now typed `CodeInternal`) reaches the wire as `INTERNAL` with nil details.
  - Add a test that a non-typed error from a stub repository (e.g., `errors.New("boom")`) reaches the wire as `INTERNAL` with nil details.
- `services/monitor-api/repository_test.go` (already touched in T5, verify complete).

Verify: `make test-go-all` green. `make lint-go` green. `make build-go` green.

## T9. Update AGENTS.md and verify docs [x]

- `AGENTS.md`: under the response-envelope section, add one line noting that handlers route through `errors.Respond` and that codes are defined in `shared/errors`. (The "Adding a new endpoint?" line stays.)
- `openspec/changes/go-error-handling-typed-codes/` self-review: re-read `proposal.md` and `design.md` against the diff; confirm every locked decision holds and every code constant in the diff matches the wire identity.

Verify: `openspec status --change go-error-handling-typed-codes` shows all tasks complete. `make lint-all` and `make test-go-all` green.
