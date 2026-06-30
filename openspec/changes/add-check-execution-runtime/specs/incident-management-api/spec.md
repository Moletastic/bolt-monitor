## ADDED Requirements

### Requirement: System opens incidents from failing execution outcomes
System SHALL create or update system-owned incidents from non-success monitor execution outcomes.

#### Scenario: First failing execution occurs for monitor
- **WHEN** system processes a completed execution result for a monitor whose outcome is not success and no open incident exists for that monitor
- **THEN** system creates a new open incident associated with that monitor

#### Scenario: Additional failing execution occurs while incident is open
- **WHEN** system processes another non-success execution result for a monitor that already has an open incident
- **THEN** system updates that existing incident instead of creating a second concurrent open incident for the same monitor

### Requirement: System resolves open incidents from recovery outcomes by default
System SHALL resolve a monitor's open incident when later execution indicates recovery.

#### Scenario: Monitor recovers after open incident
- **WHEN** system processes a success execution result for a monitor with an open incident
- **THEN** system marks that incident resolved through system-owned business logic
