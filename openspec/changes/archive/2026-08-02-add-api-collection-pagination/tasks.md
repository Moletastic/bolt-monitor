## 1. Delivery Query

- [x] 1.1 Add bounded DynamoDB delivery page query with stable creation-time and delivery-ID ordering.
- [x] 1.2 Reuse validated opaque cursor encoding scoped to the incident delivery collection.
- [x] 1.3 Replace in-memory 200-record truncation with cursor response metadata.

## 2. Consumer Contract

- [x] 2.1 Document `cursor`, bounded `limit`, and cursor response metadata in OpenAPI and Bruno.
- [x] 2.2 Update dashboard delivery reads to request continuation pages where needed.

## 3. Verification

- [x] 3.1 Add tests for empty page, limit validation, invalid cursor, 201-record boundary, and full traversal without duplicates.
- [x] 3.2 Run `make test-go-all`, `make check-api-contract`, `make check-bruno`, and dashboard checks if dashboard code changes.
