## Why

The current notification system sends a flat `incident.opened` event to a single Telegram channel when an incident opens, with no concept of ordering, delays, or escalation paths. This means critical incidents ring every on-call engineer simultaneously with no prioritization, and there's no mechanism to escalate to secondary responders if the primary is unreachable. The system also has no concept of business-hours-aware routing or escalation exhaustion handling.

## What Changes

- **Escalation policies**: New CRUD API for managing reusable escalation policy templates containing ordered notification steps
- **Service binding**: Services reference an escalation policy; each service can have a different escalation path
- **Business hours / off-hours paths**: Each policy contains two paths — one for business hours, one for off-hours — with different step sequences and delays
- **Sequential step execution**: Steps fire in order with configurable delays (0 = immediate). No branching logic — always sequential.
- **Escalation state persistence**: DynamoDB tracks which step is active/scheduled for each open incident, with CloudWatch scheduled rules for delayed invocations
- **EscalationExhausted incident**: When all steps fire and the incident is still open, a new incident type `escalation.exhausted` is created linking back to the original
- **Notification channels replaced**: The existing notification channel model is removed entirely; all notifications flow through escalation policies
- **DOWN/UP trigger scope**: Escalation starts on DOWN (incident open), stops on UP (incident resolved). Architecture is designed to make adding DEGRADED/RECOVERING trigger support straightforward.

## Capabilities

### New Capabilities

- `escalation-policy-crud`: Full CRUD API for creating, reading, updating, and deleting escalation policies. Policies contain business-hours and off-hours paths with ordered steps and channel targets.
- `escalation-path-execution`: Runtime execution engine that fires steps sequentially, persists state in DynamoDB, uses CloudWatch scheduled rules for delayed invocations, and suppresses remaining steps when incident resolves.
- `escalation-exhausted-incident`: New incident type created when an escalation path completes all steps while the original incident remains open. Contains a link to the original incident for traceability.
- `service-escalation-binding`: Service model gains an optional `escalationPolicyId` reference and `businessHours` configuration (timezone, start/end hour, days of week). Escalation path selection is determined by whether current time falls within business hours.

### Modified Capabilities

- `incident-management-api`: Incident lifecycle changes — when escalation path exhausts all steps and incident is still open, system creates an `escalation.exhausted` incident. Incident notification events are no longer emitted directly by check-runtime; they are driven by escalation step execution.
- `notifications`: Existing notification channel model (Telegram, email, etc. channel records) is removed. Notification sending is entirely driven by escalation policy steps.

## Impact

- **API**: `services/monitor-api/` — new escalation policy CRUD endpoints, service binding fields
- **Check runtime**: `services/check-runtime/` — no longer emits `incident.opened`/`incident.resolved` directly; emits `incident.down`/`incident.up` events that trigger escalation evaluation
- **Notify runtime**: `services/notify-runtime/` — generalizes from Telegram-only to multi-channel (telegram, email, sms, webhook, pagerduty)
- **DynamoDB**: New entity types `ESCALATION_POLICY` and `ESCALATION_STATE`. Removal of `NOTIFICATION_CHANNEL` entity type.
- **Shared domain**: `shared/` — new escalation domain models
- **Dashboard**: `apps/dashboard/` — escalation policy management UI, escalation state display on incidents
