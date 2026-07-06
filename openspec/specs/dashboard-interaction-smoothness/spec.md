## Requirements

### Requirement: System preserves router-safe smoothness boundaries
Dashboard smoothness improvements SHALL preserve the existing declarative routing model. Converted same-page mutation flows SHALL NOT introduce `useRouter`, `router.push`, or `router.replace` outside the polling provider.

#### Scenario: Same-page mutation is converted for smoother feedback
- **WHEN** a dashboard mutation flow is changed to show inline pending, success, or error feedback without leaving the current route
- **THEN** the flow uses a server action or typed action state rather than imperative client router navigation
- **AND** navigation-first flows continue to use server-action navigation unless a separate spec explicitly changes the UX

#### Scenario: Navigation remains route-driven
- **WHEN** an operator activates a dashboard navigation control
- **THEN** the rendered navigation uses `<Link href="...">` or server-action redirect semantics
- **AND** the component does not call `router.push(...)` or `router.replace(...)`

### Requirement: System shows one feedback surface per mutation event
The dashboard SHALL assign each mutation event to a single user-visible feedback surface. The same success or error event SHALL NOT render both a toast and an inline page banner for the same operator action.

#### Scenario: Redirect destination receives query feedback
- **WHEN** a redirected dashboard route receives a success or error query parameter from a mutation
- **THEN** exactly one feedback surface presents the result to the operator
- **AND** duplicate toast and inline banner feedback for the same event is not shown

#### Scenario: Same-page action returns typed state
- **WHEN** a same-page mutation returns typed action state
- **THEN** the component renders the pending, success, or error result in the local interaction context
- **AND** the component does not also depend on query-parameter feedback for the same result

#### Scenario: Error feedback is shown
- **WHEN** a dashboard mutation fails
- **THEN** the operator sees an actionable error message derived from the typed dashboard error message rules
- **AND** the error feedback is exposed through an accessible alert or equivalent announced feedback surface

### Requirement: System makes same-page mutations visibly continuous
Same-page dashboard mutations SHALL provide visible local continuity while the action is pending and after it completes. The operator SHALL be able to understand which control changed, whether work is pending, and what the final result was without relying on a full route cut.

#### Scenario: Operator toggles a same-page monitor state
- **WHEN** the operator enables or disables a monitor from a list or detail surface that remains on the same route
- **THEN** the affected control shows a pending state while the mutation is in flight
- **AND** completion feedback appears in the local page context without duplicate global feedback
- **AND** the final enabled or disabled state is reconciled with server-rendered data

#### Scenario: Operator changes same-page incident state
- **WHEN** the operator acknowledges or resolves an incident from a surface that remains on the same route
- **THEN** the affected action shows pending feedback while the mutation is in flight
- **AND** the page presents the completed incident state without requiring the operator to infer success from a route reload

#### Scenario: Same-page mutation fails
- **WHEN** a same-page mutation fails after showing pending feedback
- **THEN** the affected control returns to a safe interactive state
- **AND** the operator sees a local error message explaining the failure

### Requirement: System treats polling refresh as background work
Polling-driven dashboard refreshes SHALL be scheduled as non-urgent background updates and SHALL avoid redundant refresh bursts for a single visibility transition.

#### Scenario: Polling interval refreshes server data
- **WHEN** the polling provider refreshes dashboard server data on an interval
- **THEN** the refresh is scheduled as non-urgent UI work
- **AND** operator-initiated navigation or form submission remains higher priority than the polling refresh

#### Scenario: Hidden dashboard tab becomes visible
- **WHEN** the dashboard tab becomes visible after being hidden
- **THEN** the polling provider refreshes server data no more than once for that visibility transition
- **AND** regular polling resumes without creating overlapping intervals

### Requirement: System avoids nested interactive controls
Dashboard cards, rows, and list items SHALL NOT nest interactive controls inside other interactive controls. Mobile and desktop layouts SHALL preserve predictable tap, keyboard, and screen-reader behavior for navigation and inline actions.

#### Scenario: Mobile card contains a row-level destination and an inline action
- **WHEN** a mobile dashboard card provides both navigation to a detail route and an inline mutation action
- **THEN** the navigation target and mutation control are rendered as separate sibling interactive elements
- **AND** activating the mutation control does not also activate the card navigation

#### Scenario: Keyboard operator tabs through dashboard cards
- **WHEN** the operator navigates dashboard cards with the keyboard
- **THEN** focus moves through each link and button in a logical order
- **AND** no focusable element is nested inside another focusable element

### Requirement: System verifies loading and focus continuity for touched flows
Dashboard smoothness work SHALL preserve existing loading-state and destructive-action focus requirements for every route or mutation flow modified by this change.

#### Scenario: Changed route depends on server data
- **WHEN** this change modifies a dashboard route that fetches server data
- **THEN** the route has a shape-matched loading placeholder for the destination surface
- **AND** the placeholder respects existing skeleton styling and reduced-motion requirements

#### Scenario: Changed destructive flow completes successfully
- **WHEN** this change modifies a destructive delete flow and the deletion succeeds
- **THEN** the required post-delete navigation still occurs
- **AND** focus lands on a sensible next target rather than falling back to `<body>`

#### Scenario: Changed destructive flow fails
- **WHEN** this change modifies a destructive delete flow and the deletion fails
- **THEN** the operator remains on the current surface
- **AND** the failure is announced through the selected accessible feedback surface
