## 1. Token Contract

- [x] 1.1 Inventory dashboard global styles, Tailwind theme extensions, and shared primitive styling against the color, typography, spacing, radius, layout, and elevation values in `DESIGN.md`.
- [x] 1.2 Define semantic CSS custom properties and Tailwind aliases for documented dashboard color and status roles.
- [x] 1.3 Configure documented typography roles, the 4px spacing baseline, responsive gutters, and wide-layout grid utilities.
- [x] 1.4 Reconcile Tailwind compact radius tokens with the exact `DESIGN.md` radius scale and preserve intentional full-round shapes.
- [x] 1.5 Replace all-card shadow defaults with documented tonal surface layering and reserve elevated shadow treatment for overlays.

## 2. Shared Primitives

- [x] 2.1 Update shared card and panel primitives to use documented surface, border, header, padding, and compact-radius tokens.
- [x] 2.2 Update shared button, input, and select primitives for documented control surfaces, focus visibility, disabled state, and non-color-only invalid feedback.
- [x] 2.3 Add shared semantic feedback and unavailable-state primitives with status treatment, non-color cues, and accessible announcements.
- [x] 2.4 Align shared loading and skeleton primitives with dashboard surface and skeleton tokens.

## 3. Incremental Adoption

- [x] 3.1 Migrate representative high-use service, monitor, integration, and policy surfaces to shared layout, typography, feedback, and primitive contracts without changing route behavior.
- [x] 3.2 Replace duplicated local feedback class recipes in migrated surfaces with the shared semantic feedback primitives.
- [x] 3.3 Document remaining route-specific raw styling or primitive gaps as follow-up OpenSpec work rather than expanding this foundation beyond its scope.

## 4. Verification

- [x] 4.1 Add focused tests or guard coverage for documented token mappings and shared primitive contracts.
- [x] 4.2 Add DOM-level tests for feedback announcement and invalid form-control semantics.
- [x] 4.3 Run `make format-dashboard`, `make lint-dashboard`, `make check-dashboard`, `make test-dashboard`, and `make build-dashboard`.
- [x] 4.4 Verify representative dashboard and public-auth surfaces at desktop and mobile widths, including keyboard focus and reduced-motion behavior.
