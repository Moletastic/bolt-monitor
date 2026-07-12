## ADDED Requirements

### Requirement: System expires persisted execution work records
System SHALL attach TTL metadata to persisted execution work records so transient scheduler and worker coordination state is automatically removed after its operational troubleshooting window.

#### Scenario: Execution work is persisted
- **WHEN** system creates or updates an execution work record
- **THEN** the record includes numeric Unix epoch-second TTL metadata
- **AND** the TTL is later than the work record's accepted timestamp by the configured execution-work retention window

#### Scenario: Execution work retention elapses
- **WHEN** an execution work record reaches its TTL timestamp
- **THEN** the record is eligible for automatic deletion by DynamoDB Time to Live
- **AND** execution result history remains represented by `CheckRun` records, not by retained execution work records
