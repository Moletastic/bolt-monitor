## 1. Execution semantics

- [x] 1.1 Update `shared/checkexecution` to enforce `ExpectedBodyContains` alongside existing status and timeout evaluation.
- [x] 1.2 Add or update unit tests covering successful body matches, failed body matches, and normalized execution outcomes.

## 2. Execution work persistence

- [x] 2.1 Define internal execution-work record shape and repository helpers for enqueueing manual and recurring work.
- [x] 2.2 Extend monitor runtime repository code to claim runnable work, re-read monitor state, and mark skipped or completed work safely.
- [x] 2.3 Ensure completed executions persist `CheckRun` history and latest `MonitorStatus` snapshot as one completion unit.

## 3. Runtime workers

- [x] 3.1 Add manual or shared execution worker code that consumes queued work and executes HTTP checks through the shared runtime path.
- [x] 3.2 Add recurring scheduler code that reads scheduler config and materializes recurring work only when recurring execution is enabled.
- [x] 3.3 Add tests covering disabled-monitor skips, recurring scheduler gating, and manual-run-to-history flow.

## 4. Incident lifecycle

- [x] 4.1 Implement monitor-level incident open and update behavior for non-success execution outcomes.
- [x] 4.2 Implement automatic incident resolution behavior on later success outcomes.
- [x] 4.3 Add tests covering first failure open, repeated failure update, and recovery resolution.

## 5. Infrastructure wiring

- [x] 5.1 Wire new runtime Lambdas, triggers, and environment in `infra/stacks/bootstrap.ts`.
- [x] 5.2 Verify existing API routes still return accepted manual-run and scheduler responses while new runtime services consume the same table state.
