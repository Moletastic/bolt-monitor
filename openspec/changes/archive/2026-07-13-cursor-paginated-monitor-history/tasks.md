## 1. Audit Resource Index and Migration

- [x] 1.1 Add `GSI3PK` and `GSI3SK` table fields and provision `AuditByResourceIndex` in SST with an exact documented index name.
- [x] 1.2 Add canonical audit resource-index key builders and populate index attributes from `NewAuditEventRecord`.
- [x] 1.3 Add schema and record-construction tests for monitor- and service-level audit resource keys.
- [x] 1.4 Implement an idempotent operational backfill for existing `AuditEvent` records missing resource-index attributes.
- [ ] 1.5 Add backfill verification covering record selection, idempotency, and historical monitor/service audit visibility.

## 2. Cursor Pagination Contract

- [x] 2.1 Extend Go response pagination types and constructors with cursor metadata while preserving existing page-pagination behavior.
- [x] 2.2 Extend dashboard API envelope types and type guards for cursor-pagination metadata without `any`.
- [x] 2.3 Add opaque cursor encode, decode, and resource-key validation helpers with malformed-cursor tests.
- [ ] 2.4 Add API and envelope tests for first page, continuation page, final page, and invalid cursor responses.

## 3. Bounded Monitor History Readers

- [x] 3.1 Refactor monitor run repository and handler reads to accept cursor input and return 20 newest records plus DynamoDB continuation state.
- [x] 3.2 Refactor monitor incident repository and handler reads to accept cursor input and return 20 newest records plus DynamoDB continuation state.
- [x] 3.3 Replace monitor and service audit tenant-range filtering with exact `AuditByResourceIndex` queries, cursor input, and 20-record limits.
- [ ] 3.4 Update monitor API route parsing, response payloads, repository interfaces, fakes, and handler tests for cursor history endpoints.
- [x] 3.5 Confirm history readers do not apply in-memory sort or unrelated-resource filtering after DynamoDB query.

## 4. Dashboard History Loading

- [x] 4.1 Update dashboard history API helpers and types to request and return cursor pages.
- [x] 4.2 Refactor monitor detail server loading so latest runs seed timeline, metrics, and Runs tab while inactive Incidents and Audit tabs are not eagerly fetched.
- [ ] 4.3 Add client evidence-tab state that lazy-loads initial non-run tabs through server actions and retains each loaded tab page during local tab switches.
- [ ] 4.4 Add reusable outlined `Load more` behavior for Runs, Incidents, and Audit tables that appends continuation pages, prevents duplicate requests, and keeps rows visible on failure.
- [ ] 4.5 Preserve accessible tab semantics, mobile table/card behavior, empty states, loading feedback, and retry feedback.

## 5. Verification and Rollout

- [ ] 5.1 Run the audit backfill in target environment and verify indexed historical audit samples before audit reader cutover.
- [ ] 5.2 Add dashboard tests proving inactive history tabs do not request data, loaded tabs do not re-request data, and load-more appends exactly one page.
- [ ] 5.3 Run `make test-go-all`, `make lint-go`, `make test-dashboard`, `make check-dashboard`, `make lint-dashboard`, and `make check-infra`.
