## Context

Target parsing retains optional budget fields only when individually valid. Invalid input becomes absence and bypasses conditional budget infrastructure.

## Goals / Non-Goals

**Goals:** Distinguish omitted configuration from invalid supplied configuration. Make persistent budget-alert intent explicit.

**Non-Goals:** New cost anomaly service, commitment purchase, per-resource budgets, or high-cardinality cost metrics.

## Decisions

- Parse budget fields as a pair. If either appears, require finite positive USD amount and non-empty valid email list.
- Persistent target defaults to budget required; any opt-out uses explicit named boolean with rationale, never silent omission.
- Ephemeral targets may omit budget to avoid per-preview budget resource cost/noise.

## Risks / Trade-offs

- Existing persistent target lacks budget -> Validation blocks next deploy until configured or explicit opt-out chosen.
- Additional budget resources cost a small fixed amount -> Forecast alerts prevent larger unbounded spend.

## Migration Plan

Update example target and docs. Deploy persistent target only after valid budget config. No data migration.
