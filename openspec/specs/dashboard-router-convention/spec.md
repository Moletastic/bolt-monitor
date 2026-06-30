## ADDED Requirements

### Requirement: System prefers Link and server actions for navigation

The dashboard SHALL use `<Link>` from `next/link` for navigation between routes, and SHALL use server actions or `<form action={...}>` for state changes that follow navigation.

#### Scenario: Operator clicks a navigation link

- **WHEN** the operator activates any navigation link in the dashboard
- **THEN** the rendered element is a `<Link href="...">` and not an imperative `router.push(...)` call

#### Scenario: Operator submits a form

- **WHEN** the operator submits a dashboard form
- **THEN** the form uses `<form action={serverAction}>` and not an imperative `router.push(...)` after a client-side mutation

### Requirement: System restricts client router APIs to polling-driven revalidation

The dashboard SHALL only call client router APIs for polling-driven server-data revalidation, and SHALL NOT call `useRouter`, `usePathname`, or `router.push` outside the polling provider.

#### Scenario: Codebase uses client router for revalidation

- **WHEN** a component needs to re-fetch server data on an interval
- **THEN** the component MAY call `router.refresh()` inside `PollingProvider` or an equivalent provider component
- **AND** the component SHALL NOT use `router.push` for navigation

#### Scenario: Codebase avoids client router for navigation

- **WHEN** a new dashboard component needs to navigate to another route
- **THEN** the component uses `<Link href="...">` rather than calling `useRouter().push(...)` or `useRouter().replace(...)`
- **AND** the component uses `usePathname()` only if it needs to inspect the current path (for example to mark the active sidebar item), not to drive navigation
