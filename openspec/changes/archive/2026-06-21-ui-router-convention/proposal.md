## Why

The Next.js client router (`useRouter`, `usePathname`, `router.push`, `router.refresh`) is currently used in one place: `PollingProvider` calls `router.refresh()` on a 5-second polling tick to re-fetch server data. All other navigation uses `<Link>` or server actions. Codifying this as a documented convention prevents future drift toward imperative router calls where a `<Link>` or a server action would do.

## What Changes

- Document the convention in `AGENTS.md`: prefer `<Link>`, server actions, and `<form action={...}>` over `useRouter()` and `router.push()` for navigation. `router.refresh()` is reserved for the polling provider's interval-driven revalidation and is not used elsewhere.
- No code changes are required; the convention is already satisfied by the current codebase. Future PRs that introduce `useRouter`, `usePathname`, or `router.push` for navigation should be redirected to `<Link>`.
- This change captures the convention only. No new components, no spec-breaking behavior.

## Capabilities

### New Capabilities
- `dashboard-router-convention`: documentation of the client-router usage convention.

### Modified Capabilities
- (none)

## Impact

- `AGENTS.md`: add a "Router API usage" note documenting the convention.
- `openspec/specs/`: new spec file under `dashboard-router-convention`.
