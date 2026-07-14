## Overview

Every Lambda entry point returns a uniform response envelope so that consumers (dashboard, future SDKs) can parse success, failure, and pagination through a single shape. The envelope is the contract that lets the centralized error-code registry land without a follow-on consumer change.

## Requirements

### Requirement: Response envelope shape is uniform across services

All Lambda handlers SHALL continue to return a JSON body of the form:

```json
{
  "status": "success" | "error",
  "data": <T> | null,
  "reason": { "code": string, "details": Record<string, unknown> } | null,
  "message": string | null,
  "pagination": { "page": number, "size": number, "total": number, "items": unknown[] } | { "size": number, "nextCursor": string } | null
}
```

Optional fields SHALL be omitted from JSON output when not applicable rather than emitted as `null`. The envelope struct, `Ok` / `Err` / `OkPaginated` constructors, and `MarshalJSON` behavior in `shared/api/response` SHALL remain the common response mechanism; cursor-paginated responses SHALL use a dedicated constructor.

#### Scenario: Handler error sites route through shared/errors.Respond
- **WHEN** a handler produces an error response
- **THEN** the response body conforms to the envelope shape with `status: "error"` and `reason.code` sourced from a typed `shared/errors.Code` constant

#### Scenario: INTERNAL responses carry no details
- **WHEN** `shared/errors.Respond` receives a non-typed error
- **THEN** the response body's `reason.details` is empty regardless of the cause's `Error()` text

### Requirement: Status is a typed enum

The `status` field SHALL be a string literal `"success"` or `"error"`. Implementations SHALL expose this as a typed enum (Go: `Status` type with named constants; TypeScript: `Status` enum), not as a raw string.

### Requirement: Success carries `data` and optional `message`

On success, the envelope SHALL include `data` and MAY include `message`. `reason` SHALL be absent on success.

### Requirement: Failure carries `reason` with a machine-readable code

On failure, the envelope SHALL include `reason.code` as a stable, machine-readable string. `reason.details` SHALL be an object (Go: `map[string]any`; TypeScript: `Record<string, unknown>`) carrying structured error context. `data` SHALL be absent on failure. `message` SHALL be absent on failure; human-readable detail lives in `reason.details` or is logged server-side.

### Requirement: Pagination object

When an endpoint returns an offset-paginated collection, envelope SHALL include `pagination: { page, size, total, items }`. `items` SHALL be current page items and `page` SHALL be 1-indexed. When an endpoint returns a cursor-paginated collection, envelope SHALL include `pagination: { size, nextCursor? }`; `nextCursor` SHALL be omitted when no following page exists, and `page`, `total`, and `items` SHALL be omitted. Endpoints returning a single resource SHALL omit `pagination`.

#### Scenario: Cursor-paginated collection has more records
- **WHEN** a handler returns a cursor page with a continuation key
- **THEN** response includes `pagination.size` and opaque `pagination.nextCursor`
- **AND** response omits page-number and total-count fields

#### Scenario: Cursor-paginated collection is final page
- **WHEN** a handler returns a cursor page with no continuation key
- **THEN** response includes `pagination.size`
- **AND** response omits `pagination.nextCursor`

### Requirement: Go envelope is a generic struct

The Go envelope SHALL be implemented as a generic struct `Envelope[T any]` exported from `shared/api/response`. The package SHALL provide constructors `Ok`, `Err`, and `OkPaginated`. `MarshalJSON` SHALL emit the shape above, omitting nil optional fields.

### Requirement: TypeScript envelope is a class with type guards

The TypeScript envelope SHALL be implemented as an `ApiResponse<T>` type plus factory functions `ok`, `err`, `okPaginated` in `apps/dashboard/lib/api-response.ts`. The module SHALL export type guards `isSuccess` and `isError` that narrow the response type.

### Requirement: Existing services adopt the envelope

`services/api-health` and `services/monitor-api` SHALL continue to return the envelope from every handler. Handler error sites SHALL route through `shared/errors.Respond` rather than the local `errResponse` and `serverError` helpers, which are removed.

#### Scenario: monitor-api handler.go no longer defines errResponse
- **WHEN** `services/monitor-api/handler.go` is read
- **THEN** no top-level functions named `errResponse` or `serverError` exist

#### Scenario: api-health continues to use the envelope
- **WHEN** `services/api-health/main.go` returns its success response
- **THEN** the body conforms to the envelope with `status: "success"` and `data: { status: "ok" }`

### Requirement: Envelope is the only response shape

Once this change lands, no handler SHALL emit a response body that does not conform to the envelope. New code that needs a different shape SHALL propose a new capability.
