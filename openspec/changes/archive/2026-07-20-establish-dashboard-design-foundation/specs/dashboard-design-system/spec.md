## ADDED Requirements

### Requirement: Dashboard styling derives from the documented token contract
The dashboard SHALL treat `DESIGN.md` as the source of truth for its monitoring-console design tokens. Dashboard styling SHALL expose semantic tokens for documented color roles, typography roles, compact radius values, 4px-based spacing, responsive gutters, and surface elevation through the shared CSS and Tailwind styling contract. Product components SHALL consume those semantic tokens rather than introducing raw visual values where an equivalent documented token exists.

#### Scenario: Dashboard styling configuration is loaded
- **WHEN** dashboard global styles and Tailwind configuration are evaluated
- **THEN** documented semantic color, typography, spacing, radius, layout, and elevation roles are available to dashboard components
- **AND** each configured compact radius token matches its documented value in `DESIGN.md`

#### Scenario: Dashboard author needs a documented visual value
- **WHEN** a dashboard component needs a documented color, type role, spacing value, radius, or surface elevation
- **THEN** the component uses the shared semantic styling contract
- **AND** the component does not introduce an equivalent route-local raw visual value

### Requirement: Dashboard primitives provide consistent monitoring-console surfaces
Shared dashboard UI primitives SHALL implement the documented monitoring-console baseline for cards, buttons, inputs, selects, feedback, and loading surfaces. Cards and panels SHALL use tonal surface layering and low-contrast borders by default. Only overlay primitives SHALL use the documented elevated shadow treatment by default.

#### Scenario: Dashboard renders a card or panel
- **WHEN** a dashboard route renders a shared card or panel primitive
- **THEN** the surface uses documented card background, border, header, padding, and compact-radius tokens
- **AND** the primitive does not apply elevated-overlay shadow styling by default

#### Scenario: Dashboard renders an overlay
- **WHEN** the dashboard renders a menu, dialog, or other overlay primitive
- **THEN** the overlay uses the documented overlay surface, border, compact radius, and restrained elevation treatment

#### Scenario: Dashboard renders form controls
- **WHEN** the dashboard renders shared text input or select controls
- **THEN** controls use the documented dark control surface and compact radius tokens
- **AND** keyboard focus is visibly indicated with the primary token
- **AND** invalid controls expose an error presentation without relying on color alone

### Requirement: Dashboard provides semantic feedback and loading presentation
The dashboard SHALL provide shared primitives for informational, success, warning, error, unavailable, and loading presentation. Feedback states SHALL use the documented semantic status treatment and communicate their meaning through text and an icon or equivalent non-color cue. Dynamic feedback and unavailable states SHALL use an appropriate accessible live-region semantic.

#### Scenario: Dashboard operation returns feedback
- **WHEN** a dashboard operation renders success, warning, or error feedback
- **THEN** the rendered feedback uses the shared semantic feedback primitive
- **AND** its meaning remains understandable without color perception

#### Scenario: Dashboard data is unavailable
- **WHEN** a dashboard surface cannot load its backing data
- **THEN** the surface renders the shared unavailable presentation
- **AND** the message is announced to assistive technology without removing unaffected dashboard content

#### Scenario: Dashboard shows a loading surface
- **WHEN** dashboard data is loading
- **THEN** loading placeholders use shared surface and skeleton tokens
- **AND** their presentation remains visually consistent with the surrounding dashboard surface

### Requirement: Dashboard foundation supports responsive operational layouts
The dashboard SHALL provide reusable responsive layout and typography rules based on `DESIGN.md`. Wide layouts SHALL support the documented 12-column grid and desktop margins, tablet layouts SHALL use compact navigation and gutters, and narrow layouts SHALL reflow dashboard content into one column with mobile gutters and compact display typography.

#### Scenario: Dashboard renders on a wide viewport
- **WHEN** a dashboard route uses the shared layout contract on a viewport wider than 1024px
- **THEN** the route can use the documented 12-column grid and desktop gutter rules

#### Scenario: Dashboard renders on a narrow viewport
- **WHEN** a dashboard route uses the shared layout contract on a viewport narrower than 768px
- **THEN** content reflows for a single-column layout with documented mobile gutters
- **AND** display typography uses the documented compact mobile size
