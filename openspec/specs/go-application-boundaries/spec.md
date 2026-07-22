## Purpose

Go application operations use explicit command/query boundaries. Handlers adapt transport concerns, commands orchestrate mutations through narrow ports, and queries remain side-effect free.

## Requirements

### Requirement: Go application operations are commands or queries

Each Go service SHALL classify an application operation that crosses an HTTP, SQS, scheduled-event, or runtime handler boundary as a command or query. Commands SHALL perform one business mutation; queries SHALL return data without application side effects.

#### Scenario: Operator invokes a mutation endpoint
- **WHEN** an authorized operator invokes a service, monitor, incident, scheduler, channel, or policy mutation endpoint
- **THEN** the handler invokes a named application command
- **AND** the command owns validation, domain transition invocation, and write orchestration for that mutation

#### Scenario: Operator invokes a read endpoint
- **WHEN** an authorized operator invokes a read endpoint
- **THEN** the handler invokes a named application query
- **AND** the query returns data without persisting application changes, emitting audit records, sending notifications, or invoking an external side effect

### Requirement: Legacy data conversion is explicit

The system SHALL execute legacy persisted-data conversion through an explicit idempotent command, migration tool, or deployment operation. Query operations SHALL NOT perform data migration, repair, or backfill.

#### Scenario: Legacy escalation policy contains inline channel data
- **WHEN** application reads an escalation policy containing a legacy inline channel representation
- **THEN** the read returns data without writing the policy or creating a notification channel
- **AND** an explicit idempotent conversion path can migrate the legacy representation when required

### Requirement: Commands compose pure domain decisions with ports

Commands SHALL invoke pure domain functions for deterministic business decisions and SHALL use injected ports for persistence, message publication, notification delivery, time, identity generation, and audit/security emission when those dependencies affect command outcomes.

#### Scenario: Execution result changes incident state
- **WHEN** an execution result command processes a monitor state transition
- **THEN** the state and incident decision is testable without a DynamoDB, SQS, or notification client
- **AND** the command maps that decision to its required persistence and publication actions through injected ports

### Requirement: HTTP and Lambda handlers are adapters

HTTP and Lambda handlers SHALL authorize/decode event input, invoke an application command or query, and map its result to the existing event response. Handlers SHALL NOT construct AWS clients, directly implement domain transition policy, or own repository transaction construction.

#### Scenario: Handler processes authenticated monitor request
- **WHEN** the monitor API handler receives an authenticated request
- **THEN** it resolves authorization and transport input before invoking the application operation
- **AND** it returns the existing response envelope without changing route or payload contracts
