## MODIFIED Requirements

### Requirement: System defines retention for raw run history
System SHALL define raw run retention expectations for high-volume `CheckRun` records and persist TTL metadata that DynamoDB can use to delete expired raw run items.

#### Scenario: Raw runs accumulate over time
- **WHEN** system persists ongoing execution results
- **THEN** raw run items include numeric Unix epoch-second TTL metadata set to the configured raw-run retention window
- **AND** the TTL metadata is compatible with the primary table's DynamoDB Time to Live configuration
