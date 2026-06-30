## Why

Current API and shared models describe monitor runtime behavior, but repository still does not execute checks end to end. Manual runs stop at acceptance, recurring scheduler control has no runtime consumer, and HTTP checks do not yet enforce every configured assertion, so product still cannot reliably claim active monitoring.

## What Changes

- Implement runtime execution flow that turns accepted manual runs and recurring scheduler decisions into executed checks.
- Define queueable execution-work semantics so manual and recurring paths share one execution pipeline.
- Extend HTTP execution requirements to enforce configured body-content assertions in addition to status and timeout behavior.
- Define how completed execution results persist `CheckRun` history and latest `MonitorStatus` snapshots.
- Define first incident lifecycle rules derived from execution outcomes so incident APIs reflect live monitoring state.

## Capabilities

### New Capabilities

### Modified Capabilities
- `check-execution-pipeline`: Add execution-work materialization, shared manual and recurring execution flow, and full HTTP assertion behavior.
- `check-result-status-model`: Clarify persistence behavior for completed runtime executions across run history and latest status snapshots.
- `manual-run-api`: Require accepted manual runs to feed downstream execution work and produce observable run history.
- `incident-management-api`: Define how system-owned incidents open, update, and resolve from execution outcomes.
- `scheduler-admin-api`: Clarify how recurring scheduler state gates actual recurring execution behavior.

## Impact

- Affects `shared/checkexecution`, `shared/resultstatus`, and monitor runtime persistence contracts.
- Requires new runtime worker and recurring scheduler wiring in `services/` and `infra/`.
- Changes observable monitor behavior by turning accepted run commands and scheduler config into real checks, stored results, status updates, and incident state.
