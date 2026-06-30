## ADDED Requirements

### Requirement: ServiceIcon displays technology icons for services
System SHALL provide a `ServiceIcon` component that renders technology-specific icons using Devicon.

#### Scenario: ServiceIcon renders known technology
- **WHEN** `ServiceIcon` receives `technologyKey` of "postgres"
- **THEN** component renders PostgresDevicon or equivalent Devicon for postgres
- **AND** icon displays at consistent 24x24 size

#### Scenario: ServiceIcon renders unknown technology with fallback
- **WHEN** `ServiceIcon` receives `technologyKey` that is not in the known catalog
- **THEN** component renders a generic server/database fallback icon
- **AND** fallback icon is visually distinct from technology-specific icons

#### Scenario: ServiceIcon renders without technology key
- **WHEN** `ServiceIcon` receives no `technologyKey` or undefined
- **THEN** component renders a generic service/host fallback icon

### Requirement: ServiceIcon supports Tier 1 + Tier 2 technologies
ServiceIcon SHALL support the following technology keys:

Tier 1: golang, mariadb, mysql, nginx, postgres, python, typescript
Tier 2: mongodb, redis, kafka, docker, apache, javascript, rabbitmq

#### Scenario: ServiceIcon has icons for all Tier 1 technologies
- **WHEN** component renders for each Tier 1 technology key
- **THEN** each renders its corresponding Devicon

#### Scenario: ServiceIcon has icons for all Tier 2 technologies
- **WHEN** component renders for each Tier 2 technology key
- **THEN** each renders its corresponding Devicon
