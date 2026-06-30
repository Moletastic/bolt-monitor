## Context

Current monitoring model uses AWS-flavored `regions` language, but product intent is vendor-neutral probing across HTTP, TCP, and gRPC targets. Letting users invent arbitrary location strings would make scheduling, metrics, dashboard grouping, and incident semantics inconsistent across tenants. System needs a canonical probe-location catalog first, then monitors should select from that catalog.

This change touches monitor modeling, future CRUD validation, scheduler routing, and storage semantics. It also affects naming: probe execution locations should be described as probe locations, not cloud-provider regions.

## Goals / Non-Goals

**Goals:**
- Define a system-owned probe-location catalog.
- Define stable probe-location identifiers suitable for scheduler routing and metrics.
- Update monitor configuration semantics so monitors reference catalog locations instead of free-form values.
- Preserve tenant flexibility by letting tenants or users choose subsets of system-defined locations.

**Non-Goals:**
- Implement probe workers or actual multi-location execution.
- Implement tenant-level location enablement logic yet.
- Design billing or per-plan limits for locations.
- Expose custom user-defined location creation.

## Decisions

### System owns available probe locations
- Decision: available probe locations come from platform-managed catalog, not free-form user input.
- Rationale: scheduler, metrics, and incidents need canonical execution location IDs.
- Alternative considered: let each tenant or user invent arbitrary region/location strings.
- Why not: leads to typos, inconsistent analytics, and undefined worker routing.

### Monitors select subset of allowed probe locations
- Decision: monitor configuration should carry selected probe-location IDs, chosen from the catalog.
- Rationale: keeps monitor model flexible while preserving system control over valid options.
- Alternative considered: one global execution location with no monitor-level choice.
- Why not: too restrictive for future distributed probing value.

### Use vendor-neutral naming
- Decision: use `probeLocation` / `probeLocations` terminology instead of `region` / `regions`.
- Rationale: application checks generic network services, not AWS resources.
- Alternative considered: keep `regions` as shorthand.
- Why not: misleading and cloud-specific.

### Keep catalog records lightweight and operational
- Decision: probe-location records should focus on stable ID, display name, enabled state, and execution metadata needed for routing.
- Rationale: enough for validation and scheduling without over-design.
- Alternative considered: model detailed infrastructure topology immediately.
- Why not: not needed before actual workers exist.

## Risks / Trade-offs

- [Risk] Renaming `regions` to `probeLocations` creates churn in early model work. -> Mitigation: do it now before public APIs or large datasets exist.
- [Risk] Catalog may need future tenant-level override rules. -> Mitigation: keep global catalog and tenant selection concerns separate.
- [Risk] Too little metadata on locations may limit scheduler work later. -> Mitigation: keep identifiers stable and allow catalog expansion without changing monitor semantics.

## Migration Plan

1. Define probe-location catalog capability and monitor-selection rules.
2. Update monitor configuration contract to use `probeLocations` terminology and catalog validation.
3. Implement shared model and CRUD validation in follow-up changes.

Rollback is low-risk because no external clients or persisted production records depend on current region naming yet.

## Open Questions

- Should probe-location IDs be opaque (`loc_123`) or human-readable (`iad`, `dub`, `sin`)?
- Should a monitor require at least one probe location in v1, or can it inherit one default system location?
- When tenant-specific enablement lands, should it be a separate model or embedded in catalog exposure logic?
