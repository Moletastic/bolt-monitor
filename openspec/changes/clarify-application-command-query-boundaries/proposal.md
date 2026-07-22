## Why

Go Lambda handlers and repositories currently mix mutation orchestration, reads, record mapping, default construction, and external dependency assembly. Naming application operations as commands or queries, combined with disciplined manual dependency injection, makes side effects, transaction ownership, and test seams clear without adopting CQRS infrastructure.

## What Changes

- Introduce explicit application command and query conventions for Go service operations.
- Require query paths to be side-effect free: they must not migrate, repair, audit, or otherwise write application state.
- Require commands to own validation, state transitions, audit intent, and transaction boundaries for their mutation.
- Formalize existing `main` functions as manual composition roots that build configuration and concrete AWS adapters, then inject narrow dependencies.
- Inject clocks, ID generation, sender registries, and event emitters where their behavior affects application outcomes or tests.
- Use narrow consumer-owned interfaces; exclude dependency-injection containers, service locators, command/query buses, event sourcing, and separate read storage.

## Capabilities

### New Capabilities
- `go-application-boundaries`: Command/query and manual dependency-injection conventions for Go Lambda applications.

### Modified Capabilities
- `code-patterns-foundation`: Add application dependency composition and consumer-owned interface requirements.

## Impact

- Affects Go code under `services/monitor-api`, `services/check-runtime`, `services/escalation-runtime`, and shared pure domain packages.
- Preserves API routes, envelopes, DynamoDB shapes, SQS payloads, and AWS deployment topology.
- Replaces read-time data migration/repair behavior with explicit command or operator migration paths where legacy persisted data requires it.
- Complements, but does not duplicate, `strengthen-domain-dynamodb-boundaries`; that change owns shared value/storage adapters while this change owns application operation boundaries and composition rules.
