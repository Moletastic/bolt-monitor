## 1. Add Operational Read Endpoints

- [x] 1.1 Add routes and handlers for latest monitor status and recent run-history reads.
- [x] 1.2 Extend monitor read responses to support dashboard-oriented status summaries.

## 2. Connect Read Models

- [x] 2.1 Read persisted `MonitorStatus` records for status endpoints.
- [x] 2.2 Read recent `CheckRun` records for run-history endpoints.

## 3. Verify Read Behavior

- [x] 3.1 Add tests for latest status and recent run-history responses.
- [x] 3.2 Add docs/comments describing how config reads differ from operational status/history reads.
