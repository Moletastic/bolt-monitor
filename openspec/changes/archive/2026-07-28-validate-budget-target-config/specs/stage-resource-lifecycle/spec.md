## ADDED Requirements

### Requirement: Persistent target budget configuration fails closed
When budget configuration is supplied, the target validator SHALL require a finite positive USD amount and one or more non-empty alert email addresses as one paired configuration. A persistent target SHALL require valid paired budget configuration unless it declares a documented explicit FinOps opt-out; malformed or partial fields SHALL fail validation before AWS mutation.

#### Scenario: Persistent target has valid budget configuration
- **WHEN** a persistent target provides positive budget amount and alert recipients
- **THEN** target validation succeeds and conditional budget infrastructure remains enabled

#### Scenario: Persistent target has malformed budget configuration
- **WHEN** a persistent target supplies only one budget field or an invalid amount or recipient list
- **THEN** validation fails before deploy
- **AND** it does not silently disable budget alerts
