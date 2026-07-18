## ADDED Requirements

### Requirement: System preserves the shared shell during local loading
Dashboard navigation and local asynchronous interactions SHALL keep the resolved shared shell mounted. Loading UI SHALL be placed at the narrowest stable boundary that can represent the unresolved content without replacing unrelated navigation, context, or completed data.

#### Scenario: Operator navigates within the dashboard shell
- **WHEN** the destination route content is still resolving
- **THEN** shared navigation and shell context remain mounted and interactive
- **AND** only the unresolved destination segment shows its shape-matched fallback

#### Scenario: Local data island requests more data
- **WHEN** search, polling reconciliation, or history pagination is pending
- **THEN** already-resolved surrounding route content remains visible
- **AND** loading feedback is confined to the initiating island or unresolved subsection

### Requirement: System uses stable Suspense boundaries for independent server content
Routes with independently resolving server-owned content SHALL use stable Suspense boundaries where streaming provides a concrete continuity benefit. Boundary placement SHALL avoid whole-page fallback flashes and unnecessary remounting of resolved interactive islands.

#### Scenario: Independent server section suspends
- **WHEN** one independently fetchable route section has not resolved
- **THEN** its shape-matched fallback is shown within that section
- **AND** resolved sibling sections and the shared shell remain mounted

#### Scenario: Server content revalidates after a local mutation
- **WHEN** narrow revalidation causes affected server content to resolve again
- **THEN** the smallest affected boundary represents pending work
- **AND** unrelated client-island state, focus, and resolved content are preserved

#### Scenario: Suspense provides no independent streaming benefit
- **WHEN** content cannot resolve independently or a boundary would only add fallback churn
- **THEN** the implementation does not add a boundary solely for architectural uniformity
