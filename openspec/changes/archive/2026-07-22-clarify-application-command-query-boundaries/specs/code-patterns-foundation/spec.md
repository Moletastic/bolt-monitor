## ADDED Requirements

### Requirement: Lambda main functions are manual composition roots
Each Go Lambda `main` function SHALL construct its typed configuration from environment input, create concrete AWS clients and adapters, and inject dependencies into application constructors. Application code SHALL NOT read process environment or construct concrete AWS clients outside the composition root.

#### Scenario: Lambda runtime starts
- **WHEN** a Go Lambda process starts
- **THEN** its `main` validates required environment configuration and constructs concrete dependencies once
- **AND** the started handler receives typed configuration and injected dependencies

### Requirement: Dependency injection is explicit and narrow
Application constructors SHALL receive only dependencies that affect their behavior. Interfaces SHALL be declared by the consuming application operation and SHALL contain only methods that operation uses. Dependency injection containers, service locators, and global mutable dependency registries SHALL NOT be used.

#### Scenario: Command is unit tested
- **WHEN** a command requires persistence, a clock, identifier generation, or a side-effecting sender/emitter
- **THEN** its test provides focused fake dependencies through the command constructor
- **AND** the test does not require a concrete AWS client or process-global dependency override

### Requirement: Side-effect defaults are composed explicitly
Constructors SHALL NOT silently create production clocks, identifier generators, sender registries, or security/audit emitters when those collaborators affect observable application behavior. Composition roots SHALL select production implementations explicitly.

#### Scenario: Production handler is assembled
- **WHEN** a Lambda composition root builds a handler or runtime
- **THEN** it explicitly supplies the production clock, identifier generator, sender registry, and emitter needed by that component
- **AND** tests can replace each collaborator without mutating the component after construction
