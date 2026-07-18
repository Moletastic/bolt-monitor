## MODIFIED Requirements

### Requirement: Redirect behavior is preserved for destination changes
Flows whose successful result has a true destination change SHALL preserve server-action `redirect()` behavior. Mutations whose successful result remains on the current rendered route SHALL return typed `ActionState` instead of redirecting only to refresh data or communicate feedback. `redirect()` MUST remain outside `runServerAction` / `tryCatch` so Next.js redirect signals are not captured as errors.

#### Scenario: Successful mutation remains on current route
- **WHEN** a mutation completes and the operator should remain on the currently rendered route
- **THEN** the server action returns serializable typed action state
- **AND** the component renders the result without a same-route redirect or query-string feedback dependency

#### Scenario: Successful mutation changes destination
- **WHEN** a create, delete, or workflow transition has a distinct success destination
- **THEN** the server action redirects to that destination after the mutation succeeds
- **AND** the redirect is executed outside the result-catching boundary

#### Scenario: Navigation-first mutation fails before destination change
- **WHEN** a navigation-first mutation fails
- **THEN** the operator remains on an appropriate form or detail surface
- **AND** the typed dashboard error message is presented without exposing session credentials

## ADDED Requirements

### Requirement: Action-state forms prevent duplicate logical submissions
Components consuming typed `ActionState` SHALL expose the framework pending state immediately and SHALL prevent concurrent duplicate submissions of the same logical mutation until the current submission settles.

#### Scenario: Operator activates an action-state form
- **WHEN** the form submission begins
- **THEN** the initiating control displays an action-specific pending label or equivalent indicator
- **AND** duplicate submit controls for that logical action are disabled
- **AND** the pending state does not disable unrelated page interactions

#### Scenario: Action-state form settles
- **WHEN** the action returns success or error state
- **THEN** duplicate submit controls become available when the resulting domain state permits another submission
- **AND** stale pending text is removed

### Requirement: Action-state outcomes remain serializable and server-authoritative
Same-page action success and failure values SHALL remain serializable across the server action boundary. Success data MAY support local presentation, but SHALL NOT establish a browser-owned canonical resource cache.

#### Scenario: Same-page action succeeds
- **WHEN** a server action returns successful action state
- **THEN** the returned data contains only the minimum serializable information needed by the interaction
- **AND** affected server-rendered data is reconciled through narrow server-owned invalidation

#### Scenario: Same-page action fails
- **WHEN** a server action returns failed action state
- **THEN** the error arm carries the typed code, structured details, and optional safe message
- **AND** the component does not infer failure from a redirect query string
