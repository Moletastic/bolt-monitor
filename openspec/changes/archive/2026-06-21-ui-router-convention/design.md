## Context

`useRouter`, `usePathname`, and `router.push` are Next.js client-side router APIs. They trigger client-side navigation, which can introduce subtle bugs (incorrect prefetch behavior, lost form state, focus issues) when used in places where a plain `<Link>` would suffice.

A repo-wide `grep` confirms the dashboard uses the client router in exactly one place:

- `apps/dashboard/components/polling-provider.tsx` calls `router.refresh()` on a 5-second interval and on `visibilitychange` to revalidate server-rendered data.

All other navigation uses `<Link>` or server actions. Codifying this convention in `AGENTS.md` and an OpenSpec capability makes it easier for future contributors to choose `<Link>` by default and reserve `router.refresh()` for the polling use case.

## Goals / Non-Goals

**Goals:**
- Document the convention in `AGENTS.md`.
- Capture the convention as a capability spec so future changes can reference it.

**Non-Goals:**
- No code refactor.
- No new lint rule (would require ESLint custom rule work; out of scope).
- No change to the polling provider's behavior.

## Decisions

- Add a single "Router API usage" subsection to `AGENTS.md` under "Gotchas" describing the convention.
- Add a capability spec `dashboard-router-convention` that states when `useRouter`, `usePathname`, and `router.push` may be used.
- Rationale: minimal change, maximum clarity. Codebase already complies.

## Risks / Trade-offs

- [Convention without an enforcement mechanism is easy to drift from] → Mitigation: PR review mentions the convention; capability spec is referenced from related dashboard requirements.

## Migration Plan

- Documentation-only change.
- Rollback: revert the PR.
