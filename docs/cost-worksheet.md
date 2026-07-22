# Cost Worksheet

Pricing date: 2026-07-21
Region: us-east-1
Currency: USD

> **Disclaimer.** These are estimates derived from public AWS pricing for `us-east-1` at the date above. They are not AWS Free Tier guarantees. Account credits, region, log volume, PITR, Cognito usage, alarms, custom metrics, and other services affect charges independently. Refresh the worksheet when AWS pricing changes or when the deployment region changes.

The worksheet reuses the dimension definitions from [`profiles.md`](./profiles.md). Estimates are reproducible: every number is derived from the documented pricing inputs and formulas below. The dominant cost driver is identified per profile.

## Pricing inputs (us-east-1, public list price)

| Service | Unit | Price |
| --- | --- | --- |
| DynamoDB on-demand write request | per million | $1.25 |
| DynamoDB on-demand read request | per million | $0.25 |
| DynamoDB table + index storage | per GB-month | $0.25 |
| DynamoDB PITR storage | per GB-month | $0.20 |
| Cognito MAU | per MAU-month | free under 50k |
| SSM Standard parameter | per parameter-month | $0.05 |
| Lambda request | per million | $0.20 |
| Lambda compute | per GB-second | $0.00001667 |
| SQS request | per million (first 1M free) | $0.40 |
| API Gateway HTTP API request | per million | $1.00 |
| CloudWatch Logs ingested | per GB | $0.50 |
| CloudWatch Logs stored | per GB-month | $0.03 |
| CloudWatch Logs Insights queries | per GB scanned | $0.005 |
| CloudWatch standard alarm | per alarm-month | $0.10 |
| CloudWatch custom metric | per metric-month | $0.30 |

These inputs are the only authoritative numbers in this document. Update the header date and this table when AWS pricing changes.

## Per-profile estimates

All formulas: `monthly = units_per_month × price_per_unit`. Units are derived from the profile dimensions plus the documented per-monitor and per-request assumptions below.

### Common assumptions

- Scheduler fires once per minute: `60 × 24 × 30 = 43200` invocations/month.
- Monitor checks per month = `monitor_count × (60 / cadence_minutes) × 24 × 30`.
- Per check: 1 read (`ServiceStatus`, `Monitor`) + 1 read (`RUN_REQUEST`) + 1 conditional write (`RUN_REQUEST` lease) + 1 write (`CheckRun`) + 1 update (`ServiceStatus`) ≈ 2 reads, 2 writes on the happy path; +1 write for incident transition when status changes.
- Average check duration: 800 ms, 256 MB allocated → `0.8 × 0.25 = 0.20 GB-s` per check.
- Average log line size: 1 KB. Per invocation: ~10 lines → 10 KB. Scheduler logs: ~5 KB.
- 2-week log retention applied to all application Lambda log groups owned by the stack.
- 1 fixed auth alarm (when auth deployed), 1 fixed scheduler alarm, 1 fixed heartbeat alarm. Auth alarms conditional on auth resources.
- Custom metrics: 1 fixed scheduler heartbeat + 2 fixed auth-domain when auth deployed.

### Default low-cost owner profile (10 monitors, 5-minute cadence)

- Scheduler invocations: 43200/mo.
- Monitor checks: 10 × 12 × 24 × 30 = 86400/mo.
- Reads: 86400 × 2 = 172800 ≈ 0.17M → $0.04.
- Writes: 86400 × 2 = 172800 ≈ 0.17M → $0.21.
- Lambda requests: 86400 + 43200 ≈ 130000 ≈ 0.13M → $0.03.
- Lambda compute: 86400 × 0.20 = 17280 GB-s → $0.29.
- API Gateway: dashboard + monitor-API traffic ~1M requests/mo → $1.00.
- SQS: notification + execution queues, ~130k requests → $0.05.
- DynamoDB storage (CheckRun + durable items, 30-day retention): ≈ 2 GB → $0.50.
- PITR storage: ≈ 2 GB → $0.40.
- CloudWatch Logs ingested: ≈ 1.5 GB/mo → $0.75.
- CloudWatch Logs stored (2-week retention): ≈ 1 GB × 0.5 prorated → $0.02.
- CloudWatch alarms (3 fixed): $0.30.
- CloudWatch custom metrics (1 fixed): $0.30.
- SSM parameter (1): $0.05.
- Cognito: free under 50k MAU → $0.

**Estimated monthly total: ≈ $3.94**

> Dominant cost driver: API Gateway request volume and CloudWatch Logs ingestion, not DynamoDB.

### Expected validation profile (100 monitors, 1-minute cadence)

- Scheduler invocations: 43200/mo.
- Monitor checks: 100 × 60 × 24 × 30 = 4320000/mo.
- Reads: 4320000 × 2 ≈ 8.64M → $2.16.
- Writes: 4320000 × 2 ≈ 8.64M → $10.80.
- Lambda requests: 4320000 + 43200 ≈ 4.36M → $0.87.
- Lambda compute: 4320000 × 0.20 = 864000 GB-s → $14.40.
- API Gateway: ~5M requests/mo → $5.00.
- SQS: ~4.4M requests → $1.76.
- DynamoDB storage (30-day retention, 100 monitors × 1-min cadence): ≈ 20 GB → $5.00.
- PITR storage: ≈ 20 GB → $4.00.
- CloudWatch Logs ingested: ≈ 60 GB/mo → $30.00.
- CloudWatch Logs stored (2-week): ≈ 30 GB × 0.5 → $0.45.
- CloudWatch alarms (3 fixed): $0.30.
- CloudWatch custom metrics (1 fixed): $0.30.
- SSM parameter: $0.05.
- Cognito: free.

**Estimated monthly total: ≈ $75.09**

> Dominant cost driver: Lambda compute and CloudWatch Logs ingestion. DynamoDB on-demand writes are second.

### High-volume stress profile (1000 monitors, 1-minute cadence)

- Monitor checks: 1000 × 60 × 24 × 30 = 43200000/mo.
- Reads: 86400000 ≈ 86.4M → $21.60.
- Writes: 86400000 ≈ 86.4M → $108.00.
- Lambda requests: ~43.2M → $8.64.
- Lambda compute: 43200000 × 0.20 = 8.64M GB-s → $144.00.
- API Gateway: ~25M → $25.00.
- SQS: ~43M → $17.20.
- DynamoDB storage: ≈ 200 GB → $50.00.
- PITR: ≈ 200 GB → $40.00.
- CloudWatch Logs ingested: ≈ 600 GB → $300.00.
- CloudWatch Logs stored: ≈ 300 GB × 0.5 → $4.50.
- CloudWatch alarms (3 fixed): $0.30.
- CloudWatch custom metrics (1 fixed): $0.30.
- SSM parameter: $0.05.
- Cognito: free.

**Estimated monthly total: ≈ $719.59**

> Dominant cost driver: CloudWatch Logs ingestion and Lambda compute.

## Summary

| Profile | Estimated monthly total (us-east-1) | Dominant driver |
| --- | --- | --- |
| Default low-cost owner | ≈ $3.94 | API Gateway + CloudWatch Logs |
| Expected validation | ≈ $75.09 | Lambda compute + CloudWatch Logs |
| High-volume stress | ≈ $719.59 | CloudWatch Logs + Lambda compute |

## Reproduction

Any reviewer can re-derive the per-profile numbers from this document. The only authoritative inputs are the pricing table at the top and the per-profile units derived from the profile dimensions plus the documented per-request assumptions.

When AWS pricing changes, update the pricing date in the header and the pricing table, then recompute the per-profile totals using the formulas above. The dominant driver column typically shifts at profile transitions.

## Related documents

- [`profiles.md`](./profiles.md) — dimension definitions used here.
- [`optional-budget-setup.md`](./optional-budget-setup.md) — how to wire a stage-attributed AWS Budget against these numbers.
