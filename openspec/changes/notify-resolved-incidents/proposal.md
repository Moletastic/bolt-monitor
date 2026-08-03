## Why

Operators receive incident-open notifications but never learn when incidents recover. The escalation runtime treats an `incident.up` transition only as a signal to suppress remaining delayed work, so no provider request is made for a resolved incident.

## What Changes

- Deliver one recovery notification for an `incident.up` transition through channels selected by the incident's escalation policy.
- Preserve recovery suppression so delayed escalation steps cannot notify after the incident has resolved.
- Persist and claim recovery deliveries with transition-scoped deterministic identities before provider I/O, preserving retry and duplicate-delivery safety.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `notification-delivery-assurance`: Recovery transitions notify configured policy channels while still suppressing unfinished escalation work.

## Impact

- `services/escalation-runtime` incident-up handling and delivery orchestration.
- Escalation runtime regression coverage.
- Existing notification providers and DynamoDB delivery records; no API or infrastructure contract changes.
