## 1. Define Result And Status Models

- [x] 1.1 Add shared `CheckRun` and `MonitorStatus` contracts with common execution fields and extensibility for protocol-specific details.
- [x] 1.2 Map result and status contracts onto documented single-table item families.

## 2. Add Persistence Semantics

- [x] 2.1 Implement repository or mapping helpers for raw run writes and latest status updates.
- [x] 2.2 Encode TTL or retention policy for raw `CheckRun` items.

## 3. Verify Contract Behavior

- [x] 3.1 Add tests for raw-run persistence and latest-status updates.
- [x] 3.2 Add docs/comments explaining difference between historical run data and current status snapshots.
