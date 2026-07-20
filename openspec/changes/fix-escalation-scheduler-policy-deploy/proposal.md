## Why

Staging deployment fails when AWS rejects the EventBridge Scheduler execution-role policy. The policy attempts to serialize unresolved SST resource outputs, producing an invalid ARN instead of SQS queue ARNs.

## What Changes

- Resolve notification queue ARN outputs before serializing the scheduler execution-role policy.
- Add a regression guard for the rendered policy document so staging deploy receives valid SQS resource ARNs.

## Capabilities

### New Capabilities

- `escalation-scheduler-deployment`: Ensures escalation scheduling infrastructure renders a deployable least-privilege SQS execution policy.

### Modified Capabilities

None.

## Impact

- Affects EventBridge Scheduler IAM policy construction in `infra/stacks/bootstrap.ts` and its infrastructure guard tests.
- No API, dashboard, runtime, data-model, or dependency changes.
