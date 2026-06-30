## 1. Add Monitor API Surface

- [x] 1.1 Add route wiring and handler package for `POST`, `GET`, `PATCH`, enable, and disable monitor endpoints.
- [x] 1.2 Define request and response shapes for monitor create, list, detail, and mutation flows.

## 2. Add Validation And Persistence

- [x] 2.1 Reuse shared monitor and probe-location contracts to validate monitor CRUD payloads.
- [x] 2.2 Implement DynamoDB repository writes and reads for monitor items, listing refs, and basic status records using the single-table schema contract.

## 3. Add Audit And Verification

- [x] 3.1 Persist audit events for monitor create, update, enable, and disable operations.
- [x] 3.2 Add tests and run relevant checks to verify CRUD behavior, validation failures, and persistence mappings.
