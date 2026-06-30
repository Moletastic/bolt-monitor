## Context

Project now has SST bootstrap, health endpoint, and shared monitor configuration contract, but no agreed DynamoDB storage pattern. Upcoming monitor CRUD, scheduler output, incidents, alerting, and audit features all need consistent read and write paths. Without a single-table design first, each feature is likely to create isolated item shapes and incompatible query patterns.

Single-table DynamoDB fits this system because monitoring data is query-driven: list monitors by tenant, fetch monitor status, retrieve recent runs, inspect open incidents, and review audit history. Those patterns benefit from shared key design more than from one-table-per-entity modeling.

## Goals / Non-Goals

**Goals:**
- Define canonical PK/SK patterns for core monitoring entities.
- Specify item families for monitor config, status, runs, incidents, and audit records.
- Identify initial GSIs needed for dashboard and operational query patterns.
- Define retention expectations for high-volume run history items.

**Non-Goals:**
- Implement DynamoDB table or GSIs in infrastructure code.
- Implement CRUD handlers or repository code.
- Finalize every future entity such as billing or auth models.
- Optimize for analytics workloads beyond immediate application reads.

## Decisions

### Use one primary application table
- Decision: core application data should live in one DynamoDB table with typed items.
- Rationale: supports cross-entity access patterns, minimizes coordination across future features, and matches expected monitor/status/run/incident query flow.
- Alternative considered: multiple tables split by entity type.
- Why not: pushes joins and consistency burden into application code and makes later query evolution harder.

### Design around access patterns, not entity purity
- Decision: define item families and duplicate selected fields where needed for efficient reads.
- Rationale: DynamoDB rewards query-oriented modeling; monitor listings, status dashboards, and open incident views need direct keys and sorted ranges.
- Alternative considered: strictly normalized records with no duplication.
- Why not: increases read complexity and weakens latency guarantees.

### Keep tenant-aware partitioning from day 1
- Decision: tenant/workspace ownership should appear in keys and GSIs from first table design.
- Rationale: future-safe isolation, cleaner audit ownership, and simpler evolution from single-workspace product to true multi-tenant product.
- Alternative considered: add tenant later after CRUD exists.
- Why not: would force repartitioning and backfill work later.

### Treat run history as high-volume, expiring data
- Decision: `CheckRun` items should include TTL/retention strategy from first design pass.
- Rationale: run history will dominate write volume and storage cost quickly.
- Alternative considered: indefinite retention of every raw run in main table.
- Why not: cost and query noise grow fast without adding equivalent product value.

### Start GSIs with operational reads only
- Decision: initial GSIs should cover open incidents by tenant and monitor status dashboard reads, not every imaginable future query.
- Rationale: keeps first design focused and avoids speculative indexes.
- Alternative considered: prebuild many GSIs for analytics and reporting.
- Why not: higher complexity and waste before real usage appears.

## Risks / Trade-offs

- [Risk] Single-table designs are harder to reason about initially. -> Mitigation: document item families and access patterns explicitly in the spec.
- [Risk] Early item choices may need refinement as incidents and alerting ship. -> Mitigation: lock core key conventions now and leave secondary entities extensible.
- [Risk] Raw check history can overwhelm storage or partitions. -> Mitigation: require TTL and keep run access pattern scoped to recent history.

## Migration Plan

1. Define single-table item families and query expectations in spec.
2. Implement table and model-to-record mappings in follow-up changes.
3. Reuse table design in monitor CRUD, scheduler, incident, and audit work.

Rollback at this stage is low-risk because this change defines storage contract only. No persisted production data needs migration yet.

## Open Questions

- Should `MonitorRef` be a distinct tenant-listing item from day 1, or deferred until list endpoints land?
- Should recent runs across a tenant need a GSI in v1, or can that wait?
- What retention window should raw `CheckRun` items target before rollups exist?
