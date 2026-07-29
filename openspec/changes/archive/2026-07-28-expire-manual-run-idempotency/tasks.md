## 1. TTL Persistence

- [x] 1.1 Persist manual idempotency expiry as absolute Unix epoch TTL.
- [x] 1.2 Keep replay behavior correct when DynamoDB TTL cleanup is delayed.

## 2. Verification

- [x] 2.1 Add repository request-shape tests for TTL epoch and configured retention.
- [x] 2.2 Run monitor API Go tests and lint.
