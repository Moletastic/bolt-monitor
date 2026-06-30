## Context

The repository has a real monitor API but no frontend application. `DESIGN.md` defines the visual system and `code.html` shows a richer observability console concept, but current backend capabilities only cover monitor CRUD, latest status, and recent runs. Dashboard v1 should convert those real capabilities into a usable product surface without fabricating backend domains that do not yet exist.

## Goals / Non-Goals

**Goals:**
- Add the first frontend app under `apps/dashboard` using Next.js, TypeScript, and shadcn/ui.
- Show monitor list data using the existing `/api/v1/monitors` response as the primary dashboard read model.
- Support monitor detail, create, edit, and enable or disable flows.
- Translate `DESIGN.md` tokens and the strongest patterns from `code.html` into reusable UI primitives.

**Non-Goals:**
- Full observability command-center parity with every `code.html` panel.
- New backend domains for alerts, incidents, maintenance, logs, fleet aggregates, or auth.
- Solving long-term multi-tenant navigation or role-based access control.

## Decisions

### Treat monitor list as the v1 dashboard home
- Decision: the landing dashboard should be a monitor operations overview backed by `GET /api/v1/monitors`.
- Rationale: it already includes embedded status summary and maps to a real operator workflow.
- Alternative considered: recreate the fleet-summary hero from `code.html` as the main page.
- Why not: current backend cannot support the implied global aggregates honestly.

### Reuse the visual language, not the exact information architecture
- Decision: v1 should adopt the color, spacing, typography, density, and widget treatment from `DESIGN.md` and selectively adapt components from `code.html`.
- Rationale: preserves the desired brand without forcing placeholder data panels.
- Alternative considered: implement `code.html` literally and fill unsupported panels with mock data.
- Why not: that would create misleading product expectations and likely cause rework.

### Prefer same-origin or server-side API integration
- Decision: frontend should integrate through server-side fetches or local proxy routes rather than assuming browser-direct cross-origin access.
- Rationale: API CORS policy is not yet established in the bootstrap stack.
- Alternative considered: direct client fetches to API Gateway.
- Why not: can block local development and couples the UI to early infra details.

### Ship monitor detail as a focused operator page
- Decision: monitor detail should separate configuration, current status, and recent runs into distinct sections.
- Rationale: aligns with existing API boundaries and keeps the page legible in a dense monitoring UI.
- Alternative considered: embed full run history into the list or a single oversized detail payload.
- Why not: too heavy and inconsistent with existing read surfaces.

### Defer probe-location discovery unless explicitly added to scope
- Decision: v1 may start with the currently available catalog assumptions, but any dynamic location picker should become an explicit follow-on API decision.
- Rationale: only one built-in probe location currently exists and there is no location-list endpoint yet.
- Alternative considered: quietly expand backend scope during UI work.
- Why not: violates the goal of validating the current monitor API surface first.

## Risks / Trade-offs

- [Risk] The dashboard create/edit form may outgrow hardcoded probe-location options quickly. -> Mitigation: design the form so the location input can later swap to API-backed options.
- [Risk] `code.html` may bias scope toward unsupported fleet widgets. -> Mitigation: treat it as a visual reference, not a commitment to every panel.
- [Risk] First frontend app may force decisions about routing, shared tokens, and build tooling. -> Mitigation: keep v1 architecture simple and local to `apps/dashboard`.
- [Risk] Browser/API integration may hit environment or CORS friction. -> Mitigation: default to server-side reads and actions for bootstrap.

## Migration Plan

1. Create dashboard capability spec and implementation tasks.
2. Add the frontend app shell and design-token mapping.
3. Implement monitor overview and detail screens backed by existing APIs.
4. Add mutation flows for create, edit, enable, and disable.
5. Validate the app against local SST API before considering broader dashboard expansion.

Rollback is straightforward because this adds a new app without changing existing backend contracts.

## Open Questions

- Should dashboard v1 open on a card grid, dense table, or responsive hybrid view?
- Should create/edit happen in modal flows or dedicated pages?
- Do we want to capture probe-location API discovery as a separate follow-on proposal now, or wait until UI friction proves it necessary?
