## Context

The service-first model gives operators services with nested monitors, lifecycle state, rollup health, and dashboard routes under `/services`. Archive and disable are reversible lifecycle controls, but operators still need permanent deletion for mistakes and obsolete configuration. The current codebase already has API/repository primitives for `DELETE /api/v1/services/{serviceId}` and `DELETE /api/v1/services/{serviceId}/monitors/{monitorId}`, while dashboard API helpers and UI controls are not yet the complete operator-facing workflow.

Deletion is destructive, so it must remain distinct from archive and disable. Archive keeps service and monitor configuration available for inspection; delete removes configuration from normal read/list surfaces. Operational history should not be used as a reason to keep deleted configuration visible, but audit records should preserve that a deletion happened.

## Goals / Non-Goals

**Goals:**
- Provide tenant-scoped permanent deletion for services and nested monitors.
- Keep lifecycle safety guards: active services cannot be deleted directly, and the last monitor cannot be deleted from an active service.
- Remove deleted service and monitor configuration from normal API and dashboard views.
- Preserve or write audit evidence for successful deletion.
- Add dashboard delete affordances with confirmation, error feedback, and predictable redirects/refreshes.

**Non-Goals:**
- Hard-delete all historical check runs, incidents, and audit events for a deleted service or monitor.
- Add a restore/undo flow for deleted resources.
- Replace archive/reactivate or enable/disable behavior.
- Add bulk deletion.
- Add user authentication or role-based authorization beyond the current tenant boundary.

## Decisions

1. Use `DELETE` on existing resource URLs.

`DELETE /api/v1/services/{serviceId}` and `DELETE /api/v1/services/{serviceId}/monitors/{monitorId}` match the resource model and avoid action-style routes for permanent deletion. Alternatives considered were `POST /delete` action endpoints and dashboard-only deletion. Action endpoints blur the difference from archive/reactivate, and dashboard-only deletion would leave API clients without the capability.

2. Keep service deletion restricted to non-active services.

Active services represent live monitoring coverage. Requiring operators to archive or disable monitors first makes destructive intent explicit and avoids accidentally dropping active coverage. The alternative was to cascade-delete active services, but that would make one click able to remove all active checks for a service.

3. Keep monitor deletion restricted when it would leave an active service without monitors.

The service lifecycle model derives active state from enabled monitor coverage. Preventing deletion of the last monitor from an active service preserves the invariant that active services have monitor coverage. The alternative was to auto-transition the service to draft, but lifecycle auto-derive and archive/reactivate already define explicit service-state behavior and should not be hidden inside monitor deletion.

4. Delete configuration and derived state, not operational history by default.

Service and monitor metadata, service-monitor references, current status records, and notification links should be removed from active reads. Check runs, incidents, and audit records should remain governed by existing retention and read behavior unless they are purely derived reference records that would otherwise keep deleted configuration visible. This gives operators a clean management surface while retaining enough history for debugging and compliance.

5. Dashboard deletes use confirmation plus server actions.

The dashboard should call API helpers from server actions, then revalidate service pages and redirect away from deleted resource detail pages. Confirmation should be explicit and resource-specific. Native browser confirmation is acceptable initially; a custom modal can follow if the design system needs richer copy.

## Risks / Trade-offs

- Accidental permanent deletion -> Require explicit confirmation and block active-service deletion.
- Orphaned records after partial delete -> Use repository-level transactions/batches and tests covering all key families touched by delete.
- Audit trail loss when deleting service partitions -> Write deletion audit records outside the deleted configuration set or ensure audit records are not included in delete-key collection.
- Stale dashboard pages after deletion -> Revalidate `/services`, service detail, and monitor detail paths before redirecting.
- Confusion between archive and delete -> UI copy must distinguish reversible archive from permanent delete and explain why active resources may be blocked.

## Migration Plan

1. Finish or verify API delete handlers and repository behavior for service and monitor deletion.
2. Add dashboard API helpers and server actions.
3. Add guarded delete controls to service detail and monitor detail views.
4. Add API, repository, and dashboard tests for success, not-found, and conflict cases.
5. Deploy without data migration; deletion affects future operator actions only.
6. Rollback by removing dashboard controls and route exposure; already-deleted configuration cannot be restored without backup/manual reconstruction.

## Open Questions

- Should historical monitor-specific audit and incident views remain accessible by direct URL after the parent monitor is deleted, or should they only be available through future global history surfaces?
- Should service deletion require typing the service name once the UI has a modal component, or is a confirm action sufficient for the first implementation?
