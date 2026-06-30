## Context

The current notification system (`shared/notifications/`) emits `incident.opened` and `incident.resolved` events directly from `check-runtime` to a Telegram channel. There is no ordering, no delays, and no escalation path. The system also lacks business-hours awareness and multi-channel support (Telegram-only today).

The `monitor-state-machine` change (in progress) introduces DEGRADED/RECOVERING states and configurable failure/recovery thresholds. This escalation change builds on that — when a monitor goes DOWN and opens an incident, the escalation system takes over notification delivery.

## Goals / Non-Goals

**Goals:**
- Provide ordered notification steps with configurable delays per escalation policy
- Support business-hours and off-hours escalation paths within the same policy
- Persist escalation state so delayed steps survive Lambda restarts
- Replace the existing flat notification model entirely with escalation-driven notifications
- Support multiple notification channels: Telegram, Email, SMS, Webhook, PagerDuty

**Non-Goals:**
- Probe-location-aware escalation (separate change)
- Automatic maintenance-window scheduling (manual only)
- Branching/conditional escalation logic (always sequential steps)
- Escalation triggered on DEGRADED or RECOVERING states (only DOWN/UP for v1, architecture supports future extension)

## Decisions

### Decision 1: Escalation Policy Storage — DynamoDB Single-Table

**Choice**: Store `EscalationPolicy` as a DynamoDB record under entity type `ESCALATION_POLICY` using the existing single-table schema pattern.

**Rationale**: Consistent with existing entity storage pattern. Allows co-location with other entity types on the same table, leveraging existing repository patterns.

**Key design**: Escalation state (`ESCALATION_STATE`) is a separate entity type tracking per-incident progression through steps.

### Decision 2: Delayed Step Scheduling — CloudWatch Scheduled Rules

**Choice**: Use CloudWatch EventBridge scheduled rules for delayed step invocations, with escalation state persisted in DynamoDB.

**Rationale**:
- In-memory scheduling (Lambda cold start storage) dies on freeze/restart
- SQS delay queues max out at 15 minutes per hop; chaining for 1hr delays is complex
- DynamoDB TTL alone requires a poller and has inherent delay variability
- CloudWatch scheduled rules are persistent AWS primitives that survive Lambda restarts

**Flow**:
1. Incident DOWN → lookup service's escalation policy → create `ESCALATION_STATE` record
2. Fire Step 1 immediately → emit notifications
3. Schedule CloudWatch rule for Step 2 with `delayMinutes`
4. [Lambda freezes/restarts] → CloudWatch rule still fires at scheduled time
5. Escalation handler Lambda invoked → reads `ESCALATION_STATE` → fires Step 2
6. Repeat until all steps exhausted or incident resolves

**Alternatives considered**: DynamoDB TTL + polling (lower reliability, higher cost), SQS delay chaining (complex, max 15min per hop).

### Decision 3: Notification Channels — Inline in Policy Steps

**Choice**: Channel configuration (type, target, config) is embedded directly in each escalation step, not stored as separate `NotificationChannel` records.

**Rationale**: Simplifies the model — no separate channel CRUD, no indirection. The current `NOTIFICATION_CHANNEL` entity type is removed entirely. Operators configure channels when creating escalation policies.

**Channel types supported**: `telegram`, `email`, `sms`, `webhook`, `pagerduty`.

### Decision 4: Service Binding — Explicit Policy ID

**Choice**: Service model gets an `escalationPolicyId` field. When set, the service uses that policy. When null, no escalation occurs for that service's incidents.

**Rationale**: Explicit over implicit. Operators always know which policy governs a service. Simpler to audit and debug than convention-based matching.

**Hybrid option deferred**: Could add `useDefaultPolicy` convention later if needed (service `payments` defaults to policy `payments-escalation`).

### Decision 5: Business Hours Determination — Service-Level Config

**Choice**: Each service has a `businessHours` config: `{ timezone, startHour, endHour, daysOfWeek }`. Escalation handler determines current path by evaluating whether current time falls within the configured window.

**Rationale**: Business hours are a service-level concern (payments team may work different hours than platform team). Per-policy business hours would require the same config in every policy.

### Decision 6: Escalation State Machine

```
ESCALATION_STATE per open incident:
  - incidentId, policyId, currentStep, stepsFired[]
  - status: ACTIVE | SUPPRESSED | EXHAUSTED

Incident DOWN → EscalationState created (status=ACTIVE, currentStep=1)
  ↓
Step N fires → Update EscalationState.stepsFired, advance currentStep
  ↓
All steps exhausted + incident still open → status=EXHAUSTED, create EscalationExhausted incident
  ↓
Incident UP → status=SUPPRESSED (remaining steps never fire)
```

## Risks / Trade-offs

[Risk] CloudWatch rule orphaning if Lambda fails to process
→ Mitigation: Escalation handler re-validates incident state before firing step. If already resolved/acknowledged in a way that should suppress, it skips the step.

[Risk] Clock skew between Lambda timezone and business hours config
→ Mitigation: Business hours evaluation uses the configured timezone explicitly. Store timestamps in UTC, convert for comparison.

[Risk] Long delay between steps with Lambda costs
→ Mitigation: Lambda is only invoked at step boundaries, not continuously. Cost is minimal for step-wise execution.

[Risk] EscalationExhausted incident creates infinite loop
→ Mitigation: EscalationExhausted incident does NOT trigger escalation itself. It is terminal — requires manual resolution.

[Risk] Policy deletion while incidents are active
→ Mitigation: Policy deletion is rejected if any service references it. Operator must reassign services first.

## Migration Plan

1. **Phase 1 — Schema and models**: Add `EscalationPolicy` and `EscalationState` entity types. Add `escalationPolicyId` and `businessHours` to Service model. No behavior change yet.
2. **Phase 2 — CRUD API**: Implement escalation policy CRUD endpoints. Implement service binding. Existing notification system still operates in parallel.
3. **Phase 3 — Notify runtime generalization**: Extend notify-runtime to support multiple channel types (email, sms, webhook, pagerduty) in addition to Telegram.
4. **Phase 4 — check-runtime integration**: check-runtime emits `incident.down` and `incident.up` events instead of directly creating/resolving incidents. Escalation handler receives these events and starts/suppresses escalation.
5. **Phase 5 — Escalation execution engine**: Implement step firing, CloudWatch scheduling, state persistence, and EscalationExhausted incident creation.
6. **Phase 6 — Cleanup**: Remove `NOTIFICATION_CHANNEL` entity type and related code. Remove direct `incident.opened`/`incident.resolved` notification emission from check-runtime.

**Rollback**: Phase 1-3 can be rolled back without breaking incidents. Phase 4-6 require coordinated deployment since they change the incident lifecycle.

## Open Questions

1. **What is the format of the EscalationExhausted incident `linkedIncidentId`?** Is it a new field on the incident record, or a separate entity type linking two incidents?

2. **Should acknowledgement pause the escalation clock?** Currently not — escalation continues regardless of ack status. But the design supports adding this as a future condition.

3. **What is the max number of steps per escalation path?** Should there be a hard limit (e.g., 10) to prevent unbounded escalation chains?

4. **Do escalation policies have a version/history?** If an operator edits a policy while incidents are active, should active escalations continue with the old policy or pick up the new one?
