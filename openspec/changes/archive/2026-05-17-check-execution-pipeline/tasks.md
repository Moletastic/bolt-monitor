## 1. Define Execution Contracts

- [x] 1.1 Define scheduler/worker execution contract for selecting enabled monitors and routing them to probe locations.
- [x] 1.2 Define normalized execution result shape shared by all supported protocol runners.

## 2. Add Stop-Safe Execution Control

- [x] 2.1 Implement hard execution gate so disabled monitors never run.
- [x] 2.2 Ensure periodic execution cannot be enabled without a reliable stop/disable control path.

## 3. Verify Pipeline Behavior

- [x] 3.1 Add tests for enabled selection, disabled skip behavior, and probe-location routing.
- [x] 3.2 Add tests or docs proving periodic execution can be stopped before runaway billing risk.
