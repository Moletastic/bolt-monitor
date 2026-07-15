## MODIFIED Requirements

### Requirement: Monitor execution location is not operator-configurable
System SHALL execute monitors from its managed runtime without exposing a probe-location API or dashboard selection control.

#### Scenario: Operator creates or updates a monitor
- **WHEN** an operator configures a monitor
- **THEN** the request does not include a probe-location identifier
- **AND** the API exposes no probe-location discovery route
