## Context

Delivery plan/claim/complete persistence exists but initial and scheduled handlers send first. At-least-once SQS then duplicates external notifications after post-send failure.

## Goals / Non-Goals

**Goals:** Fence provider attempts with durable delivery state. Preserve sanitized outcomes and bounded retry behavior. Reduce duplicate provider, Lambda, and SQS cost.

**Non-Goals:** Guarantee exactly-once effects at providers without idempotency support, add tables/GSIs, or change public delivery state names.

## Decisions

- Create deterministic plan/delivery records before provider I/O. Claim one eligible delivery with lease/fencing token; only claimant sends.
- Complete each claimed delivery after sender result. Delivered records skip retries. Uncertain post-send outcomes remain ambiguous, not falsely delivered.
- Resolve each step's channels from durable identities. Existing DynamoDB item model and notification queue remain; extra conditional writes trade small DynamoDB cost for far lower duplicate provider/SQS retry cost.

## Risks / Trade-offs

- Provider accepts then persistence fails -> state is ambiguous; later retry can still duplicate where provider lacks dedup support.
- More conditional writes -> bounded per attempted channel; no GSI/scan/new resource cost.

## Migration Plan

Deploy writer and consumer together. Existing pending records remain eligible. Observe delivery state transitions and duplicate-alert reports. Rollback retains records safely; old code may not honor them.
