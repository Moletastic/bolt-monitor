## ADDED Requirements

### Requirement: System consumes the retry-safe canonical transition outbox
The DynamoDB single-table design SHALL use the single canonical tenant-scoped transition event/outbox item created atomically with recurring result and incident transition state by `make-check-execution-retry-safe`. This change SHALL add no competing transition item or producer protocol. The item SHALL be keyed by stable `eventId`, carry immutable causal identity and pending/acknowledged dispatch metadata, and remain the authority for Stream dispatch acknowledgement and repair.

#### Scenario: Incident changes notification-relevant state
- **WHEN** retry-safe execution commits an incident down or recovery transition
- **THEN** its transaction stores exactly one canonical outbox item containing safe immutable routing context and the stable incident activity transition identity
- **AND** notification assurance consumes rather than recreates it

#### Scenario: Queue acceptance is acknowledged
- **WHEN** the dispatcher confirms SQS acceptance for a canonical identity
- **THEN** it conditionally changes that same item from pending to acknowledged
- **AND** an ambiguous acknowledgement leaves the item pending

#### Scenario: Duplicate transition transaction is retried
- **WHEN** incident transition persistence is retried with the same transition identity
- **THEN** the table contains at most one corresponding outbox item

### Requirement: Pending dispatch has a durable bounded access path
Canonical pending transition and replay dispatch records SHALL remain queryable through a sparse tenant/time-bucketed access path until acknowledged. Automatic reconciliation SHALL bound buckets, pages, and items per invocation; manual repair SHALL use a point lookup by canonical identity. Pending records SHALL NOT expire merely because Stream retries exhaust, and no repair path SHALL scan unrelated table records.

#### Scenario: Stream retries exhaust
- **WHEN** a canonical dispatch insert exhausts DynamoDB Stream retries
- **THEN** the item remains pending in its sparse bucket
- **AND** its canonical identity supports direct manual repair

#### Scenario: Automatic reconciliation runs
- **WHEN** the reconciler searches for unacknowledged dispatch work
- **THEN** it reads only configured recent tenant/time buckets and bounded pages
- **AND** does not scan the primary table

#### Scenario: Dispatch is acknowledged
- **WHEN** an item transitions to acknowledged
- **THEN** it leaves the pending access path and receives the configured bounded acknowledged-record TTL

### Requirement: System stores incident-scoped notification delivery records
The DynamoDB single-table design SHALL define a canonical item pattern that groups delivery records under their incident for bounded API queries and enforces deterministic identity for transition, policy step, and channel. Delivery records SHALL persist safe operational fields only.

#### Scenario: Delivery record is written
- **WHEN** escalation work resolves a policy step and channel
- **THEN** the system stores the delivery under the incident partition with a sort key containing stable delivery ordering and identity
- **AND** a conditional write prevents duplicate identity creation

#### Scenario: Incident deliveries are queried
- **WHEN** the API lists delivery outcomes for an incident
- **THEN** it performs a bounded query of that incident partition
- **AND** does not scan unrelated incidents or tenants

#### Scenario: Delivery data is retained
- **WHEN** an incident and its delivery records become historical
- **THEN** delivery outcomes remain available with incident history under the repository's operational retention policy
- **AND** secret-bearing provider data is never stored in those records

### Requirement: System stores replay commands and idempotency records boundedly
Delivery replay SHALL use the canonical dispatch-record schema with `sourceKind=delivery_replay` and the same Stream dispatcher. The replay transaction SHALL store a tenant/incident/delivery/operation/key-scoped idempotency item containing a canonical request fingerprint and result identity. Idempotency records SHALL use a named bounded TTL longer than the maximum replay dispatch/retry window and SHALL be conditionally unique during retention.

#### Scenario: First replay request commits
- **WHEN** an eligible replay with a new `Idempotency-Key` is accepted
- **THEN** one transaction updates the delivery, creates one replay dispatch record, and creates one idempotency record

#### Scenario: Same replay request is retried
- **WHEN** the same key and request fingerprint are presented during retention
- **THEN** the stored result identity is returned without another delivery update or dispatch record

#### Scenario: Replay payload conflicts
- **WHEN** the same key is presented with a different request fingerprint during retention
- **THEN** the conditional operation rejects it as an idempotency conflict
