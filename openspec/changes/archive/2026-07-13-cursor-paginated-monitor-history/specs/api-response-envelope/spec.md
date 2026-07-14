## MODIFIED Requirements

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
- **THEN** response body conforms to envelope shape with `status: "error"` and `reason.code` sourced from a typed `shared/errors.Code` constant

#### Scenario: INTERNAL responses carry no details
- **WHEN** `shared/errors.Respond` receives a non-typed error
- **THEN** response body's `reason.details` is empty regardless of cause's `Error()` text

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
