# AWS Budget Setup

This document describes how to wire a stage-attributed AWS Budget against the cost estimates in [`cost-worksheet.md`](./cost-worksheet.md). The budget is alert-only: it never takes automatic action and never disables monitoring.

Persistent targets require budget configuration. Ephemeral targets may omit it because preview environments do not warrant a budget resource. A persistent installation that cannot use AWS Budgets must record an explicit opt-out and rationale in its target file.

## What you get

When configured, one AWS Budget scoped to the stage tags (`service` and `stage`) with two alerts:

| Alert | Threshold | Type | Meaning |
| --- | --- | --- | --- |
| Forecast | 80% of configured amount | Forecasted spend | Projected to exceed amount this month |
| Actual | 100% of configured amount | Actual spend | Current month spend equals or exceeds amount |

Both alerts notify recipients sourced from deployment configuration (never from source-controlled personal addresses). No automatic action is attached. Alerts only.

## Configuration

Add the paired fields to a persistent deployment target file at `infra/targets/<name>.target.json`:

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
| `budgetAmountUsd` | persistent targets | positive finite number | Monthly cost amount in USD. |
| `alertEmails` | persistent targets | non-empty string[] | One or more email recipients. |

Both fields must be supplied together. Partial, empty, and malformed configuration fails target validation before AWS mutation.

### Explicit persistent opt-out

Use this only when the installation deliberately cannot create AWS Budgets. The reason is required so the exception is visible in deployment configuration:

```json
{
  "budgetAlertsOptOut": true,
  "budgetAlertsOptOutReason": "The account disallows AWS Budgets for this installation."
}
```

The opt-out cannot be combined with budget fields. Ephemeral targets omit both the budget configuration and the opt-out.

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

## Rollback

For a persistent target, replace the paired budget fields with the explicit documented opt-out above and run `make deploy-infra`. The budget resource is removed; no other resources change.

## Limitations

- AWS Budgets is account-level. Multiple Bolt Monitor stages in the same account must use distinct `stage` names so their budgets do not overlap. The `TagKeyValue` filter on `service` and `stage` ensures attribution.
- AWS Budget notifications are not guaranteed to be delivered. Use the budget as one signal alongside Cost Explorer and your own monitoring.
- Forecast alerts use AWS's projected spend for the month. They are heuristic, not guarantees.

## Related documents

- [`profiles.md`](./profiles.md) — dimension definitions.
- [`cost-worksheet.md`](./cost-worksheet.md) — per-profile cost estimates and reproduction instructions.
