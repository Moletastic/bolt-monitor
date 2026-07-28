## Context

The orchestrator reads shared local `outputs.json`, optionally selects a stage entry, then permits absent `apiUrl`. Persistent checks use only table output.

## Goals / Non-Goals

**Goals:** Fail closed on missing/wrong outputs. Verify public health after every deploy. Preserve persistent protection checks.

**Non-Goals:** Add synthetic monitoring, dashboard browser tests, new AWS alarms, or change output-producing stack resources.

## Decisions

- Parse output shape explicitly and reject unscoped fallback when a stage map exists. Require `apiUrl`, `dashboardUrl`, `appTableName`, and `authTableName` appropriate to deployed stack contract.
- Run existing single public health request after output validation for persistent and ephemeral deployments.
- Keep postflight local and one-shot. It adds one request per deploy, no daemon, schedule, metric, or storage cost.

## Risks / Trade-offs

- SST output shape can vary -> Support documented flat or keyed shape with unambiguous validation only.
- Eventual endpoint readiness can fail immediately -> Surface failure for operator retry; do not hide it with unbounded retries.

## Migration Plan

Deploy stricter postflight with updated tests. Operators correct missing target outputs before retrying. No AWS data migration.
