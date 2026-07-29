## Context

One DynamoDB Stream consumer publishes canonical transition outbox records to notification SQS. It currently acknowledges after enqueue, although SQS consumer acknowledgement is the delivery boundary. Envelope `runId` contains causal run identity while outbox lookup needs event identity.

## Goals / Non-Goals

**Goals:** Preserve one processable pending outbox identity until transition handling completes. Cover Stream-to-SQS path end to end. Reduce retry/DLQ waste without changing AWS topology.

**Non-Goals:** Exactly-once provider delivery, new queues, new indexes, or historical outbox repair.

## Decisions

- Stream dispatcher sends canonical `transitionId` and does not acknowledge. SQS consumer uses `transitionId`, processes only pending record, then conditionally acknowledges. This keeps acknowledgement at completed downstream handling boundary.
- Preserve causal `runId` in envelope for traceability but never use it as outbox key. Avoids corrupting existing causal model.
- Reuse existing sparse pending-record reconciliation. No Scan, GSI, queue, or metric addition; retry reduction lowers Lambda/SQS/DLQ consumption.

## Risks / Trade-offs

- Duplicate SQS enqueue after ambiguous Stream send -> Consumer conditional acknowledgement and delivery identities converge safely.
- Existing records acknowledged by broken behavior remain skipped -> Repair remains operator-driven through existing reconciliation tools.

## Migration Plan

Deploy consumer/dispatcher fix together. Verify one synthetic transition reaches the notification path and pending records drain. Rollback restores prior binary only; no schema migration exists.
