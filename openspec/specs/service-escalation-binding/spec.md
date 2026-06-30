# service-escalation-binding Specification

## Purpose
TBD - created by archiving change notification-channels-and-routes. Update Purpose after archive.
## Requirements
### Requirement: Service binds a notification route
The system SHALL bind each service to a notification route via the service's `escalationPolicyId` field. The dashboard SHALL display the bound entity under the heading "Notification route" and link to its edit page.

#### Scenario: Service detail shows assigned route label
- **WHEN** an operator opens a service detail page
- **AND** the service has `escalationPolicyId` set
- **THEN** the page shows a section titled "Notification route" with the route's name and a link to its edit page

#### Scenario: Service has no route bound
- **WHEN** operator opens a service detail page
- **AND** the service has `escalationPolicyId` null
- **THEN** the page shows the heading "Notification route" with the body "No notification route assigned" and a secondary button "Assign a route"
