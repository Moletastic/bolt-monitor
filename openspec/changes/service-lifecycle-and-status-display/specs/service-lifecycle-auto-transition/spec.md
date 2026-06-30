## ADDED Requirements

### Requirement: Service transitions from Draft to Active when first monitor is enabled
System SHALL automatically transition a service from "draft" lifecycle state to "active" when the first monitor on that service is enabled.

#### Scenario: First monitor enabled on draft service
- **WHEN** a monitor on a draft service is enabled
- **THEN** system SHALL transition the service lifecycle state from "draft" to "active"

#### Scenario: Subsequent monitors enabled on active service
- **WHEN** additional monitors on an already-active service are enabled
- **THEN** system SHALL NOT change the service lifecycle state

#### Scenario: Monitor disabled does not revert service to draft
- **WHEN** a monitor on an active service is disabled
- **THEN** system SHALL NOT change the service lifecycle state back to "draft"

### Requirement: Service lifecycle state is managed by system for Draft → Active transition
System SHALL automatically manage the Draft → Active transition. This transition is not initiated by user action but by system state change (monitor enablement).

#### Scenario: Transition occurs atomically with monitor enablement
- **WHEN** system enables a monitor on a draft service
- **THEN** both the monitor enablement and service lifecycle transition SHALL occur in a single transaction
- **OR** if transaction fails, neither change SHALL persist

### Requirement: Active → Archived transition remains manual
System SHALL only allow service lifecycle transition from "active" to "archived" via explicit user action.

#### Scenario: User archives active service
- **WHEN** user explicitly requests to archive an active service
- **THEN** system SHALL transition the service lifecycle state from "active" to "archived"

#### Scenario: User cannot manually set service to draft
- **WHEN** user attempts to manually change service lifecycle state to "draft"
- **THEN** system SHALL reject the request
