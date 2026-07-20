## Context

The EventBridge Scheduler execution role must send delayed escalation messages to the notification queue. Its IAM policy currently uses `JSON.stringify` on unresolved SST/Pulumi queue ARN outputs, which generates invalid resource strings and blocks staging deployment.

## Goals / Non-Goals

**Goals:**

- Render concrete queue ARNs into the scheduler execution-role policy.
- Preserve least-privilege `sqs:SendMessage` access to the notification queue and its DLQ only.
- Add a regression test for output-aware policy serialization.

**Non-Goals:**

- Change escalation scheduling behavior, queue topology, IAM permissions, or runtime code.
- Introduce dependencies or change dashboard/API behavior.

## Decisions

Use Pulumi output composition to wait for both queue ARNs before serializing the IAM policy JSON. This preserves deferred resource resolution and produces a valid AWS document at deploy time.

Direct `JSON.stringify` is rejected because it serializes unresolved outputs. Constructing ARN strings manually is rejected because it duplicates AWS resource identity and risks drift.

## Risks / Trade-offs

- [Policy rendering regresses without a deploy] -> Guard output-aware serialization in the infrastructure test suite.
- [IAM scope broadens during repair] -> Assert only notification queue and DLQ ARNs receive `sqs:SendMessage`.

## Migration Plan

1. Update policy serialization and guard coverage.
2. Run infrastructure checks and tests.
3. Redeploy staging through the lifecycle wrapper.
4. Roll back by reverting the infrastructure-only patch if deployment fails.

## Open Questions

None.
