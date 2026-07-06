## ADDED Requirements

### Requirement: Dashboard exposes notification route deletion
The dashboard SHALL allow operators to delete an unreferenced notification route from its edit page using the existing escalation-policy delete API.

#### Scenario: Operator deletes an unreferenced notification route
- **WHEN** an operator confirms deletion from a notification route edit page for a route that is not referenced by any service
- **THEN** the dashboard SHALL call the delete escalation-policy action with that route ID
- **AND** the operator SHALL be redirected to the notification routes list with a deletion success message

#### Scenario: Operator attempts to delete a referenced notification route
- **WHEN** an operator confirms deletion from a notification route edit page for a route that is still referenced by a service
- **THEN** the monitor API SHALL reject the delete request with `POLICY_REFERENCED`
- **AND** the dashboard SHALL keep the operator on the route edit page with an inline error message
