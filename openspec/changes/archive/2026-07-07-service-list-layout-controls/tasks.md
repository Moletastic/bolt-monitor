## 1. Service List Layout

- [x] 1.1 Replace the existing Services page operations card with the new desktop header, description, and equal health indicators.
- [x] 1.2 Render active, draft, and down-now indicators as equal peers with compact labels suitable for desktop and mobile.
- [x] 1.3 Preserve deleted-service feedback, empty state, unavailable state, and service status toast behavior.

## 2. List Controls

- [x] 2.1 Add a controls row above the service card list with a service search field and filter affordances.
- [x] 2.2 Add the desktop `Create service` action to the controls row using link-based navigation to `/services/new`.
- [x] 2.3 Apply search behavior to narrow the visible service cards without changing backend API behavior.

## 3. Responsive Behavior

- [x] 3.1 On narrow viewports, show only the health indicator row at the top of visible content while preserving an accessible page heading.
- [x] 3.2 On narrow viewports, render search and filter controls below the health indicators and service cards one per row.
- [x] 3.3 On narrow viewports, render `Create service` as a bottom-right floating action button with an accessible name and safe spacing.

## 4. Loading And Verification

- [x] 4.1 Update the Services loading skeleton to mirror the new layout structure.
- [x] 4.2 Run `make lint-dashboard`.
- [x] 4.3 Run `make check-dashboard`.
- [x] 4.4 Run `make test-dashboard` if search/filter behavior adds or changes dashboard tests.
