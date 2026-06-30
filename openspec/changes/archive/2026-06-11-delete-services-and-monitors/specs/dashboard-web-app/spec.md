## ADDED Requirements

### Requirement: System supports service deletion through dashboard
System SHALL allow operators to permanently delete eligible services from the dashboard.

#### Scenario: Operator deletes service from service detail
- **WHEN** operator confirms deletion for an eligible service from the service detail view
- **THEN** system deletes the service through the service delete API
- **AND** system redirects the operator to the service list
- **AND** system refreshes dashboard service data so the deleted service is no longer shown

#### Scenario: Operator attempts to delete active service
- **WHEN** operator attempts to delete an active service from the dashboard
- **THEN** system explains that active services must be archived or otherwise made inactive before deletion
- **AND** system preserves the current service detail view

#### Scenario: Service deletion fails
- **WHEN** the service delete API rejects or fails a dashboard delete request
- **THEN** system shows an actionable error message
- **AND** system does not navigate as if deletion succeeded

### Requirement: System supports monitor deletion through dashboard
System SHALL allow operators to permanently delete eligible monitors from the dashboard.

#### Scenario: Operator deletes monitor from monitor detail
- **WHEN** operator confirms deletion for an eligible monitor from the monitor detail view
- **THEN** system deletes the monitor through the nested monitor delete API
- **AND** system redirects the operator to the parent service detail view
- **AND** system refreshes dashboard service and monitor data so the deleted monitor is no longer shown

#### Scenario: Operator attempts to delete last active-service monitor
- **WHEN** operator attempts to delete the only monitor under an active service
- **THEN** system explains why the monitor cannot be deleted in the current service state
- **AND** system preserves the current monitor detail view

#### Scenario: Monitor deletion fails
- **WHEN** the monitor delete API rejects or fails a dashboard delete request
- **THEN** system shows an actionable error message
- **AND** system does not navigate as if deletion succeeded
