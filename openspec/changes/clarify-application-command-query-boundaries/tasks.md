## 1. Establish Application Boundary Conventions

- [x] 1.1 Document and add code-level conventions for named application commands, queries, handlers, ports, and composition roots in Go services.
- [x] 1.2 Add typed environment configuration constructors for monitor API, check runtime, and escalation runtime without changing environment variable contracts.
- [x] 1.3 Update each Lambda `main` to validate config, construct concrete dependencies once, and inject all behavior-affecting collaborators explicitly.
- [x] 1.4 Add composition-root tests for required configuration failures and production dependency assembly where feasible without AWS network access.

## 2. Extract Pure Execution Decisions

- [x] 2.1 Characterize current execution-result monitor status, threshold, incident, audit, and notification-transition behavior with table-driven tests.
- [x] 2.2 Extract pure monitor execution/incident transition decision logic from the check-runtime repository without changing retry-safe or notification-delivery semantics owned by active changes.
- [x] 2.3 Define focused persistence, publication, clock, and ID-generation ports consumed by the execution-result command.
- [x] 2.4 Implement the execution-result command with injected ports and migrate scheduler/worker handler orchestration to invoke it.
- [x] 2.5 Add unit tests proving transition decisions run without AWS clients and command tests prove required persistence/publication intent.

## 3. Remove Query Side Effects

- [x] 3.1 Characterize escalation-policy reads for legacy inline-channel records, including returned payload and current write behavior.
- [x] 3.2 Design and implement one explicit idempotent legacy inline-channel conversion path as either a protected command or repository-owned migration tool.
- [x] 3.3 Add migration progress/error handling, operator documentation, and tests for repeated safe execution.
- [x] 3.4 Remove inline-channel conversion writes from escalation-policy get/list query paths.
- [x] 3.5 Add regression tests proving policy reads do not write DynamoDB records, audit entries, notifications, or external side effects.

## 4. Split Monitor API Commands And Queries

- [ ] 4.1 Extract service command and query operations with consumer-owned narrow interfaces.
- [ ] 4.2 Extract monitor command, status, history, and manual-run operations with consumer-owned narrow interfaces.
- [ ] 4.3 Extract incident command and query operations with consumer-owned narrow interfaces.
- [ ] 4.4 Extract scheduler, escalation-policy, notification-channel, and search operations with consumer-owned narrow interfaces.
- [ ] 4.5 Reduce HTTP handlers to authorization, transport decode, application-operation invocation, and envelope mapping while preserving all routes and payloads.
- [ ] 4.6 Replace monolithic handler fakes with operation-focused test fakes.

## 5. Inject Observable Collaborators

- [ ] 5.1 Define explicit clock and identifier-generator dependencies for commands that create timestamps or IDs.
- [x] 5.2 Inject notification sender registries and security/audit emitters through composition roots rather than silently constructing them in application constructors.
- [ ] 5.3 Update tests to supply deterministic clocks, IDs, senders, and emitters through constructors instead of mutating constructed handlers.
- [x] 5.4 Confirm pure helpers and stable mappers remain concrete functions rather than unnecessary interfaces.

## 6. Verify Contracts And Architecture

- [x] 6.1 Add architecture guards that reject process environment reads and concrete AWS client construction outside Go composition roots.
- [x] 6.2 Add architecture guards or focused tests that reject writes from declared query operations.
- [x] 6.3 Run `make test-go-all`.
- [x] 6.4 Run `make lint-go`.
- [x] 6.5 Run `make check-api-contract` and `make check-bruno` to verify unchanged API routes and contracts.
