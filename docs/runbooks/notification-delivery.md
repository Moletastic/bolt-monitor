# Notification Delivery Operations Runbook

This runbook covers the notification delivery surface added by
`assure-notification-and-escalation-delivery`. Use it to correlate
delivery outcomes, recover from poison work, and validate the safety
boundaries that protect credential material and queue durability.

## Concepts

- **Transition identity.** Each accepted incident down/up change creates
  one canonical `TransitionOutbox` row. The dispatch path is the
  sole writer of the notification queue. The outbox `eventId` is also
  the `transitionId` used by deliveries, schedules, and dashboards.
- **Delivery identity.** One `dlv_…` identity per incident × transition
  × step × channel. Replays never change identity; they re-queue the
  same record through the canonical dispatcher.
- **Delivery state.** Exactly one of `pending`, `in_flight`,
  `retryable_failed`, `ambiguous`, `delivered`, `terminal_failed`.
  Suppression is escalation eligibility, never a delivery state.
- **Provider acceptance.** `delivered` means the configured provider
  accepted the request according to its API. The system does not claim
  a human read or acted on the notification.

## Common Tasks

### List deliveries for an incident

```sh
curl -fsS \
  -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/incidents/$INCIDENT/deliveries"
```

Order is stable: created-at ascending with `deliveryId` as tie-breaker.
Use the response to grep `state` and `lastOutcomeClass`.

### Replay a terminal failure

```sh
curl -fsS \
  -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $(uuidgen)" \
  "$API_BASE/api/v1/incidents/$INCIDENT/deliveries/$DELIVERY/replay"
```

- `DELIVERY_NOT_REPLAYABLE` → state is not `terminal_failed`. Inspect
  the current state; do not retry.
- `IDEMPOTENCY_CONFLICT` → idempotency key was reused with a different
  request body. Generate a fresh key.
- Cross-tenant access returns `INCIDENT_NOT_FOUND`; no record is
  mutated.

### Correlate delivery with incident activity

`GET /api/v1/incidents/{id}/activities` returns rows whose
`activityId` equals the canonical `transitionId`. The deliveries
endpoint surfaces the same `transitionId` per row. Join them
client-side to render a single timeline.

## Notification DLQ (Source-Kind Aware)

Every DLQ envelope carries `sourceKind`:

- `dynamodb_stream` — canonical dispatch exhausted retries. Do not
  re-enqueue to the notification queue. Use the runbook section
  **Stream reconciliation** below.
- `scheduler_target` — Scheduler retry policy exhausted for a delayed
  step. Re-enqueue only after the operator has re-validated the
  scheduled-step identity and incident/escalation eligibility.
- `notification_sqs` — repeated SQS-side poison (malformed envelope,
  unknown version/kind, repeated retryable provider outcome). Inspect
  the body; redrive only after root cause is fixed.

Unrelated envelopes (Lambda failures outside this change) share the
queue but use different kinds and may not be redrivable. Confirm the
shape before any redrive.

## Stream Reconciliation (Bounded Recovery)

The Stream dispatcher attempts to ack only after SQS acceptance.
Exhaustion writes a source-tagged envelope to the notification DLQ.
The record remains `pending` in the sparse bucketed access path and
must be reconciled, never silently re-acknowledged.

### Recent bucket (preferred, bounded)

The reconciler reads configured recent buckets and pages only; it
never scans. Verify it ran on schedule:

```sh
# Operator-supplied command (script provided by this change)
./tools/admin-bootstrap/notification-reconciler --tenant DEFAULT \
    --buckets 4 --shards 4 --page-limit 50
```

### Point repair by `eventId`

For a single identified event, run:

```sh
./tools/admin-bootstrap/notification-repair --tenant DEFAULT \
    --event-id "$EVENT_ID"
```

The repair path uses a point lookup against `TransitionOutbox` and
returns the same conditional acknowledgement as Stream. It does not
scan unrelated rows.

## Manual Repair for Provider Misconfiguration

Channel credentials are resolved at every attempt, not at delivery
creation. Replays pick up corrected configuration automatically.

1. `UpdateNotificationChannel` with corrected config.
2. Verify via `POST /api/v1/notification-channels/{id}/test` and confirm
   the sanitized outcome shows `accepted`.
3. Replay eligible `terminal_failed` deliveries one channel at a time.

## Ambiguous Outcome Handling

`ambiguous` indicates the provider may have received the request but
did not confirm acceptance. The system intentionally preserves the
delivery record and discloses duplicate-side-effect risk. Operators
must:

1. Inspect `providerRequestId` and the provider audit log to determine
   whether the request landed.
2. Decide replay-vs-do-not based on idempotency support. Senders pass
   the `deliveryId` as the idempotency key when the provider contract
   supports it.
3. Do not replay `ambiguous` deliveries through the API; the API only
   replays `terminal_failed`. Manual reconciliation is required.

## Notification-DLQ Redrive Restrictions

Source-aware redrive:

- `dynamodb_stream` → run reconciliation or point repair. **Never**
  re-enqueue the Stream envelope directly to the notification queue.
- `scheduler_target` → confirm eligibility, then re-enqueue the
  canonical scheduled-step envelope only.
- `notification_sqs` → only after the underlying envelope validates.

Unknown kinds: quarantine, do not blind-redrive.

## Rollback

Producer rollback pauses recurring and manual execution before
restoring prior code. Pending outbox and delivery records are retained
for the dispatcher and are not deleted during rollback. New-format
items remain durable under bounded retention.

The Scheduler and Stream subscriptions must be disabled before the
prior runtime is restored to avoid duplicate work.

## Logs and Metrics

Structured logs include `tenantId`, `incidentId`, `transitionId`,
`deliveryId`, `stepNumber`, `channelType`, `attempt`, and
`outcomeClass`. Metrics are namespace `BoltMonitor/Notifications` with
the following dimensions:

- `OutcomeClass`: counts by `accepted`, `timeout`, `transport`,
  `throttled`, `provider_5xx`, `provider_4xx`, `invalid_config`,
  `unsupported_channel`, `retry_exhausted`.
- `State`: per-state counts.
- `SourceKind`: dispatch attempts per source.

The DLQ depth alarm `EscalationTransitionDispatchAlarm` alerts on any
message visible in the notification DLQ.

## Cost

See `docs/notification-delivery-cost.md` for the bounded incremental
cost of Stream invocations, pending-index reconciliation,
delivery/replay/idempotency writes, retries, and short-lived schedules.

## Legacy Cleanup (One-Off)

Legacy `esc-*-step-*` EventBridge rules and `allow-events-*` Lambda
permission statements are created by the prior scheduler. The
`./tools/admin-bootstrap/legacy-eventbridge-cleanup` tool removes
matching resources in dry-run mode by default. The runbook is in
`docs/legacy-eventbridge-cleanup.md`.
