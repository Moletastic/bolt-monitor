## ADDED Requirements

### Requirement: MonitorProtocolBadge displays styled text for monitor types
System SHALL provide a `MonitorProtocolBadge` component that renders text pills for monitor protocol types.

#### Scenario: HTTP monitor displays HTTP badge
- **WHEN** `MonitorProtocolBadge` receives `type` of "http"
- **THEN** component renders "HTTP" as a styled text pill

#### Scenario: HTTPS monitor displays HTTPS badge
- **WHEN** `MonitorProtocolBadge` receives `type` of "http" with https target
- **THEN** component renders "HTTPS" as a styled text pill

#### Scenario: TCP monitor displays TCP badge
- **WHEN** `MonitorProtocolBadge` receives `type` of "tcp"
- **THEN** component renders "TCP" as a styled text pill

#### Scenario: gRPC monitor displays gRPC badge
- **WHEN** `MonitorProtocolBadge` receives `type` of "grpc"
- **THEN** component renders "gRPC" as a styled text pill

#### Scenario: DNS monitor displays DNS badge
- **WHEN** `MonitorProtocolBadge` receives `type` of "dns"
- **THEN** component renders "DNS" as a styled text pill

### Requirement: MonitorProtocolBadge is text-only (no icons)
System SHALL render protocol badges as styled text, not icons.

#### Scenario: Badge uses text styling only
- **WHEN** `MonitorProtocolBadge` renders any protocol
- **THEN** badge contains only styled text characters
- **AND** no icon elements are rendered
