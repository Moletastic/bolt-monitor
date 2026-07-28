## Context

Target loading applies future-expiry validation to every operation. Removal resolves that target before calling the existing exact-stage SST cleanup and residual verifier.

## Goals / Non-Goals

**Goals:** Remove expired disposable stages safely. Keep expired targets unavailable for all non-removal operations. Prevent billable leftovers without a new always-on service.

**Non-Goals:** Automatic deletion at expiry, changing persistent deletion controls, or cleanup by broad tag/name matching.

## Decisions

- Split target validation into structural/lifecycle validation and operation-aware expiry validation. Removal accepts only expired ephemeral targets with `disposable=true`.
- Pass operation intent from orchestrator before target validation. Existing exact stage, account, region, SST state, and residual checks remain mandatory.
- No janitor resource. Operator-triggered removal has zero idle cost and uses existing cleanup flow.

## Risks / Trade-offs

- Expired file enables destructive removal -> Scope remains target account/region/stage and cleanup rejects non-ephemeral targets.
- Forgotten target still incurs cost until operator runs removal -> Document expiry as cleanup trigger, not automatic enforcement.

## Migration Plan

Deploy validation change. Test removal against an expired disposable fixture. No AWS schema or resource migration.
