## Why

Bolt Monitor runs as a self-deployed serverless installation with no promised SLA, RPO, or RTO. Operators running small installations (single tenant, tens of monitors, one region) need a reproducible cost model and an optional budget guardrail so a misconfigured schedule, an unbounded query, or an unexpected retention change does not produce a surprise bill. Recovery drill tooling, full data inventory, and bounded scheduler/API refactors are deferred until a profile demands them.

This trimmed change ships only the FinOps foundation: three named profiles, a reproducible cost worksheet, and documented optional budget setup that does not block clean deployment.

## What Changes

- Document three named installation profiles (default low-cost owner, expected validation, high-volume stress) and their dimensions, so the cost worksheet, support boundaries, and stress evidence share one vocabulary.
- Produce a reproducible cost worksheet as a static markdown document under `docs/`, covering `AppTable` and `AuthTable` on-demand reads/writes, storage, indexes, PITR, Cognito, SSM, Lambda, SQS, API Gateway, logs/alarms, and dashboard attribution. Identify the dominant cost driver per profile. Pricing date and region are recorded in the document header. Disclaimer: low-cost estimate is not a Free Tier guarantee.
- Document an optional stage-attributed AWS Budget setup (forecast at 80%, actual at 100%, no automatic shutdown, recipients from deployment configuration not source-controlled). Absent configuration creates no budget and does not fail deployment.
- Verify that absent budget configuration yields no AWS Budget resource and does not block the default deploy path.

## Capabilities

### New Capabilities

- `cost-and-budget-guardrails`: Named profiles, reproducible cost worksheet, and optional stage-attributed AWS Budget setup with documented no-op default.

## Impact

- Adds documentation under `docs/` (profiles, cost worksheet, optional budget setup).
- Adds optional, opt-in infrastructure wiring for one stage-attributed AWS Budget that is gated by target configuration and stays absent otherwise.
- No runtime code, API, dashboard, scheduler, or persistence changes.
- No data inventory, retention class changes, recovery runbook, scheduler projection, or bounded API refactor in this change.
