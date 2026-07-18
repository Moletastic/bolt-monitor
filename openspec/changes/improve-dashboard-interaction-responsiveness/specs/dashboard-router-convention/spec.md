## MODIFIED Requirements

### Requirement: System prefers Link and server actions for navigation
The dashboard SHALL use `<Link>` from `next/link` for navigation between internal routes so Next.js can perform soft navigation and appropriate prefetching, and SHALL use server actions or `<form action={...}>` for state changes that follow navigation. Links SHALL retain default prefetch behavior unless measured cost, volatility, or scale justifies an explicit route-local override.

#### Scenario: Operator clicks an internal navigation link
- **WHEN** the operator activates an internal dashboard navigation link
- **THEN** the rendered element is a `<Link href="...">` and not an imperative `router.push(...)` call
- **AND** navigation does not perform a full document reload

#### Scenario: Dashboard renders a likely next destination
- **WHEN** an internal destination is eligible for framework prefetching
- **THEN** its `<Link>` retains default prefetch behavior
- **AND** an explicit prefetch override is added only with a documented performance reason and coverage

#### Scenario: Operator submits a navigation-first form
- **WHEN** the operator submits a dashboard form whose success changes destination
- **THEN** the form uses `<form action={serverAction}>` and server redirect semantics
- **AND** it does not call imperative client navigation after the mutation

#### Scenario: Operator submits a same-page form
- **WHEN** the operator submits a dashboard form whose success remains on the current route
- **THEN** the form uses a typed server action state
- **AND** it does not use a same-route redirect or imperative client router call to refresh feedback

## ADDED Requirements

### Requirement: System guards the polling router exception
The polling provider SHALL remain the only dashboard boundary permitted to use the client router for interval-driven server-data revalidation. Automated guards SHALL reject new imperative router navigation or refresh calls outside that boundary.

#### Scenario: Polling refreshes server-rendered data
- **WHEN** the polling interval requests fresh server-rendered data
- **THEN** the polling provider MAY invoke `router.refresh()` as non-urgent work
- **AND** it does not use `router.push()` or `router.replace()`

#### Scenario: Dashboard source introduces a client router call
- **WHEN** router-convention guard coverage scans or analyzes dashboard components
- **THEN** calls outside the polling provider fail the guard
- **AND** `usePathname()` remains limited to passive path inspection allowed by the existing convention
