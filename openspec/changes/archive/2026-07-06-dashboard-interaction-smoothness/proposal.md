## Why

The dashboard currently feels responsive but abrupt: user actions complete quickly, yet many interactions jump between states through redirects, duplicate feedback surfaces, or background refreshes that do not preserve visual continuity. This change improves perceived smoothness while preserving the existing dashboard router convention and avoiding a broad rewrite of navigation-first forms.

## What Changes

- Define product rules for smoother dashboard interactions without introducing new imperative router navigation.
- Standardize mutation feedback so a single user action produces one clear feedback surface, not competing toast and inline banner messages.
- Improve same-page mutation flows, such as monitor enable/disable and incident acknowledge/resolve, so pending and completed states are visible in-place where navigation is not required.
- Make polling-driven refreshes non-urgent and avoid redundant refresh bursts when the tab becomes visible.
- Fix invalid nested interactive controls in mobile dashboard cards so touch, keyboard, and screen-reader behavior remains predictable.
- Verify and fill gaps in route-specific loading placeholders only where the current loading UI does not match the destination shape.
- Verify destructive-delete focus restoration against the existing dashboard requirement and close gaps without changing required post-delete redirects.

## Capabilities

### New Capabilities
- `dashboard-interaction-smoothness`: Covers dashboard interaction continuity, feedback-surface ownership, same-page mutation feedback, polling refresh smoothness, and interactive-control accessibility.

### Modified Capabilities
- None. This change intentionally preserves `dashboard-router-convention`, `dashboard-ui-action-state-results`, `dashboard-loading-states`, and existing delete redirect requirements in `dashboard-web-app`.

## Impact

- Affected app area: `apps/dashboard` components, server actions, route pages, loading UI, and tests.
- No API contract changes are expected.
- No dependency changes are expected.
- Existing navigation-first flows may continue to use server-action redirects unless a task explicitly converts a same-page mutation to typed action state.
