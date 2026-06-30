## Requirements

### Requirement: System exposes scheduler configuration through admin HTTP API
System SHALL allow administrators to read recurring execution control state through HTTP API.

#### Scenario: Administrator requests scheduler configuration
- **WHEN** administrator calls `GET /api/v1/admin/scheduler-config`
- **THEN** system returns current scheduler configuration state

### Requirement: System allows scheduler configuration updates through admin HTTP API
System SHALL allow administrators to change recurring execution control state through HTTP API.

#### Scenario: Administrator updates scheduler configuration
- **WHEN** administrator calls `PATCH /api/v1/admin/scheduler-config` with valid scheduler settings
- **THEN** system persists updated scheduler control state
- **AND** returns updated scheduler configuration resource

### Requirement: Scheduler admin API enforces recurring execution safety rules
System SHALL reject scheduler configurations that enable recurring execution without reliable stop control.

#### Scenario: Administrator enables recurring execution without stop control
- **WHEN** administrator submits scheduler configuration that enables recurring execution without valid stop control mode
- **THEN** system rejects request without applying configuration change

### Requirement: Scheduler control remains separate from monitor CRUD
System SHALL expose recurring execution control through an admin control-plane surface rather than through per-monitor CRUD routes.

#### Scenario: Client inspects scheduler control route
- **WHEN** client needs to pause, resume, or inspect global recurring execution control state
- **THEN** system provides scheduler configuration through `/api/v1/admin/scheduler-config`
- **AND** does not require monitor configuration endpoints to serve as scheduler control surface
