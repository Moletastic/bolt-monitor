## ADDED Requirements

### Requirement: Exclusions do not fabricate incident recovery
System SHALL NOT resolve or recover an incident solely because maintenance or disabled intervals exclude scheduled opportunities.

#### Scenario: Maintenance starts with incident open
- **WHEN** a monitor enters maintenance while an incident is open
- **THEN** covered scheduled slots are excluded from objective accounting
- **AND** the incident is not resolved by the exclusion itself

#### Scenario: Maintenance ends with incident open
- **WHEN** maintenance ends while an incident remains open
- **THEN** the monitor becomes `UNKNOWN` awaiting recurring observation
- **AND** incident recovery waits for a new qualifying recurring result under existing recovery rules

#### Scenario: Manual success occurs after maintenance
- **WHEN** a manual check succeeds before a qualifying post-maintenance recurring result
- **THEN** the success does not resolve the incident by default

#### Scenario: Missing expected slots are finalized
- **WHEN** independent finalization classifies one or more mature expected slots as `missing` because scheduler or pipeline evidence is absent
- **THEN** those accounting classifications do not fabricate a target failure, recovery transition, or notification event
- **AND** incident behavior continues to use canonical recurring results and existing retry-safe recovery thresholds
