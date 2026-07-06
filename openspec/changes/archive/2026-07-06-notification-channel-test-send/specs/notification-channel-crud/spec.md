## ADDED Requirements

### Requirement: Dashboard channel detail includes test action
The dashboard notification channel detail page SHALL include a non-destructive test-send action for existing channels in addition to edit and delete controls.

#### Scenario: Existing channel detail shows send test action
- **WHEN** an operator views an existing notification channel detail page
- **THEN** the page includes a `Send test` action
- **AND** the action is visually distinct from destructive delete controls

#### Scenario: New channel form does not show send test action
- **WHEN** an operator is creating a new notification channel that has not been saved
- **THEN** the dashboard does not show a `Send test` action
