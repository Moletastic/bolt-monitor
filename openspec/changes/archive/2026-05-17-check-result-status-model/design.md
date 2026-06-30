## Context

Monitor CRUD and execution pipeline create configuration and runtime activity, but neither provides durable monitoring state without result persistence. System needs clear distinction between append-only run history and latest status snapshot so dashboards and incidents do not rebuild everything from scratch.

## Goals / Non-Goals

**Goals:**
- Define canonical `CheckRun` model for raw execution results.
- Define canonical `MonitorStatus` model for latest status snapshot.
- Define mapping of both models into single-table DynamoDB item families.
- Clarify TTL expectations for high-volume run data.

**Non-Goals:**
- Full incident generation or alert thresholds.
- Frontend rendering details.
- Long-term analytics rollups beyond retention guidance.

## Decisions

### Separate raw history from latest state
- Decision: store append-only run records separately from one mutable latest-status record.
- Rationale: balances auditability and fast reads.
- Alternative considered: compute status dynamically from runs only.
- Why not: expensive and slow for dashboard/API reads.

### Keep raw runs high-volume and expiring
- Decision: raw `CheckRun` items should carry TTL and limited retention expectations.
- Rationale: constant healthchecks will create large volume quickly.
- Alternative considered: retain all raw runs indefinitely.
- Why not: cost grows too fast for little short-term value.

### Status stays minimal before incidents land
- Decision: `MonitorStatus` should focus on current operational summary, not full incident semantics.
- Rationale: keeps model usable for read APIs now and extensible later.
- Alternative considered: bake incident fields directly into status from start.
- Why not: couples status too tightly to future alerting work.

## Risks / Trade-offs

- [Risk] Too little status data may limit early dashboard reads. -> Mitigation: include core fields like state, last checked time, latency, and last error summary.
- [Risk] TTL window may be wrong initially. -> Mitigation: document policy and keep it adjustable.
- [Risk] Result schema may evolve with TCP/gRPC details. -> Mitigation: keep canonical common fields plus protocol-specific extension area.

## Migration Plan

1. Define `CheckRun` and `MonitorStatus` models and item mappings.
2. Implement writers from execution pipeline to these models.
3. Build read APIs and incidents on top of these persisted records.

Rollback is low-risk because this is contract-first work before large production history exists.

## Open Questions

- Should `MonitorStatus` include rolling uptime percentages now or later?
- How much protocol-specific detail belongs in common `CheckRun` versus nested extension fields?
- What exact TTL window should raw runs use in v1?
