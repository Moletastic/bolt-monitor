## MODIFIED Requirements

### Requirement: System defines canonical key patterns for core item families
System SHALL define canonical partition and sort key conventions for core item families used by monitoring workflows.

#### Scenario: Core entity type is mapped to storage
- **WHEN** system maps a monitor, status, run, incident, or audit record to DynamoDB
- **THEN** it uses documented PK/SK conventions for that item family, including append-only `CheckRun` items and mutable `MonitorStatus` items
