## Why

Malformed supplied budget fields are silently discarded, disabling budget alerts without operator awareness. Persistent target configuration must fail loudly when FinOps intent is incomplete or invalid.

## What Changes

- Reject invalid or partial supplied `budgetAmountUsd` and `alertEmails` configuration.
- Require valid paired budget configuration for persistent targets, unless a deliberate documented opt-out is modeled explicitly.
- Add target validation tests for malformed, partial, and opt-out configurations.
- FinOps: uses existing conditional AWS Budget resource; no added runtime resource beyond intentionally configured budget alerts.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `stage-resource-lifecycle`: Persistent target FinOps configuration must be explicit and fail closed.

## Impact

- `infra/deployment-target.ts`
- `infra/budget-infrastructure.test.ts`
- Target examples and lifecycle documentation
