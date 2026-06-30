## Context

System now has monitor configuration, probe-location catalog, and CRUD work queued, but no execution engine. Product cannot claim to monitor anything until configured monitors are actually executed. At same time, periodic monitoring can create accidental cost if the project enables constant checks without an immediate way to stop them.

## Goals / Non-Goals

**Goals:**
- Define how enabled monitors become scheduled and executed checks.
- Define routing of checks to selected probe locations.
- Normalize execution output for downstream result/status storage.
- Guarantee an operational stop path before periodic healthchecks are enabled.

**Non-Goals:**
- Full alerting or incident generation.
- Frontend dashboard work.
- Authentication or multi-tenant access controls.
- Vendor-specific worker infrastructure details beyond probe-location routing needs.

## Decisions

### Disabled monitors must never execute
- Decision: monitor `enabled=false` is hard gate that prevents scheduling and execution.
- Rationale: cheapest, clearest operational kill switch already present in core monitor lifecycle.
- Alternative considered: allow scheduler to continue but suppress notifications only.
- Why not: does not stop cost or probe traffic.

### Do not enable periodic execution without stop control
- Decision: periodic execution must not ship unless operators can stop checks immediately through monitor disablement or equivalent global kill path.
- Rationale: development-phase cost safety is mandatory.
- Alternative considered: ship periodic loop first, add stop later.
- Why not: unacceptable billing and operational risk.

### Normalize execution output before persistence
- Decision: workers emit protocol-agnostic normalized result envelope consumed by result/status model.
- Rationale: HTTP/TCP/gRPC checks need one downstream pipeline shape.
- Alternative considered: protocol-specific result shapes written directly.
- Why not: complicates status, incidents, and reads.

### Probe-location routing comes from system catalog + monitor selection
- Decision: execution target derives from monitor-selected `probeLocations` validated against system catalog.
- Rationale: keeps execution vendor-neutral and consistent.
- Alternative considered: infer location at runtime without stored selection.
- Why not: weak user control and unclear routing.

## Risks / Trade-offs

- [Risk] Scheduler design may overfit current dev environment. -> Mitigation: keep contract about selection/routing/result normalization, not infrastructure brand.
- [Risk] Periodic execution can create surprise cost. -> Mitigation: require disable/stop semantics before shipping recurring runs.
- [Risk] Multi-protocol execution could sprawl quickly. -> Mitigation: keep pipeline generic but start with one protocol implementation first if needed.

## Migration Plan

1. Define execution selection, routing, and normalized output contracts.
2. Implement safety gate using monitor enable/disable lifecycle before recurring runs are active.
3. Connect execution results to result/status persistence in follow-up or paired work.

Rollback is straightforward: stop scheduler and worker routing while retaining monitor CRUD and configuration state.

## Open Questions

- Should v1 execution start with HTTP only, or define TCP/gRPC contracts immediately?
- Is monitor disable enough as stop control, or do we also want a global pause switch?
- Should manual run and periodic run share same execution path from day 1?
