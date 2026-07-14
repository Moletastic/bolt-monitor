## 1. Sidebar Implementation

- [x] 1.1 Add GitHub and external-link icon treatment to the sidebar utility area in `apps/dashboard/components/app-shell.tsx`.
- [x] 1.2 Render the fixed `https://github.com/Moletastic/bolt-monitor` URL with `target="_blank"`, `rel="noreferrer"`, and accessible link text.
- [x] 1.3 Keep the source link outside `navItems` and style it as a separated footer utility affordance across desktop and mobile layouts.

## 2. Verification

- [x] 2.1 Add or update dashboard tests covering source URL, label, new-tab semantics, and separation from active module navigation.
- [x] 2.2 Run `make lint-dashboard`, `make check-dashboard`, and `make test-dashboard`.
