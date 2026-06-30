## 1. Define Probe-Location Read Surface

- [x] 1.1 Add OpenSpec requirements for selectable probe-location reads.
- [x] 1.2 Define `/api/v1/probe-locations` response fields and API-side ordering for dashboard consumption.

## 2. Implement Backend Read Endpoint

- [x] 2.1 Add API route for probe-location collection reads.
- [x] 2.2 Add response model and handler that translate the canonical catalog into API response shape.
- [x] 2.3 Ensure returned locations reflect the set valid for operator selection in the current environment.

## 3. Verify Read Behavior

- [x] 3.1 Add tests for probe-location collection responses.
- [x] 3.2 Verify response ordering and that disabled locations are excluded from the public collection response.
- [x] 3.3 Document dashboard dependency on this endpoint for create/edit monitor flows.
