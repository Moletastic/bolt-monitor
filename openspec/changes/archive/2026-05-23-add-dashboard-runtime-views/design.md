## Context

Dashboard currently covers monitor CRUD, enable/disable, and run history read. Runtime backend is built but the dashboard cannot yet observe incident state, trigger on-demand checks, view audit history, or control recurring execution. UI must be ready before or at the same time runtime goes live so operators can see results and act on failures.

Existing dashboard routes:
- `/` — monitor list
- `/monitors/new` — create form
- `/monitors/{id}` — monitor detail with status, config, runs table

Missing runtime-observation surfaces on monitor detail: incidents tab, audit history tab, manual run trigger.
Missing pages: incidents list, incidents/{id}, admin/scheduler, locations.

## Goals / Non-Goals

**Goals:**
- Complete runtime-observation UI for monitor detail (incidents, audit, manual run).
- Add incidents list and detail views with operator action buttons.
- Add scheduler admin view for recurring execution control.
- Add probe locations view.
- Extend API client with missing endpoint functions.

**Non-Goals:**
- No backend API changes — all endpoints already exist.
- No real-time polling or WebSocket updates — rely on page refresh and no-store fetch.
- No authentication or multi-tenancy changes.
- No monitoring alerting configuration UI beyond what already exists in API.

## Decisions

### Extend monitor detail with tabbed sections instead of new route
- Decision: add incidents and audit as tabbed sections within the existing `/monitors/{id}` page rather than new routes.
- Rationale: keeps incident and audit context tied to the monitor being operated on, avoids route proliferation.
- Alternative considered: separate `/monitors/{id}/incidents` and `/monitors/{id}/audit` routes.
- Why not: operator workflow stays anchored to one monitor; tab switching is sufficient.

### Use server actions for incident ack/resolve
- Decision: implement acknowledge and resolve as Next.js server actions rather than client-side fetch calls.
- Rationale: aligns with existing pattern for enable/disable toggle using `toggleMonitorAction`, keeps formPOST pattern consistent.
- Alternative considered: client-side fetch with optimistic update.
- Why not: simpler to implement with existing pattern, optimistic UI not required for v1.

### Probe locations as a read-only reference page
- Decision: expose probe locations as a simple read-only page rather than integrating into monitor creation flow beyond the existing dropdown.
- Rationale: operators only need to know what regions exist when creating or editing a monitor; no action needed on locations.
- Alternative considered: inline location selector with live availability.
- Why not: catalog is static in v1; selector already exists in monitor form.

## Risks / Trade-offs

- [Risk] Empty states on incidents and audit may confuse operators if runtime has not yet executed. -> Mitigation: show helpful empty states explaining that runs produce incidents and mutations produce audit entries.
- [Risk] Scheduler admin view allows toggling recurring execution which could cause cost if misconfigured. -> Mitigation: clearly label the control, show current state, require confirmation on disable.
- [Risk] Manual run button may be clicked repeatedly, queueing many work items. -> Mitigation: disable button briefly after click or show accepted state.

## Migration Plan

1. Extend `lib/api.ts` with new endpoint functions.
2. Add incident types to `lib/types.ts`.
3. Add incidents tab and audit tab to monitor detail page.
4. Add manual run trigger button and server action to monitor detail.
5. Create incidents list page at `/incidents`.
6. Create incidents detail page at `/incidents/{id}`.
7. Create scheduler admin page at `/admin/scheduler`.
8. Create probe locations page at `/locations`.
9. Verify all pages render with empty or mock data without errors.

Rollback: remove new page files and revert API client additions. No data migration needed.
