## Context

The dashboard is a Next 15 App Router application using Tailwind and small shared UI primitives. `DESIGN.md` defines the intended monitoring-console visual system: semantic dark surfaces, Inter and JetBrains Mono roles, a 4px spacing baseline, a compact radius scale, responsive grid behavior, and tonal elevation. The current Tailwind configuration exposes some color, font, radius, and shadow values, but it does not encode the full contract. Its radius scale also conflicts with the documented values. Shared primitives and route components therefore carry duplicated raw utilities and inconsistent visual decisions.

This change establishes a stable frontend foundation without changing routes, data flow, authentication, or operator workflows. It must remain compatible with the active `improve-dashboard-interaction-responsiveness` change, which owns interaction behavior, loading continuity, and focused client-island work.

## Goals / Non-Goals

**Goals:**

- Make `DESIGN.md` the source of truth for dashboard visual tokens.
- Expose semantic design tokens through the dashboard styling system for colors, typography, spacing, radius, layout, and elevation.
- Give shared primitives a consistent monitoring-console baseline for cards, inputs, selects, buttons, feedback, and loading surfaces.
- Define responsive and accessibility-safe presentation rules that product surfaces can reuse without route-specific styling drift.
- Add focused checks for the token contract and primitive semantics.

**Non-Goals:**

- No redesign of dashboard information architecture, routes, or workflows.
- No replacement of Tailwind, Next.js, Radix, or existing icon libraries.
- No new global client state, API, backend, or authentication changes.
- No migration of every existing route in one change; route-specific visual refinement remains incremental.
- No change to action pending, polling, search, tab, menu, or data-freshness behavior owned by the interaction-responsiveness change.

## Decisions

### Keep `DESIGN.md` as the token authority and encode semantic aliases

Decision: Translate the documented tokens into CSS custom properties and Tailwind theme aliases. Product code consumes semantic names such as surface, foreground, primary, status, spacing, typography, and elevation instead of raw color values or one-off measurements. The existing design document remains human-readable source material; generated tooling or a second standalone token file is not introduced.

Rationale: one design authority reduces drift while CSS variables preserve runtime theming compatibility and Tailwind keeps component usage ergonomic.

Alternative considered: use raw Tailwind palette classes per component. Rejected because it has already created duplicated visual decisions and cannot express dashboard semantic intent consistently.

### Reconcile the configured compact scale with the documented scale before primitive migration

Decision: configure Tailwind radius tokens to exactly match the `DESIGN.md` contract: `sm` 2px, default 4px, `md` 6px, `lg` 8px, and `xl` 12px. Encode the documented 4px baseline with named spacing aliases and expose typography roles for display, headline, title, body, data, and label text. Preserve `rounded-full` only for intentionally semantic pills, dots, circular controls, and progress forms.

Rationale: shared token correction is safer and lower churn than compensating through per-component utility overrides.

Alternative considered: retain the archived reduced-radius scale. Rejected because current `DESIGN.md` is the agreed design source and its named values must match configured values.

### Make primitives semantic, composable, and narrow

Decision: update shared UI primitives to own stable baseline appearance and variants while callers retain layout and content composition. Card primitives provide documented surface, border, header, and elevation treatment. Form controls expose default, focus, disabled, and invalid presentation. Feedback primitives express informational, success, warning, error, and unavailable states with icon, text, and appropriate live-region semantics. Loading primitives use surface and skeleton tokens.

Rationale: semantic primitives eliminate repeated class recipes and centralize accessible defaults without inventing a large component framework.

Alternative considered: create domain-specific components for every dashboard card and form. Rejected because it would couple reusable visual foundations to route-specific business content.

### Use tonal layering, not default heavy shadows

Decision: default cards and panels use documented surface layers and low-contrast borders. Elevated overlays such as dialogs and menus use the overlay border and restrained `panel` shadow. Individual routes do not add arbitrary shadows to establish hierarchy.

Rationale: preserves dense, calm operational scanning and matches the design language.

Alternative considered: retain the current shadow on all cards. Rejected because it visually flattens elevation meaning and conflicts with the documented system.

### Provide responsive rules as reusable layout utilities, not page templates

Decision: encode desktop, tablet, and mobile container/gutter, grid, and display typography rules as reusable styling contracts. Dashboard routes adopt these utilities as touched; no universal page wrapper migration is required in this change.

Rationale: creates consistent responsive constraints without risking route-wide layout regressions.

Alternative considered: replace all page layouts with a single fixed dashboard template. Rejected because route content density differs and existing layouts need incremental verification.

### Verify contract and behavior at right level

Decision: add unit or guard coverage for token mappings and shared primitive class/semantic contracts, plus DOM-level coverage for feedback announcement and form-control invalid states. Visual review of representative desktop and mobile routes validates token impact; production build, lint, typecheck, and dashboard tests remain required.

Rationale: token drift is deterministic, while accessibility behavior needs rendered-DOM evidence.

Alternative considered: source-string tests only. Rejected because they do not verify rendered roles, labels, or responsive presentation.

## Risks / Trade-offs

- [Global token correction changes existing surfaces] -> Migrate primitives first and visually verify representative service, monitor, integration, policy, and authentication routes at desktop and mobile widths.
- [Semantic aliases can grow into another ungoverned palette] -> Add only names documented in `DESIGN.md`; new semantic tokens require a spec-backed design decision.
- [Primitive variants can become over-general] -> Keep variants limited to shared visual states and leave business-specific layout/content in callers.
- [Concurrent interaction work may edit same components] -> Keep interaction behavior changes out of this change and rebase implementation sequencing around shared primitive ownership.

## Migration Plan

1. Inventory current global CSS, Tailwind tokens, and shared primitive consumers against `DESIGN.md`.
2. Encode token aliases and reconcile radius, typography, spacing, responsive, and elevation contracts.
3. Update shared card, form-control, feedback, and loading primitives to consume those tokens.
4. Migrate representative high-use surfaces, then remaining callers incrementally without changing route behavior.
5. Add token, primitive, and DOM accessibility coverage; run dashboard checks and responsive visual review.

Rollback is code-only: revert token and primitive commits together. No persisted data, API contract, or deployment migration is involved.

## Open Questions

None. Future design additions require updates to `DESIGN.md` and this capability before implementation.
