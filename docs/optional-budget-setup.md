# Optional AWS Budget Setup

This document describes how to wire an optional stage-attributed AWS Budget against the cost estimates in [`cost-worksheet.md`](./cost-worksheet.md). The budget is alert-only: it never takes automatic action and never disables monitoring.

The wiring is opt-in via the deployment target file. When the target omits budget configuration, no `AWS::Budgets::Budget` resource is provisioned and deployment succeeds unchanged. A clean account without budget permissions can install Bolt Monitor.

## What you get

When configured, one AWS Budget scoped to the stage tags (`service` and `stage`) with two alerts:

| Alert | Threshold | Type | Meaning |
| --- | --- | --- | --- |
| Forecast | 80% of configured amount | Forecasted spend | Projected to exceed amount this month |
| Actual | 100% of configured amount | Actual spend | Current month spend equals or exceeds amount |

Both alerts notify recipients sourced from deployment configuration (never from source-controlled personal addresses). No automatic action is attached. Alerts only.

## Configuration

Add the optional fields to the deployment target file at `infra/targets/<name>.target.json`:

```json
{
  "stage": "staging",
  "profile": "bolt-monitor",
  "accountId": "123456789012",
  "region": "us-east-1",
  "lifecycle": "persistent",
  "owner": "Your Team",
  "service": "bolt-monitor",
  "dashboardOrigin": "https://staging.example.com",
  "approved": true,
  "budgetAmountUsd": 10,
  "alertEmails": ["ops@example.com"]
}
```

| Field | Required | Type | Meaning |
| --- | --- | --- | --- |
| `budgetAmountUsd` | optional | number | Monthly cost amount in USD. |
| `alertEmails` | optional | string[] | One or more email recipients. |

Both fields must be present and non-empty for the budget resource to be provisioned. If either is missing, the field is treated as absent, no budget is created, and no error is raised.

## Choosing the amount

Start from the per-profile estimates in [`cost-worksheet.md`](./cost-worksheet.md):

| Profile | Worksheet estimate | Suggested budget amount |
| --- | --- | --- |
| Default low-cost owner | ≈ $3.94 | $10 |
| Expected validation | ≈ $75.09 | $100 |
| High-volume stress | ≈ $719.59 | $800 |

Adjust the amount above the worksheet estimate so a forecast alert at 80% gives lead time before the actual limit. Refresh the worksheet when AWS pricing changes; the budget amount and the worksheet stay independent.

## Verification

After `make deploy-infra`, verify the budget was provisioned:

```bash
aws budgets describe-budgets --account-id <account-id> --query "Budgets[?BudgetName=='bolt-monitor-<stage>-monthly']"
```

The result should include:

- `BudgetName` matching `bolt-monitor-<stage>-monthly`.
- `BudgetLimit.Amount` matching the configured `budgetAmountUsd`.
- `BudgetLimit.Unit` set to `USD`.
- `CostFilters` including `TagKeyValue: 'service$<service>'` and `TagKeyValue: 'stage$<stage>'`.
- Two `Notification` entries: `FORECASTED` at 80% and `ACTUAL` at 100%.
- `Subscribers` containing the configured email addresses.

Verify the alert path before relying on it. From the AWS console or CLI:

1. Temporarily set `budgetAmountUsd` to a value below the current month's spend.
2. Redeploy. Within one billing cycle, AWS should deliver an `ACTUAL` alert to the configured recipients.
3. Restore the desired amount and redeploy.

If no recipient receives the alert, fix the subscription before treating the budget as operational.

If budget configuration is absent, run the same query: the result should be empty for the `bolt-monitor-<stage>-monthly` budget name.

## No-op default

The repository deploys cleanly without any budget configuration. The SST wiring conditionally creates the `AWS::Budgets::Budget` resource only when both `budgetAmountUsd` and `alertEmails` are present and non-empty in the target file. Missing fields produce no resource and no deployment error.

The deployment target validator does not require budget fields. Targets without budget configuration pass `validateDeploymentTarget` unchanged.

## Rollback

Remove the `budgetAmountUsd` and `alertEmails` fields from the target file and run `make deploy-infra`. The budget resource is removed; no other resources change.

## Limitations

- AWS Budgets is account-level. Multiple Bolt Monitor stages in the same account must use distinct `stage` names so their budgets do not overlap. The `TagKeyValue` filter on `service` and `stage` ensures attribution.
- AWS Budget notifications are not guaranteed to be delivered. Use the budget as one signal alongside Cost Explorer and your own monitoring.
- Forecast alerts use AWS's projected spend for the month. They are heuristic, not guarantees.

## Related documents

- [`profiles.md`](./profiles.md) — dimension definitions.
- [`cost-worksheet.md`](./cost-worksheet.md) — per-profile cost estimates and reproduction instructions.
