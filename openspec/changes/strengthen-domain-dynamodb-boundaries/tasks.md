## 1. Characterize Existing Contracts

- [x] 1.1 Add table-driven characterization tests for existing tenant, service, monitor, incident, run, policy, and channel normalization forms.
- [x] 1.2 Add monitor-status and interval characterization tests covering persisted/API casing, allowed cadence, and invalid cadence behavior.
- [x] 1.3 Add DynamoDB record fixtures for legacy-cased shared records and assert unchanged PK, SK, entity type, and attribute names after round trips.
- [x] 1.4 Inventory monitor API, check runtime, and escalation runtime read methods that currently discard or cannot expose `LastEvaluatedKey`.

## 2. Add Shared Domain Values

- [ ] 2.1 Create a dependency-light shared domain value package with validated canonical identifier constructors and string serialization.
- [ ] 2.2 Define `MonitorRef` and migrate shared status-map/key call sites that require tenant, service, and monitor identity.
- [ ] 2.3 Define validated monitor-state conversion at status transition and adapter boundaries while preserving existing stored/API strings.
- [ ] 2.4 Define `CheckInterval` from supported cadence values with seconds and duration accessors.
- [ ] 2.5 Converge `auth.TenantID` on the neutral shared tenant value or add explicit adapter conversion without import cycles.
- [ ] 2.6 Add unit tests for invalid input, canonicalization, composite references, state conversion, and interval duration/seconds behavior.

## 3. Centralize Storage Records And Facade Use

- [ ] 3.1 Move escalation-policy and notification-channel DynamoDB records, constructors, and decoders from monitor API into `shared/dynamodbrecord`.
- [ ] 3.2 Move escalation-state record mapping into `shared/dynamodbrecord` and preserve its existing primary-key shape.
- [ ] 3.3 Replace escalation-runtime duplicate incident record mapping with shared incident records and constructors.
- [ ] 3.4 Migrate escalation runtime repository and tests from `shared/dynamodb`/direct AWS SDK types to `shared/aws.DynamoDBAPI`.
- [ ] 3.5 Consolidate shared monitor, status, execution-work, and incident conversion paths on the new value-object adapters.
- [ ] 3.6 Add record round-trip tests for each moved item family, including legacy-cased input and defensive-copy behavior for mutable config fields.

## 4. Make DynamoDB Reads Explicit And Reusable

- [ ] 4.1 Add a canonical exact-key helper using `shared/aws` and a typed primary-key key value.
- [ ] 4.2 Add a primary-index prefix-query page helper that accepts explicit limit, cursor, and sort direction and returns continuation state.
- [ ] 4.3 Add facade/helper tests proving exact-key input, prefix conditions, cursor propagation, ordering, and `LastEvaluatedKey` propagation.
- [ ] 4.4 Migrate monitor API monitor/history/incident primary-index reads to explicit page helpers without changing endpoint response behavior.
- [ ] 4.5 Migrate check runtime scheduler/worker primary-index reads to explicit page helpers and choose bounded behavior for each path.
- [ ] 4.6 Keep GSI audit and search queries as named methods; make their pagination and limits explicit without introducing a generic expression builder.

## 5. Narrow Monitor API Boundaries

- [ ] 5.1 Split monitor API repository implementation files by service, monitor, incident/history, scheduler, escalation/channel, and search capabilities.
- [ ] 5.2 Replace the monolithic handler repository interface with narrow domain-local interfaces while retaining one DynamoDB-backed implementation where transaction ownership requires it.
- [ ] 5.3 Split handler test fakes by domain and remove unrelated fake methods from each handler test fixture.
- [ ] 5.4 Migrate one complete monitor execution vertical slice to `MonitorRef`, then migrate remaining monitor-scoped repository calls.

## 6. Verify Compatibility And Quality

- [ ] 6.1 Add regression tests that assert unchanged REST response payloads and SQS payload fields for representative monitor, incident, escalation, and channel flows.
- [ ] 6.2 Add multi-page repository tests that prove callers either continue with a bounded cursor or preserve incompleteness instead of silently truncating data.
- [ ] 6.3 Run `make test-go-all`.
- [ ] 6.4 Run `make lint-go`.
- [ ] 6.5 Run `make check-api-contract` and `make check-bruno` to confirm unchanged API contracts and route coverage.
