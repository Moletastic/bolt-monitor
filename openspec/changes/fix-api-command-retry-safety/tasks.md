## 1. Idempotency Foundation

- [x] 1.1 Define generic command record identity (`tenantId`, operation, resource), pending/completed sanitized result storage, TTL, and typed conflict mapping.
- [x] 1.2 Use case-insensitive request header access for every idempotency-key path.
- [ ] 1.3 Add unit tests for record reservation, expiry, payload mismatch, and concurrent equivalent requests.

## 2. Command Adoption

- [x] 2.1 Make manual runs replay one canonical pending or completed public result.
- [x] 2.2 Fix delivery replay header normalization and fingerprint extraction.
- [x] 2.3 Add idempotency to notification test sends without duplicate provider delivery; replay pending results without resending.
- [x] 2.4 Add idempotency to incident acknowledgement and resolution with transactionally coupled record reservation and mutation.

## 3. API Coverage

- [x] 3.1 Document required command idempotency headers and conflict responses in OpenAPI and Bruno.
- [ ] 3.2 Add handler tests for lowercase headers, retries, partial failure, and no duplicate side effects.
- [ ] 3.3 Run `make test-go-all`, `make check-api-contract`, and `make check-bruno`.
