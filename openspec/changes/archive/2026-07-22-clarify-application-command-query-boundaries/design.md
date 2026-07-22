## Context

The three Go Lambda services already use constructor injection in their `main` functions, but their application boundaries are inconsistent. The monitor API has one broad repository interface and handlers that combine HTTP adaptation, mutation policy, reads, and side effects. The check runtime embeds execution/incident transition decisions in its repository. The escalation policy read path can migrate legacy inline channels, making a GET request mutate persisted state.

The system needs a shared vocabulary that distinguishes writes from reads and makes external dependencies visible. It does not need infrastructure normally associated with full CQRS, such as command buses, event sourcing, separate data stores, or a dependency injection container.

## Goals / Non-Goals

**Goals:**

- Make every Go application operation identifiable as a command, a query, or an adapter concern.
- Prevent query execution from writing application state or triggering external side effects.
- Make commands own validation, state changes, audit intent, and transaction orchestration.
- Treat Lambda `main` functions as explicit composition roots that inject typed configuration and narrow ports.
- Improve deterministic tests by injecting time, identifier generation, notification senders, and security/audit emitters when outcomes depend on them.

**Non-Goals:**

- Introduce a command bus, query bus, mediator, service locator, reflection container, or code generation.
- Separate read and write databases, add projections, or adopt event sourcing.
- Change externally visible API, queue, storage, or authorization contracts.
- Require interfaces for pure domain functions, values, mappers, or stable implementation details.
- Replace the storage/value-object work planned by `strengthen-domain-dynamodb-boundaries`.

## Decisions

### 1. CQRS is a naming and side-effect discipline

Application operations use command or query naming in code, tests, and documentation. A command performs one business mutation and may return the updated entity or action result. A query returns data and has no application writes, audit writes, notification sends, migrations, repairs, or externally observable side effects.

HTTP handlers remain adapters: authorize, decode, construct command/query input, invoke the application operation, and map response/error envelopes. HTTP request DTOs do not become domain command types by alias; commands contain normalized application input.

Alternatives considered:
- Full CQRS platform: complexity without separate consistency, throughput, or storage needs.
- Keep informal terminology: hidden writes and mixed responsibilities remain difficult to detect in review.
- Require commands to return no data: artificial limitation for mutation endpoints that already return updated resources.

### 2. Commands own business mutation boundaries

Commands validate input, call pure domain transitions, assemble audit intent, and invoke the required repository transaction. Repositories own storage mechanics only. Pure transition functions remain independent of repositories and external clients.

Legacy persisted-data conversion cannot happen in a query. It moves to an explicit idempotent command, controlled migration tool, or deployment operation with auditable behavior. Selection depends on existing data volume and operator requirements at implementation time.

### 3. Manual constructor injection stays the only DI mechanism

Each Lambda `main` reads environment variables once, validates a typed config, creates concrete AWS clients/adapters, and passes dependencies to constructors. Constructors receive narrow interfaces defined by the consumer and values/configuration needed for behavior.

Use dependency structs only for an application component with several cohesive dependencies. Do not use a global registry or pass a broad container through the call stack. Constructors supply no production defaults for injected side-effecting behavior; composition roots choose defaults explicitly.

Alternatives considered:
- `fx`, `wire`, or a reflection container: dependency graph is small and static; added tooling obscures Lambda startup wiring.
- Globals for clock/senders: weakens test isolation and hides side effects.
- One interface per concrete helper: creates unnecessary mocks and indirection.

### 4. Interfaces belong to callers

Command/query handlers declare only the repository and external-port methods they consume. Shared facade interfaces remain authoritative for AWS SDK boundaries. Test fakes implement only a focused operation interface, keeping unrelated behavior out of unit fixtures.

### 5. Migration order follows behavior ownership

Start with one monitor execution result command because it crosses pure transition, incident/audit write intent, DynamoDB persistence, and notification publication. Then split monitor API service/monitor/incident/escalation operations. Query behavior is characterized before moving hidden read-time migration so legacy data behavior remains deliberate.

## Risks / Trade-offs

- [Command/query labels become cosmetic] → Add tests/guards for query write absence and organize files/interfaces by operation category.
- [Commands grow into god services] → Keep one command focused on one business outcome; extract pure domain transitions before adding collaborators.
- [Read-time migration removal breaks legacy records] → Characterize legacy route reads and introduce explicit idempotent migration before removing query writes.
- [Dependency structs hide required collaborators] → Use a cohesive typed dependency struct only at constructor boundary and keep fields private.
- [Too many interfaces] → Add interfaces only at external/variable boundaries and define them at consuming command/query boundary.
- [Active changes conflict] → Coordinate execution retry, notification delivery, value-object/storage, and pipeline-health state work before migrating those operations.

## Migration Plan

1. Add conventions and tests around existing composition roots and pure operation seams.
2. Characterize query side effects, especially legacy escalation-policy channel migration.
3. Extract a monitor execution-result command with injected clock, IDs, persistence, and publication ports while preserving behavior.
4. Migrate legacy inline-channel conversion to an explicit idempotent path, then make policy reads side-effect free.
5. Split monitor API operations and interfaces by service, monitor, incident, scheduler, escalation/channel, and search roles.
6. Validate Go tests/lint and API/Bruno contracts; deploy as code-only change with no public contract migration.

Rollback redeploys prior handlers. If an explicit data migration writes records, its idempotency marker and runbook define its own rollback/forward recovery before execution.

## Open Questions

- Whether legacy inline-channel conversion is best exposed as a protected operator command or a one-shot deployment tool depends on deployed legacy record inventory; implementation must choose and document one before removing read-time writes.
