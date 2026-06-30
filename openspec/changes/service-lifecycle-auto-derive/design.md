## Context

The current service lifecycle model requires clients to explicitly manage lifecycle state transitions. When creating a service, clients must decide whether to start in `draft` or `active`. When enabling the first monitor on a draft service, the client must remember to update the service lifecycle to `active`. This is error-prone and creates confusing UX: the API allows setting `active` but returns 409 if no monitors are enabled.

The current activation guard at `handler.go:208` checks `current.EnabledCount == 0` but the error message references `MonitorCount`, indicating the guard was miswired at some point (now fixed to check `EnabledCount`).

**Current flow:**
1. Client creates service with optional `lifecycleState: "draft"`
2. Client creates monitors under the service
3. Client enables monitors one by one
4. Client must manually update service lifecycle to `active` when ready
5. 409 Conflict if client sets `active` before enabling any monitors

**Desired flow:**
1. Client creates service (lifecycle defaults to `draft` automatically)
2. Client creates and enables monitors
3. Service lifecycle auto-derives to `active` when first monitor is enabled
4. No explicit lifecycle management needed for draft→active transition
5. Explicit `/archive` action to move to archived state

## Goals / Non-Goals

**Goals:**
- Eliminate client-side lifecycle management for draft↔active transitions
- Derive lifecycle from monitor state (`enabledCount`) automatically
- Keep archived as explicitly mutable (client must request it)
- Simplify rollup derivation logic (remove lifecycle-based shortcuts)
- Add `/archive` and `/reactivate` endpoints for explicit state transitions

**Non-Goals:**
- Multi-tenant support (currently single-tenant with `DEFAULT` tenant)
- Automatic archival of stale services (could be future enhancement)
- Migration of existing services' lifecycle state (deferred to operational runbook)

## Decisions

### Decision 1: Lifecycle derivation triggers

**Option A: Derive on every monitor enable/disable (transactional)**
- When `SetMonitorEnabled` is called, re-derive parent service lifecycle in same transaction
- Guarantees consistency but adds latency to enable/disable calls
- Requires repository access to update service record

**Option B: Derive on read (lazy)**
- Compute derived lifecycle when service is read
- Simpler implementation, no extra writes
- Risk: brief inconsistency if service is read between monitor enable and rollup write

**Option C: Derive on write + background reconciliation**
- Immediately update service lifecycle on monitor enable/disable
- Background job corrects any drift
- Most robust but most complex

**Chosen: Option A (derive on write, transactional)**
The existing `SetMonitorEnabled` already performs a read-modify-write on the service status record. We extend this transaction to also recompute and write the derived lifecycle state. This keeps lifecycle derivation atomic with monitor state changes.

### Decision 2: Remove lifecycle from CreateServiceRequest

**Option A: Remove entirely, always start as draft**
- Client cannot request `active` on create (must enable monitors first)
- Simplest model, no ambiguity

**Option B: Accept but ignore lifecycleState**
- Accept the field for API compatibility but derive internally
- Breaks expectation that `lifecycleState: "active"` does something

**Chosen: Option A**
Breaking API change is acceptable since this is a new change. Clients must enable monitors before service becomes active.

### Decision 3: UpdateService — lifecycle field behavior

**Option A: Remove lifecycle field entirely**
- Cannot update lifecycle via PATCH
- Forces clients to use `/archive` action

**Option B: Accept lifecycle field but ignore it, derive instead**
- Maintains API surface but silently ignores writes
- Confusing UX

**Chosen: Option A**
Lifecycle is derived, not stored. Clients must use explicit actions.

### Decision 4: Archived state transitions

**Option A: Archived is a terminal state**
- Once archived, service cannot be reactivated
- Requires creating a new service

**Option B: Archived can be reactivated via `/reactivate`**
- Reactivates with current monitor configuration
- New lifecycle derived from enabledCount at time of reactivation

**Chosen: Option B**
Allows accidental archival to be undone. New lifecycle recomputed from current monitor state.

### Decision 5: Simplify rollup derivation

**Current logic** (`deriveServiceRollup`):
```go
switch lifecycle {
case draft: return "draft"
case archived: return "archived"
}
enabled := filter enabled monitors...
if len(enabled) == 0 { return "paused" }
// ... status aggregation
```

**Simplified logic:**
```go
enabled := filter enabled monitors...
if len(enabled) == 0 { return "paused" }
// ... status aggregation (same for all states)
```

**Rationale:** Since lifecycle is now derived from monitor state, the draft/archived shortcuts in rollup become redundant. A service with `enabledCount == 0` will always have rollup `paused` regardless of whether it's `draft` or `archived`. An archived service with enabled monitors will have a meaningful status derived from those monitors.

## Risks / Trade-offs

[Risk] Existing services with explicit lifecycleState may behave differently after deploy
→ **Mitigation**: The derivation logic matches current behavior for services with correct lifecycle. Services with mismatched lifecycle (e.g., `active` with zero enabled monitors) will auto-correct on next enable/disable event. Consider a one-time migration job.

[Risk] Clients currently relying on lifecycleState in responses may break
→ **Mitigation**: Lifecycle still returned in API responses as read-only computed field. Only the mutability is removed.

[Risk] Race condition between monitor enable and lifecycle derivation
→ **Mitigation**: Transactional update ensures atomicity. If two monitor enables happen concurrently, DynamoDB transactions ensure consistency.

## Migration Plan

1. **Deploy** new code with lifecycle derivation logic (behind feature flag off)
2. **Run migration**: Re-derive lifecycle for all existing services based on their `enabledCount`
3. **Enable feature flag**: Start deriving lifecycle from monitor state
4. **Monitor**: Watch for 409 conflicts or unexpected lifecycle values
5. **Cleanup**: Remove old lifecycle field handling code (future PR)

## Open Questions

1. Should we emit audit events for lifecycle state transitions (auto-derive)?
2. Do we need to notify clients when lifecycle auto-derives (webhook/polling hint)?
3. Should archived services with no enabled monitors derive to `draft` on reactivate, or maintain `archived` until explicit action?
