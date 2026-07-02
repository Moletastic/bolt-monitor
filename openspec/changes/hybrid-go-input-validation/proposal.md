## Why

The Go backend validates domain structs through explicit `Validate()` methods, `shared/rules`, and typed `VALIDATION_FAILED` errors. That works well for business invariants, but request DTO validation is still spread across handler functions as repetitive inline checks for required fields, string length, JSON object shape, and enum presence.

As monitor API surface area grows, mixing request-shape validation with domain invariants makes handlers harder to scan and increases the chance that new endpoints return inconsistent `details.field` / `details.reason` payloads. A small hybrid validation layer can let request DTOs use declarative tags while preserving explicit domain validation for rules that need business context.

## What Changes

- Add `go-playground/validator` for API request DTO/input-shape validation in Go service code.
- Keep persisted/domain structs validated by explicit `Validate()` methods and `shared/rules`.
- Add a backend adapter that converts validator failures into existing `shared/errors.CodeValidationFailed` typed errors.
- Refactor monitor API request validation for DTO-level checks first, including notification channel and escalation policy request payloads.
- Preserve the existing response envelope and error wire format.

## Capabilities

### New Capabilities
- `go-input-validation`: Hybrid Go input validation for API DTOs.

### Modified Capabilities
- `code-patterns-foundation`: Clarifies the boundary between DTO input validation and domain rule validation.

## Impact

- Adds a Go dependency on `github.com/go-playground/validator/v10` in modules that validate request DTOs.
- Adds shared or service-local validation adapter code for translating validator errors to typed API errors.
- Updates `services/monitor-api` request DTOs and handler validation paths.
- Does not replace `shared/rules` or domain `Validate()` methods.
- Does not change response envelope shape, error code strings, or dashboard parsing behavior.
