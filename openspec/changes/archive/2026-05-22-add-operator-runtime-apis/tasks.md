## 1. Manual run API

- [x] 1.1 Add route and handler support for `POST /api/v1/monitors/{id}/run`.
- [x] 1.2 Persist or emit enough run identity metadata for clients to correlate the accepted manual run with later run-history reads.
- [x] 1.3 Add tests for accepted manual runs, missing monitors, and disabled monitors.

## 2. Incident management API

- [x] 2.1 Add incident collection and detail read routes.
- [x] 2.2 Add monitor-scoped incident read route.
- [x] 2.3 Add acknowledge and resolve command routes for existing incidents.
- [x] 2.4 Add tests for open/closed incident reads and valid incident action transitions.

## 3. Scheduler admin API

- [x] 3.1 Add admin route and handler support for reading scheduler configuration.
- [x] 3.2 Add admin route and handler support for updating scheduler configuration.
- [x] 3.3 Add tests for recurring execution safety validation and persisted control-state reads.

## 4. Audit event read API

- [x] 4.1 Add monitor-scoped audit history read route.
- [x] 4.2 Define stable audit event response fields for mutation history views.
- [x] 4.3 Add tests for empty and populated audit history responses.

## 5. Contract and docs updates

- [x] 5.1 Add schemas and routes for new APIs to `openapi/openapi.yaml`.
- [x] 5.2 Update README or dashboard docs where new operator/admin flows become discoverable.
