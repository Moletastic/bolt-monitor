## 1. Search Index Storage

- [x] 1.1 Add DynamoDB schema helpers and record types for tenant-scoped search index records.
- [x] 1.2 Implement normalization, tokenization, safe-display extraction, bounded prefix generation, and de-duplication helpers.
- [x] 1.3 Write search index entries for service create and update operations.
- [x] 1.4 Write search index entries for monitor create and update operations.
- [x] 1.5 Write search index entries for escalation policy create and update operations.
- [x] 1.6 Write search index entries for notification channel create and update operations.
- [x] 1.7 Delete search index entries when services, monitors, escalation policies, or notification channels are deleted.
- [x] 1.8 Add a bounded backfill path or helper for existing services, monitors, escalation policies, and notification channels.

## 2. Search API

- [x] 2.1 Add repository method for tenant-scoped search index prefix queries with bounded reads, de-duplication, and ranking.
- [x] 2.2 Add `GET /api/v1/search` handler with `q`, `limit`, and `types` query parameters.
- [x] 2.3 Return typed search result payloads through the shared response envelope.
- [x] 2.4 Add the `GET /api/v1/search` SST route.
- [x] 2.5 Add API tests for valid search, short query, type filtering, result limit, safe display text, and no-scan behavior.

## 3. Dashboard API Client

- [x] 3.1 Add dashboard TypeScript types for global search result resource types and payloads.
- [x] 3.2 Add dashboard API helper for calling `GET /api/v1/search` and unwrapping the response envelope.
- [x] 3.3 Add error handling that surfaces search API failures without disrupting the current page.

## 4. Top-Bar Search UI

- [x] 4.1 Add a top bar to `AppShell` with a search icon and search input.
- [x] 4.2 Implement debounced search requests with minimum query length and loading feedback.
- [x] 4.3 Render result popover/list with resource-specific icons, primary text, secondary text, and link navigation.
- [x] 4.4 Render idle, no-results, and error feedback states.
- [x] 4.5 Preserve keyboard and mobile accessibility for the search input and result links.

## 5. Verification

- [x] 5.1 Run `make test-go-all`.
- [x] 5.2 Run `make lint-go`.
- [x] 5.3 Run `make lint-dashboard`.
- [x] 5.4 Run `make check-dashboard`.
- [x] 5.5 Run `make test-dashboard`.
- [x] 5.6 Run `make check-infra`.
