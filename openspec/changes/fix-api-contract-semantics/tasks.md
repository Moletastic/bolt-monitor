## 1. Runtime Semantics

- [x] 1.1 Add typed channel-in-use conflict code and safe reference details.
- [x] 1.2 Route referenced channel deletion through the standard error envelope.
- [x] 1.3 Add handler tests for `201`, `204`, validation, authorization, not found, and conflict envelopes.

## 2. OpenAPI Contract

- [x] 2.1 Define reusable success and error envelope schemas.
- [x] 2.2 Document request bodies, path/query parameters, cursor metadata, location headers, idempotency headers, and typed errors.
- [x] 2.3 Update Bruno examples and assertions for changed conflict semantics.

## 3. Release Gates

- [x] 3.1 Extend API contract validators with semantic envelope and required operation metadata checks.
- [x] 3.2 Add deterministic validator fixtures for passing and failing semantic cases.
- [x] 3.3 Run `make test-go-all`, `make check-api-contract`, and `make check-bruno`.
