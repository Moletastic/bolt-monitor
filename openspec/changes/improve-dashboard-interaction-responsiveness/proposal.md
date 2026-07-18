## Why

Dashboard interactions currently mix responsive same-page action state with same-route redirects, broad cache invalidation, and manually coordinated client requests. This makes common mutations and data exploration feel slower or less stable than the underlying server work while also creating inconsistent pending, focus, and announcement behavior.

## What Changes

- Convert mutations that remain on the current route to typed `ActionState` results with immediate pending feedback and duplicate-submission prevention; retain redirects for successful creates, deletes, or other actions whose result has a true destination change.
- Permit optimistic presentation only for explicitly identified safe, reversible operations, reconcile every optimistic result with server-rendered truth, and roll back on typed failure.
- Narrow server-action revalidation to the routes affected by each mutation and preserve the shared shell through stable Suspense/loading boundaries.
- Keep Server Components and server actions as the default architecture while using focused client islands for interactive search, polling, and incremental history pagination.
- Use declarative `Link` navigation so Next.js soft navigation and appropriate prefetching remain available; do not expand imperative router usage beyond the polling convention.
- Standardize accessible pending, success, failure, rollback, and focus behavior for changed interactions.
- Add behavioral guards and performance-oriented tests for duplicate submissions, optimistic rollback, revalidation scope, stable boundaries, island scope, and router conventions.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dashboard-interaction-smoothness`: Define responsive same-page mutation ownership, safe optimistic reconciliation, narrow revalidation, focused client islands, accessibility continuity, and responsiveness guard coverage.
- `dashboard-ui-action-state-results`: Require typed action state for mutations that do not change destination and define pending/duplicate-submission behavior while preserving redirect semantics for destination changes.
- `dashboard-router-convention`: Require declarative links to preserve soft navigation and intentional prefetch behavior without broadening imperative router access.
- `dashboard-loading-states`: Preserve the shared shell and resolved content through granular Suspense/loading boundaries during navigation and local data loading.

## Impact

- Affects dashboard server actions, mutation forms, search, polling, monitor-history pagination, route loading boundaries, navigation links, and dashboard guard/unit tests under `apps/dashboard`.
- Does not change monitor API contracts, authentication/session handling, the opaque server-side API boundary, or the Next.js server-first deployment architecture.
- Adds no client canonical cache, direct browser access to refresh tokens or backend credentials, static-site migration, broad visual redesign, or new data-fetching library without a demonstrated need.
