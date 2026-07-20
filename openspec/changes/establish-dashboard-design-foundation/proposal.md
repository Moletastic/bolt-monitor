## Why

`DESIGN.md` defines the dashboard's monitoring-console visual language, but its token contract is only partially represented in dashboard code. Raw utilities and inconsistent primitive styling make visual consistency depend on individual implementation choices and make future UI work expensive to review and evolve.

## What Changes

- Establish a dashboard design-system foundation derived from `DESIGN.md`, including semantic color, typography, spacing, radius, layout, and elevation tokens.
- Reconcile the configured Tailwind radius scale with the documented design tokens before further surface work expands the mismatch.
- Standardize shared dashboard primitives for cards, form controls, feedback states, and loading surfaces around semantic design tokens.
- Provide clear responsive layout and typography rules for dashboard surfaces without redesigning existing operator workflows.
- Add focused verification that prevents foundational tokens and shared primitive contracts from drifting from the documented design system.

## Capabilities

### New Capabilities
- `dashboard-design-system`: Defines the token contract and shared visual primitives for the dashboard monitoring-console interface.

### Modified Capabilities

None.

## Impact

- Affects `DESIGN.md`, dashboard Tailwind configuration and global styles, and shared UI primitives under `apps/dashboard/components/ui`.
- Affects dashboard loading, form, feedback, and card surfaces as they migrate to the shared foundation.
- Adds no backend, API, authentication, routing, or data-model changes.
