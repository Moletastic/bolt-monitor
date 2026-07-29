## ADDED Requirements

### Requirement: Incident delivery route takes precedence over incident detail
The API dispatcher SHALL route `GET /api/v1/incidents/{incidentId}/deliveries` to incident delivery listing before generic incident-detail matching.

#### Scenario: Operator requests incident deliveries through API Gateway
- **WHEN** an authenticated request targets `GET /api/v1/incidents/{incidentId}/deliveries`
- **THEN** the response contains incident-scoped delivery outcomes under the standard envelope
- **AND** the request is not handled as generic incident detail
