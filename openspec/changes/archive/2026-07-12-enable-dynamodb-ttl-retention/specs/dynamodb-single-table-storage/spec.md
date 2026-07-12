## ADDED Requirements

### Requirement: System enables native TTL for eligible DynamoDB items
The primary DynamoDB application table SHALL enable DynamoDB Time to Live on the `TTL` attribute so item families that write numeric epoch-second TTL values can expire without custom cleanup jobs.

#### Scenario: Table is provisioned
- **WHEN** infrastructure provisions the primary application DynamoDB table
- **THEN** the table has DynamoDB Time to Live enabled on the `TTL` attribute
- **AND** item families that do not write `TTL` remain persistent

## MODIFIED Requirements

### Requirement: System defines retention strategy for high-volume run data
System SHALL define and enforce a retention strategy for raw check-run history items using native DynamoDB TTL.

#### Scenario: Check-run history is persisted over time
- **WHEN** system stores high-volume run results
- **THEN** storage design includes explicit TTL retention expectations for raw run items
- **AND** the primary DynamoDB table is configured so expired raw run items are eligible for automatic deletion
