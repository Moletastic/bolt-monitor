## MODIFIED Requirements

### Requirement: System provides operator dashboard web application
System SHALL provide an authenticated, authorized web application for operators to inspect target health, monitoring pipeline health, notification delivery health, evidence completeness, and setup gaps through a module-oriented console layout without conflating those concerns.

#### Scenario: Operator opens dashboard home
- **WHEN** operator navigates to the dashboard application
- **THEN** system shows an operational overview framed inside the shared dashboard sidebar shell
- **AND** the overview summarizes service target health, incident state, installation pipeline health/completeness, scheduler control state, notification delivery health, and setup gaps using available protected dashboard APIs

#### Scenario: Operator sees prioritized attention
- **WHEN** operator opens the dashboard home and there are down services, open incidents, delayed monitoring, failed notification delivery, disabled scheduler state, services without monitors, disabled monitor coverage, or draft services
- **THEN** system shows a prioritized attention area that identifies the items needing operator review
- **AND** each actionable item links to the existing module route, pipeline runbook, or remediation route where the operator can inspect or manage it

#### Scenario: Operator distinguishes operational failure domains
- **WHEN** target, execution, or notification evidence is unhealthy
- **THEN** the dashboard labels target failures as `DOWN`, monitoring freshness or execution failures as `DELAYED`, and terminal notification delivery failures as `FAILED`
- **AND** one state does not replace or imply either of the other states

#### Scenario: Operator sees incomplete evidence
- **WHEN** pipeline traversal, projection coverage, source version, or snapshot freshness is incomplete or stale
- **THEN** the dashboard labels the affected stage `UNKNOWN` or `INCOMPLETE`
- **AND** it never renders that stage as healthy

#### Scenario: Operator reviews service health matrix
- **WHEN** operator opens the dashboard home and services exist
- **THEN** system shows a compact service health matrix with service identity, target rollup status, lifecycle state, monitor coverage, and recent update context
- **AND** each service row links to the existing service detail route
- **AND** installation pipeline health is not presented as a per-service or per-monitor target status

#### Scenario: Operator opens dashboard with no configured services
- **WHEN** operator navigates to the dashboard application and no services exist
- **THEN** system shows an empty-state path to create the first service
- **AND** system does not show misleading zero-target-health summaries as if monitoring coverage exists
- **AND** system can still show installation pipeline health and scheduler control state

#### Scenario: Dashboard overview API context is unavailable
- **WHEN** one or more dashboard overview API requests fail
- **THEN** system shows an actionable unavailable or partial-state message for the affected overview content
- **AND** system preserves the shared dashboard shell and navigation
- **AND** unavailable pipeline health is not rendered as healthy

## ADDED Requirements

### Requirement: Dashboard exposes installation pipeline remediation
The dashboard SHALL present administrator-only installation scheduler, execution, notification, and evidence-completeness state with freshness and direct remediation context.

#### Scenario: Pipeline health is healthy
- **WHEN** the pipeline health API reports current, complete, healthy scheduler, execution, and notification stages
- **THEN** the dashboard shows the last evaluation time and a concise healthy installation state
- **AND** target-down services remain visible as independent target health items

#### Scenario: Pipeline stage requires attention
- **WHEN** a pipeline stage is `DELAYED`, `FAILED`, `UNKNOWN`, or `INCOMPLETE`
- **THEN** the dashboard shows the affected stage, exact or explicitly lower-bound safe aggregate evidence, freshness/completeness, and reason
- **AND** it provides a link to the supplied runbook or remediation location

#### Scenario: Unbounded unhealthy evidence exists
- **WHEN** aggregate unhealthy counts exceed the API evidence sample or evaluator traversal budget
- **THEN** the dashboard shows exact counts only after complete traversal and otherwise labels the observed count as a lower bound
- **AND** it does not attempt to render every affected run, incident, message, or monitor
- **AND** it does not create per-monitor CloudWatch controls or links

#### Scenario: Non-admin requests pipeline summary
- **WHEN** a caller lacks current authenticated tenant `ADMIN` authorization
- **THEN** the dashboard does not expose installation pipeline evidence
- **AND** it follows the authentication/RBAC failure behavior rather than bypassing the protected API
