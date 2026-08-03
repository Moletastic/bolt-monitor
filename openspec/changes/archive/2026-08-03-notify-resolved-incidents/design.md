## Context

The escalation runtime currently treats `incident.up` only as cancellation: it marks the incident's escalation state suppressed and exits. The successful recovery transition has a canonical outbox identity but the runtime neither creates recovery delivery records nor calls notification providers.

Recovery must stop delayed escalation work while notifying recipients who were contacted about the outage. Delivery safety requires recovery work to use its own canonical transition identity; reusing the incident ID would collide with outage delivery identities.

## Goals / Non-Goals

**Goals:**

- Send one recovery notification to each channel in every escalation step that fired for the incident.
- Persist and conditionally claim recovery deliveries before provider I/O, using the recovery transition identity.
- Preserve suppression of unfinished delayed steps, including recovery-before-schedule and duplicate-work races.

**Non-Goals:**

- Notify channels in steps that never fired.
- Replay historical recoveries or add operator-triggered recovery broadcasts.
- Change provider-specific message formats, policy APIs, or delivery-history APIs.

## Decisions

### Notify channels from fired steps only

Recovery follows the incident's persisted escalation state (`stepsFired`) and selected policy path. It notifies each unique channel selected by a fired step, rather than every channel in the policy. This reaches people and provider integrations that received the outage while avoiding recovery noise for escalation levels never reached.

Alternative: notify only step one. Rejected because a PagerDuty or other channel first used in a later fired step would remain open. Alternative: notify every policy channel. Rejected because it announces an incident to recipients never notified of it.

### Recovery is a distinct durable transition

Carry the canonical outbox `eventId` into the runtime event as an optional transition identity. Legacy queue events fall back to the incident ID. Recovery delivery identities derive from this transition ID, channel, and a stable recovery step position, preventing collisions with outage deliveries and retaining duplicate-message safety.

Alternative: reuse the incident ID. Rejected because delivery records are transition-scoped and would collide with existing outage records.

### Suppress after durable recovery work is prepared

The runtime resolves policy channels and creates/claims recovery deliveries before marking escalation state suppressed. It then suppresses the state regardless of whether recovery has zero eligible channels. Scheduled work still rechecks resolved incident state before sending, so it cannot race through after recovery.

Alternative: suppress before recovery work. Rejected because existing eligibility checks would treat recovery work as suppressible and skip it.

## Risks / Trade-offs

- [Policy or channel changes after outage] → Recovery resolves the current channel references from the selected policy; deleted channels are skipped and existing delivery history remains intact.
- [A recovery delivery fails] → Existing durable claim, retry, DLQ, and provider outcome behavior applies; suppression still prevents new outage escalation.
- [Duplicate or reordered queue events] → Canonical recovery transition identity and conditional delivery claims make sends idempotent; incident-state checks continue blocking stale scheduled work.

## Migration Plan

1. Deploy runtime code and tests; no data migration is needed.
2. New recoveries create distinct delivery records and notify channels used by fired outage steps.
3. Roll back by redeploying prior runtime; existing recovery delivery history remains queryable and incomplete recovery queue work follows normal retry/DLQ handling.
