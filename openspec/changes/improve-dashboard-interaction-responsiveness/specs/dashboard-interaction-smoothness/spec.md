## MODIFIED Requirements

### Requirement: System makes same-page mutations visibly continuous
Same-page dashboard mutations SHALL return typed action state and provide visible local continuity while the action is pending and after it completes. The operator SHALL be able to understand which control changed, whether work is pending, and what the final server-confirmed result was without a redirect back to the current route.

#### Scenario: Operator submits a same-page mutation
- **WHEN** the operator submits a mutation whose successful outcome remains on the current route
- **THEN** the action returns typed pending, success, or error state to the local interaction surface
- **AND** the route is not redirected merely to display feedback or refresh the same page

#### Scenario: Same-page mutation is pending
- **WHEN** a same-page mutation is in flight
- **THEN** the initiating control shows action-specific pending feedback
- **AND** all controls that could submit the same logical mutation are disabled until it settles
- **AND** unrelated controls and already-resolved page content remain interactive

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
- **AND** server-rendered data remains the canonical state

## ADDED Requirements

### Requirement: System limits optimistic UI to safe reversible operations
The dashboard SHALL use optimistic presentation only for operations that are explicitly classified as safe and reversible in the interaction implementation. Every optimistic change SHALL preserve enough prior state to roll back and SHALL reconcile with the server action result and subsequent server-rendered truth.

#### Scenario: Safe reversible operation is submitted
- **WHEN** an interaction explicitly supports an optimistic update
- **THEN** the affected local presentation updates immediately
- **AND** the server action remains the authority for whether the operation succeeded
- **AND** duplicate submissions for the same logical operation remain disabled

#### Scenario: Optimistic operation fails
- **WHEN** the server rejects or cannot complete an optimistic operation
- **THEN** the local presentation rolls back to its captured prior state
- **AND** the failure is announced in the interaction context
- **AND** no stale optimistic value is retained as canonical data

#### Scenario: Operation is destructive or not safely reversible
- **WHEN** an operation cannot be reliably rolled back or has a true destination change
- **THEN** the dashboard does not optimistically present it as complete
- **AND** existing confirmation and server-confirmed navigation behavior is preserved

### Requirement: System narrowly revalidates mutation dependencies
Dashboard server actions SHALL revalidate only the route segments or cache dependencies whose server-rendered data can be changed by the completed mutation. Revalidation SHALL occur after a successful mutation and SHALL NOT replace the typed error path for failed same-page actions.

#### Scenario: Same-page mutation succeeds
- **WHEN** a mutation changes data shown by the current detail route and one parent summary route
- **THEN** the action invalidates only those dependent surfaces
- **AND** unrelated dashboard route trees are not invalidated

#### Scenario: Mutation fails
- **WHEN** a server action returns a typed failure
- **THEN** the action does not trigger success-only revalidation
- **AND** the current resolved UI remains available while local failure feedback is shown

### Requirement: System preserves a server-first data ownership boundary
The dashboard SHALL retain Server Components, server actions, server-side data truth, and the opaque server-side session boundary as its default architecture. Client components SHALL be limited to focused interaction islands that receive safe serializable data or invoke server-owned interfaces.

#### Scenario: Dashboard route renders canonical data
- **WHEN** an operator loads or navigates to a dashboard route
- **THEN** canonical route data is fetched and rendered through the server architecture
- **AND** browser code is not given refresh tokens, backend credentials, or a client-owned canonical cache

#### Scenario: Search, polling, or history pagination needs local interaction state
- **WHEN** search input coordination, polling cadence, or incremental history pagination requires browser interaction
- **THEN** that behavior is contained in a focused client island
- **AND** the surrounding route, shell, and canonical resource data remain server-owned
- **AND** no new client data library is introduced without a concrete requirement that existing React and Next.js primitives cannot satisfy

### Requirement: System announces interaction status and preserves useful focus
Changed dashboard interactions SHALL expose pending, success, failure, and rollback status to assistive technology and SHALL preserve or deliberately move focus according to the interaction outcome.

#### Scenario: Same-page action reports status
- **WHEN** a same-page action enters pending state or settles
- **THEN** status text is programmatically associated with or announced near the initiating interaction
- **AND** success uses a non-interruptive status announcement
- **AND** failure or rollback uses an alert or equivalently assertive announcement

#### Scenario: Same-page action settles without navigation
- **WHEN** a same-page action succeeds or fails without changing destination
- **THEN** focus remains on the initiating control or moves to the first actionable validation target
- **AND** focus does not fall back to the document body

#### Scenario: Successful action changes destination
- **WHEN** a create, delete, or other navigation-first action redirects after success
- **THEN** the destination provides a sensible focus target and a single result announcement

### Requirement: System guards interaction responsiveness behavior
The dashboard SHALL include behavioral and performance-oriented automated coverage for changed interaction flows. Coverage SHALL verify observable behavior and bounded work rather than relying only on implementation-string assertions.

#### Scenario: Mutation responsiveness guards run
- **WHEN** dashboard tests exercise a changed same-page mutation
- **THEN** they verify immediate pending feedback, duplicate-submission prevention, typed success and failure, and optimistic rollback when applicable
- **AND** they verify that unrelated resolved content remains mounted and usable

#### Scenario: Navigation and data-work guards run
- **WHEN** dashboard responsiveness guards exercise navigation, polling, search, or pagination
- **THEN** they verify no full document navigation is introduced for internal links
- **AND** they verify request, refresh, or revalidation work is bounded to the interaction being performed
- **AND** they verify the shared shell is not replaced by local loading work
