## Context

Current dashboard UX already treats `Services` as the operator-facing module, but backend APIs, shared Go contracts, and DynamoDB storage still model `Monitor` as the top-level managed resource. That mismatch means the UI is effectively projecting a service concept over a monitor-first system, and it breaks down as soon as one logical service needs multiple probes to represent partial failure modes.

The repository is still in development phase. Production migration, backward compatibility, and table-preserving dual-write flows are not required. That allows a hard pivot to a cleaner service-first model instead of layering compatibility shims over the current monitor-first hierarchy.

## Goals / Non-Goals

**Goals:**
- Make `Service` the top-level operator-facing resource and `Monitor` a child resource.
- Support one service with many monitors and one monitor belonging to exactly one service.
- Use stable slug identities for `serviceId` and `monitorId` that fit nested routes and operator workflows.
- Allow draft services to exist with zero monitors.
- Derive service rollup state from enabled child monitor states.
- Preserve monitor-scoped runtime evidence for status snapshots, runs, incidents, and audit history.
- Add one optional validated `technologyKey` per service for primary service icon rendering.
- Keep monitor icon presentation frontend-derived from monitor protocol or type.
- Redesign DynamoDB item families and access patterns around service-first reads.

**Non-Goals:**
- Add push or cron heartbeat monitor types.
- Preserve flat top-level monitor APIs as long-term canonical routes.
- Maintain backward-compatible DynamoDB item families or migration shims.
- Support multiple icons per service.
- Persist icon metadata for monitors.

## Decisions

### Make `Service` the top-level managed resource
- Decision: replace monitor-first domain modeling with service-first modeling.
- Rationale: operators reason about services, while monitors are evidence about service health.
- Alternative considered: keep monitor as top-level resource and add service grouping only.
- Why not: that leaves the existing semantic mismatch in place and makes service health a UI-only fiction.

### Nest monitors under services and scope monitor identity within service
- Decision: monitor routes, storage keys, and resource identity all include parent service ownership.
- Rationale: one monitor belongs to one service, and monitor uniqueness only needs to hold within that service.
- Alternative considered: keep globally unique monitor IDs.
- Why not: unnecessary for the intended hierarchy and worse for human-readable URLs and fixtures.

### Use stable slugs instead of opaque generated identifiers
- Decision: `serviceId` and `monitorId` are stable slug identifiers, distinct from editable display names.
- Rationale: IDs will appear in nested paths, logs, fixtures, manual API use, and dashboard navigation.
- Alternative considered: ULIDs or server-owned opaque identifiers.
- Why not: tenant-scoped and service-scoped uniqueness already solve identity needs without sacrificing readability.

### Let the client suggest slugs and let the server validate them
- Decision: UI may derive initial slug suggestions from names, but server validates, enforces uniqueness, and persists the final identifier.
- Rationale: this gives good UX without introducing silent server-side identity rewriting.
- Alternative considered: server silently generating or mutating slugs on write.
- Why not: silent mutation makes identity less predictable and complicates links and user expectations.

### Allow draft services with zero monitors
- Decision: a service can exist in `draft` lifecycle state before any monitor is created.
- Rationale: service metadata and ownership can be created incrementally rather than forcing a large one-shot create flow.
- Alternative considered: require at least one monitor at service creation time.
- Why not: that couples unrelated setup concerns and adds unnecessary form friction.

### Derive service rollup from enabled child monitors only
- Decision: service rollup status is computed from service lifecycle plus enabled child monitor states; disabled monitors do not affect current service health.
- Rationale: stale disabled monitors should not poison triage or operational summaries.
- Alternative considered: include all monitors in rollup whether enabled or not.
- Why not: disabled monitors would distort service state and produce misleading summaries.

### Keep runtime artifacts monitor-scoped
- Decision: mutable latest status, run history, incidents, and audit reads remain attached to monitor-scoped partitions and APIs.
- Rationale: runtime evidence is fundamentally per monitor, while service state is a derived summary.
- Alternative considered: move runs and incidents under service partitions.
- Why not: service partitions become noisy and monitor detail queries become less direct.

### Use tenant-aware composite keys for service and monitor partitions
- Decision: service and monitor partition keys include tenant identity even though the current product still uses one built-in tenant.
- Rationale: service IDs are tenant-scoped and monitor IDs are service-scoped; tenant-aware composite keys avoid future repartitioning.
- Alternative considered: omit tenant from non-tenant partitions until multi-tenant behavior ships.
- Why not: future service ID reuse across tenants would require a key redesign later.

### Store one optional semantic `technologyKey` on services
- Decision: service metadata may include one optional validated `technologyKey` such as `postgres`, `mariadb`, `nginx`, `golang`, or `python`.
- Rationale: service cards benefit from one recognizable technology anchor without introducing icon taxonomy sprawl.
- Alternative considered: multiple icons or raw icon-library identifiers.
- Why not: multiple icons add complexity early, and raw icon strings couple domain data to frontend rendering details.

### Validate technology keys against a closed supported catalog
- Decision: backend persists only supported semantic technology keys and frontend maps those keys to concrete icon components.
- Rationale: prevents corrupted or user-injected class or icon selectors while decoupling persisted data from icon library implementation details.
- Alternative considered: free-form icon strings or CSS class fragments.
- Why not: those approaches invite data inconsistency, broken rendering, and accidental injection surfaces.

### Derive monitor icons from monitor protocol or type in frontend
- Decision: monitor resources do not persist icon metadata.
- Rationale: monitor icon semantics follow monitor type and protocol directly.
- Alternative considered: persist one icon key per monitor.
- Why not: that duplicates data the frontend can derive safely from canonical monitor type.

### Use nested service-monitor APIs as canonical surface
- Decision: create, read, update, lifecycle, status, run, incident, audit, and manual-run monitor routes all use nested service paths as canonical resource identity.
- Rationale: path structure matches ownership and removes ambiguity around flat monitor lookup.
- Alternative considered: keep flat monitor paths as canonical and treat service context as optional.
- Why not: that preserves the old top-level monitor worldview the change is meant to replace.

### Prefer hard reset over compatibility layers
- Decision: old monitor-first DynamoDB items and route assumptions can be removed or reset during implementation.
- Rationale: repository is still in development phase and no production migration constraints were provided.
- Alternative considered: dual-read, dual-write, or old/new translation layers.
- Why not: those add complexity without protecting live users.

## Risks / Trade-offs

- [Risk] Hard pivot touches many modules and specs at once. -> Mitigation: use one change that updates all affected capability contracts together.
- [Risk] Immutable slug identities make renames a display-name concern rather than an identity concern. -> Mitigation: keep `name` editable and keep IDs stable.
- [Risk] Service rollups can hide detail if over-relied on. -> Mitigation: keep monitor detail, runs, incidents, and audit history as first-class drill-ins.
- [Risk] Supported technology catalog may need expansion. -> Mitigation: validate semantic keys against a small allowlist that can grow without data-model changes.
- [Risk] Nested routes increase path length and handler scope. -> Mitigation: gain cleaner ownership semantics and avoid global monitor identity ambiguity.

## Migration Plan

1. Add new service-first proposal, design, and capability deltas.
2. Replace shared domain contracts with service and nested monitor identities.
3. Replace DynamoDB key helpers and repository mappings with service-first item families.
4. Replace flat monitor API routes with canonical nested service and monitor routes.
5. Update status, run, incident, audit, and manual-run paths to use tenant-aware composite monitor identities.
6. Update dashboard to render real service summaries, nested monitor flows, and service `technologyKey` icons.
7. Reset or delete obsolete development DynamoDB items and fixtures.

Rollback remains development-only and can rely on deleting development data rather than preserving old storage contracts.

## Open Questions

- Should deletion of a non-active service hard-delete all child monitor runtime history during development, or should archival be preferred for operator UX even before production?
- Should creating the first enabled monitor under a draft service remain a separate activation step, or is later auto-activation worth considering as a follow-on UX refinement?
- Should service-level incident summaries be added later as a derived read model, or should incident views remain strictly monitor-scoped for now?
