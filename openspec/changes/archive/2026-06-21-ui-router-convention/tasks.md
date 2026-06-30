## 1. AGENTS.md update

- [x] 1.1 Add a "Router API usage" subsection under "Gotchas" in `AGENTS.md` describing the convention: prefer `<Link>`, server actions, and `<form action={...}>`; reserve `router.refresh()` for the polling provider.
- [x] 1.2 Reference the `dashboard-router-convention` spec by name in the new note.

## 2. Verification

- [x] 2.1 Confirm a repo-wide `rg "useRouter|usePathname|router\.push"` search returns only the polling provider.
- [x] 2.2 Run `openspec validate ui-router-convention --strict` (if available) to confirm the change is well-formed.
