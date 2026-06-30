## MODIFIED Requirements

### Requirement: Response envelope shape is uniform across services

All Lambda handlers SHALL continue to return a JSON body of the form:

```json
{
  "status": "success" | "error",
  "data": <T> | null,
  "reason": { "code": string, "details": Record<string, unknown> } | null,
  "message": string | null,
  "pagination": { "page": number, "size": number, "total": number, "items": unknown[] } | null
}
```

Optional fields SHALL be omitted from the JSON output when not applicable rather than emitted as `null`. The envelope struct, `Ok` / `Err` / `OkPaginated` constructors, and `MarshalJSON` behavior in `shared/api/response` are unchanged.

#### Scenario: Handler error sites route through shared/errors.Respond
- **WHEN** a handler produces an error response
- **THEN** the response body conforms to the envelope shape with `status: "error"` and `reason.code` sourced from a typed `shared/errors.Code` constant

#### Scenario: INTERNAL responses carry no details
- **WHEN** `shared/errors.Respond` receives a non-typed error
- **THEN** the response body's `reason.details` is empty regardless of the cause's `Error()` text

### Requirement: Existing services adopt the envelope

`services/api-health` and `services/monitor-api` SHALL continue to return the envelope from every handler. Handler error sites SHALL route through `shared/errors.Respond` rather than the local `errResponse` and `serverError` helpers, which are removed.

#### Scenario: monitor-api handler.go no longer defines errResponse
- **WHEN** `services/monitor-api/handler.go` is read
- **THEN** no top-level functions named `errResponse` or `serverError` exist

#### Scenario: api-health continues to use the envelope
- **WHEN** `services/api-health/main.go` returns its success response
- **THEN** the body conforms to the envelope with `status: "success"` and `data: { status: "ok" }`
