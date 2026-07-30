## Context

Manual runs reserve an idempotency record before executing, but retries return different bodies. Replay reads a case-sensitive header. Test sends and incident commands can repeat external or stateful work after a lost response.

## Goals / Non-Goals

**Goals:** one scoped key and canonical request return one durable outcome; records expire through DynamoDB TTL; conflicts are typed.

**Non-Goals:** idempotency for every mutation, a generic workflow engine, or indefinite response retention.

## Decisions

- Store command state and canonical public result in one generic TTL-bounded idempotency record. Each record has `tenantId`, `operation`, `resourceId`, `idempotencyKey`, canonical request fingerprint, `pending` or `completed` state, sanitized public response, absolute TTL, and optional `runId`. This makes retries deterministic without redoing side effects.
- Use `requestHeader` for all idempotency extraction. API Gateway header casing is not contractually stable.
- Add command support incrementally: manual run, replay compatibility, test send, ack, resolve. Resource creation and monitor toggles remain later work.
- Return `409 IDEMPOTENCY_CONFLICT` for same scoped key with different fingerprint. Do not use validation failure.

## Risks / Trade-offs

- Response storage increases DynamoDB writes/storage -> retain only small sanitized response data with bounded TTL.
- A command can remain pending after a process failure -> expose stable accepted/pending result; a retry MUST NOT resume an external side effect or an incident mutation. Operators can use a new key after inspecting state when recovery requires another attempt.
- Notification providers cannot guarantee exactly-once delivery after a timeout -> reserve the idempotency record before calling a provider, then return the stored pending result on retry. This favors no duplicate sends over automatic recovery.

## Migration Plan

Deploy readers compatible with existing manual-run records, then write generic command records with canonical result fields. Existing records retain current expiry. Incident commands reserve their record and durable incident mutation in one DynamoDB transaction. Rollback preserves records; handlers continue safe no-duplicate behavior.

## Open Questions

None.
