## Why

The current service lifecycle model is client-managed: clients explicitly set `lifecycleState` when creating or updating a service. This creates a problematic coupling where clients must track monitor state (enabled/disabled counts) to correctly transition services between draft, active, and archived states. The existing activation guard attempts to enforce this relationship but creates a confusing API where clients can request `active` but receive a conflict if no monitors are enabled.

The correct model is to **derive service lifecycle from monitor state**: a service with zero enabled monitors is semantically a draft, a service with at least one enabled monitor is active, and archived is the only explicitly-requested state transition.

## What Changes

1. **Server-derived lifecycle states** (draft/active computed from monitor state, not client-provided)
   - `draft`: auto when `enabledCount == 0`
   - `active`: auto when `enabledCount > 0`
   - `archived`: only mutable state (requires explicit `/archive` action)
2. **Remove `lifecycleState` from CreateServiceRequest** — server assigns initial `draft` state
3. **Remove `lifecycleState` from UpdateServiceRequest** — lifecycle is no longer client-mutable (except via `/archive`)
4. **New `POST /api/v1/services/{serviceId}/archive` endpoint** — explicit archive action
5. **New `POST /api/v1/services/{serviceId}/reactivate` endpoint** — move from archived back to draft (if enabledCount==0) or active (if enabledCount>0)
6. **Remove activation guard** — since lifecycle auto-derives, no longer needed
7. **Remove lifecycle from Service response's modifiable fields** — lifecycle becomes read-only computed field
8. **Simplify rollup derivation** — since lifecycle is now derived from monitor state, draft/archived rollup shortcuts become redundant; use monitor-based derivation for all states

## Capabilities

### New Capabilities

- **service-lifecycle-auto-derive**: Lifecycle state is derived server-side from `enabledCount`. Draft when zero monitors are enabled, active when at least one is enabled. Archived is the only explicitly-requested transition. This eliminates the need for clients to manually manage lifecycle and removes the activation guard.
- **service-archive-action**: Explicit `/archive` endpoint to transition a service to archived state. Archived is the only lifecycle state that is mutable (can transition back to draft/active via `/reactivate`).
- **service-reactivate-action**: Allow archived services to be reactivated. The new lifecycle is derived from current enabledCount.

### Modified Capabilities

- **service-management-api**: Remove `lifecycleState` from create/update requests. Lifecycle becomes a computed field derived from monitor state, not a settable field. Add `/archive` and `/reactivate` endpoints.
- **monitor-crud-api**: When monitor is enabled/disabled, trigger re-evaluation of parent service's derived lifecycle state (no explicit action needed from client).

## Impact

- **API**: CreateService and UpdateService no longer accept `lifecycleState`. New endpoints for archive/reactivate.
- **Client code**: Dashboard must remove lifecycle state selectors; lifecycle becomes a read-only badge derived from monitor configuration.
- **Rollup logic**: `deriveServiceRollup` can be simplified since lifecycle no longer short-circuits to draft/archived statuses — all rollups derive from monitor states.
- **Migration**: Existing services with explicit lifecycleState values will have their lifecycle re-derived on next monitor enable/disable event. Consider a migration script for existing services.
