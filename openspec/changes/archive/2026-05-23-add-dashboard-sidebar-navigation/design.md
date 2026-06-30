## Context

`apps/dashboard` already centralizes shared framing in `AppShell`, so sidebar work should stay concentrated there instead of duplicating layout logic per page. Existing root page is a full monitor overview, which makes `Dashboard` effectively mean monitor operations instead of a neutral product home. This change needs to separate those concerns: root becomes an empty `Dashboard` landing page, while existing monitor overview moves into `Services`.

## Goals / Non-Goals

**Goals:**
- Replace bootstrap navigation labels with module-oriented sidebar items: `Dashboard`, `Services`, `Integrations`, `Audit Trail`, and `Config`.
- Move current monitor overview from `/` to the `Services` module.
- Make `/` render a simple `Dashboard` landing page with WIP messaging.
- Keep monitor create and detail flows grouped with `Services` rather than `Dashboard`.
- Add lightweight landing pages for non-dashboard modules so every sidebar destination resolves successfully.
- Preserve existing visual language and responsive shell behavior.

**Non-Goals:**
- No backend API work for services, integrations, audit trail, or config.
- No attempt to fully implement each module's real product workflows in this change.
- No redesign of monitor forms or monitor detail content beyond navigation/frame adjustments.

## Decisions

### Use one shared sidebar model in `AppShell`
- Decision: define sidebar modules in one navigation config inside `AppShell` or a nearby dashboard-only helper.
- Rationale: active-state rules and route ownership stay centralized, so new pages do not drift from shell behavior.

### Treat monitor surfaces as `Services` module content
- Decision: move current monitor overview to `/services` and map monitor overview, create, and detail flows to the `Services` module.
- Rationale: user explicitly wants root page to become `Dashboard` WIP landing page, so monitor operations need a different owning module.

### Keep root dashboard page intentionally empty for v1
- Decision: replace current `/` monitor overview with a minimal `Dashboard` page that communicates work-in-progress state.
- Rationale: this preserves requested information architecture without inventing unsupported summary content for the `Dashboard` module.

### Ship module landing pages as honest placeholders
- Decision: create simple landing routes for `/`, `/integrations`, `/audit-trail`, and `/config`, while `/services` becomes the real home for current monitor overview content.
- Rationale: sidebar should never link to missing pages, but this change should not expand into full module implementation.

## Risks / Trade-offs

- [Risk] Route-to-module matching can become brittle as dashboard grows. -> Mitigation: encode active matching rules alongside sidebar definitions instead of scattering path checks.
- [Risk] New root dashboard page may feel empty. -> Mitigation: present it as explicit WIP module landing page, not broken or unfinished navigation.
- [Risk] Existing monitor pages may regress on mobile layout if sidebar sizing changes. -> Mitigation: keep current responsive `AppShell` structure and only adjust navigation content/model.

## Migration Plan

1. Update dashboard sidebar configuration and active-state logic.
2. Move current monitor overview to `Services` route and add new module landing routes.
3. Verify existing dashboard and monitor pages still render correctly inside the shared shell with updated module ownership.

Rollback: revert sidebar config and remove new module route files.

## Open Questions

- Should monitor create route move under `/services/...` in a follow-on change, or remain at current path with `Services` highlighted?
- Should `Dashboard` WIP landing stay intentionally empty until product metrics exist, or gain curated summary cards later?
