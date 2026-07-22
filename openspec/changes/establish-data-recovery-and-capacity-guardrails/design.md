## Context

Bolt Monitor is single-region, single-tenant (`DEFAULT`), serverless, and self-deployed. Operators running small installations need a defensible cost estimate and an optional safety rail. The full data recovery and bounded-access refactor is intentionally out of scope for this trimmed change — those land when a profile exceeds its named envelope or when a recovery drill is actually needed.

This change owns the cost vocabulary and the optional budget guardrail only. It does not redefine lifecycle protection (owned by `standardize-stage-resource-lifecycle`), retry semantics (owned by `make-check-execution-retry-safe`), notification assurance (owned by `assure-notification-and-escalation-delivery`), or pipeline health (owned by `expose-monitoring-pipeline-health` — now archived).

## Goals / Non-Goals

**Goals:**

- Define three named profiles with explicit dimensions so cost estimates and operational support boundaries share one vocabulary.
- Produce a reproducible cost worksheet per profile from public pricing for the deployed region.
- Document an optional stage-attributed AWS Budget setup with explicit no-op default.
- Guarantee that absent budget configuration creates no AWS resource and does not fail deployment.

**Non-Goals:**

- No SLA, RPO, or RTO commitment.
- No data inventory, retention class constants, or TTL assertions.
- No recovery runbook, recovery drill, or restore tooling.
- No scheduler projection overhaul or bounded API refactor.
- No load runner, drill fixtures, or p50/p95/p99 evidence collection.
- No multi-region, active-active, or cross-region failover.
- No automatic monitoring shutdown when a budget threshold is reached.

## Decisions

### Three named profiles

Three profiles prevent stress evidence from becoming a default-cost claim:

- **Default low-cost owner profile:** one active tenant, up to 10 monitors at five-minute cadence, modest operator traffic, 30-day raw-run retention. This is the owner-operated baseline used for low-cost estimates; it is not a Free Tier guarantee.
- **Expected validation profile:** one active tenant, up to 100 services and 100 monitors at one-minute cadence, representative histories, up to 10 concurrent operator requests.
- **High-volume stress profile:** one active tenant, up to 100 services and 1,000 enabled monitors at one-minute cadence, representative 30-day history, up to 25 concurrent operator requests. Explicitly not a default or free-tier claim.

Dimensions: tenant count, service count, monitor count, monitor cadence, run retention, operator request concurrency, audit/incident/activity volume.

### Reproducible cost worksheet

The worksheet is a static markdown document under `docs/cost-worksheet.md`. It records the pricing date and region in the header, lists the per-service pricing inputs used, and shows the resulting monthly cost estimate per profile broken down by service (`AppTable`, `AuthTable`, Cognito, SSM, Lambda, SQS, API Gateway, logs/alarms, dashboard attribution, PITR). Each profile section identifies the dominant cost driver and shows the formula used. The document explicitly disclaims Free Tier eligibility for the low-cost owner profile.

Reproduction means any reviewer can re-derive the numbers from the documented inputs and formulas without running code. Pricing date in the header signals when estimates need a refresh.

### Optional stage-attributed AWS Budget

The budget is opt-in via target configuration. When enabled:

- One AWS Budget scoped to the stage tags (`service`, `stage`).
- Forecast alert at 80% of configured amount.
- Actual alert at 100% of configured amount.
- Recipients come from deployment configuration (target file or env), not source-controlled personal addresses.
- No automatic action; alerts only.
- Absent configuration yields no resource and no deployment failure.

Documentation covers setup, verification (`aws budgets describe-budgets`), alert validation, and the no-op default.

### No-op default is load-bearing

The repository must deploy cleanly without an AWS Budget configured. The SST wiring conditionally adds the budget only when the target provides the required inputs (`budgetAmountUsd`, `alertEmails`). Missing inputs = no resource, no error. This guarantees that a clean account without budget permissions can still install bolt-monitor.

## Risks / Trade-offs

- [Pricing drift] Public pricing changes invalidate the worksheet. → Mitigation: pricing date in artifact header, regenerate periodically.
- [Budget recipient misconfiguration] Wrong addresses receive alerts. → Mitigation: deployment configuration only, never source-controlled.
- [Scope too thin] Future needs will require a follow-on change for inventory and recovery. → Acknowledged and deferred by design.
- [Profile misuse] Treating stress profile cost as default cost. → Mitigation: explicit disclaimers in artifact and proposal.

## Migration Plan

1. Add `docs/profiles.md`, `docs/cost-worksheet.md`, `docs/optional-budget-setup.md` with the three profile definitions, worksheet contents, and budget setup steps.
2. Optionally add SST wiring under `infra/` for one stage-attributed AWS Budget gated by target configuration.
3. Add an infrastructure test that proves absent budget configuration yields zero `AWS::Budgets::Budget` resources and the stack deploys successfully.

Rollback: remove the docs, script, and optional SST wiring. No runtime impact.

## Open Questions

None for this trimmed scope.
