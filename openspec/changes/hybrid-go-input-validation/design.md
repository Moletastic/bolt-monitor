## Context

The repository currently has two validation styles:

- Domain validation: explicit `Validate()` methods, often using `shared/rules`, returning `shared/errors.CodeValidationFailed` with `details.field` and `details.reason`.
- Handler/request validation: ad hoc inline checks in `services/monitor-api/handler.go`, especially around notification channels and escalation policies.

The existing domain validation is a good fit for invariants such as supported monitor intervals, probe location catalog membership, HTTP configuration semantics, and cross-resource checks. A tag-based validator is a better fit for repetitive DTO input shape checks such as required names, max lengths, required nested arrays, and simple enum presence.

## Boundary

Use validator tags for request DTO checks that can be decided from the decoded payload alone:

- Required field presence after JSON decoding.
- Trimmed string non-emptiness when supported by a custom tag or DTO normalization step.
- Maximum string lengths.
- Minimum slice length.
- Basic one-of enum values when the enum set is stable and local to input parsing.

Keep explicit Go code for validation that needs domain behavior, repository state, or clearer branching:

- Domain model invariants in `shared/monitorconfig` and similar shared packages.
- Probe location catalog checks.
- Escalation policy channel existence checks against the repository.
- Type-specific config requirements when the required fields depend on runtime channel type, unless a custom validator keeps this clearer than explicit code.
- Normalization and defaulting such as trimming, generated IDs, default thresholds, and tenant/service ID assignment.

## Shape

The preferred flow is:

```text
API Gateway request
    │
    ▼
json.Unmarshal into request DTO
    │
    ▼
normalize request DTO where needed
    │
    ▼
validator.Struct(dto)
    │
    ▼
validation adapter maps first failure
    │
    ▼
shared/errors.CodeValidationFailed
    │
    ▼
domain constructor / domain Validate()
```

DTO validation and domain validation are both allowed in the same request path. DTO validation rejects malformed input early; domain validation remains the source of truth for persisted model correctness.

## Error Mapping

Validator failures must preserve the existing API contract:

```json
{
  "status": "error",
  "reason": {
    "code": "VALIDATION_FAILED",
    "details": {
      "field": "name",
      "reason": "required"
    }
  }
}
```

The adapter should:

- Return `shared/errors.CodeValidationFailed`.
- Use JSON field names, not Go struct field names.
- Support dotted and indexed paths for nested fields when possible.
- Map validator tags to stable reason strings such as `required`, `must be 80 characters or less`, or `must have at least one item`.
- Return one deterministic failure at a time unless a future spec explicitly changes validation responses to include multiple field errors.

## Dependency Placement

The validator dependency should live where request DTO validation happens. If only `services/monitor-api` uses it initially, keep it service-local. Move an adapter into a shared package only when a second Go service needs the same mapping behavior.

This avoids turning domain packages into validator consumers and keeps shared modules lightweight.

## Testing Strategy

- Unit test the validator adapter with simple, nested, and slice/indexed DTO examples.
- Handler tests should continue asserting response envelopes for invalid requests.
- Existing domain validation tests should remain in place and should not depend on validator internals.
- Add drift/guard tests or grep-based tests only if needed to enforce that domain structs do not gain validator tags accidentally.

## Non-Goals

- Do not replace `shared/rules`.
- Do not add `validate` tags to persisted/domain structs unless a future spec explicitly chooses that.
- Do not change `VALIDATION_FAILED` wire identity.
- Do not return arrays of validation errors in this change.
- Do not introduce a second response envelope shape.
