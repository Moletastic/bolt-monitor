## 1. Sidebar navigation model

- [x] 1.1 Replace current dashboard shell nav items with module-oriented sidebar entries for `Dashboard`, `Services`, `Integrations`, `Audit Trail`, and `Config`.
- [x] 1.2 Update active-state logic so root `/` resolves to `Dashboard` while monitor overview, create, and detail pages resolve to `Services`.

## 2. Route restructuring

- [x] 2.1 Replace root dashboard page with a lightweight `Dashboard` WIP landing page.
- [x] 2.2 Move current monitor overview content from `app/page.tsx` to the `Services` module route.
- [x] 2.3 Add routable landing page for `Integrations` module.
- [x] 2.4 Add routable landing page for `Audit Trail` module.
- [x] 2.5 Add routable landing page for `Config` module.

## 3. Shell integration and verification

- [x] 3.1 Ensure root dashboard landing, services overview, and existing monitor pages continue to render inside shared shell without broken links or missing-route navigation.
- [x] 3.2 Update dashboard docs if needed to describe sidebar module structure.
- [x] 3.3 Verify dashboard lint or equivalent frontend checks still pass after sidebar changes.
