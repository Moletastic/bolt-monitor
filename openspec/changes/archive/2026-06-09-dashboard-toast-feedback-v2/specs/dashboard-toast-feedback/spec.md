## ADDED Requirements

### Requirement: Toast on successful service creation
System SHALL show a success toast when a service is created successfully.

#### Scenario: Service created successfully
- **WHEN** user completes the service creation form and it succeeds
- **THEN** system SHALL show a success toast with message "Created successfully"
- **AND** the toast SHALL auto-dismiss after 4 seconds

### Requirement: Toast on successful monitor creation
System SHALL show a success toast when a monitor is created successfully.

#### Scenario: Monitor created successfully
- **WHEN** user completes the monitor creation form and it succeeds
- **THEN** system SHALL show a success toast with message "Created successfully"
- **AND** the toast SHALL auto-dismiss after 4 seconds

### Requirement: Toast when service goes DOWN
System SHALL show a destructive toast when a service status changes to DOWN.

#### Scenario: Service transitions to DOWN
- **WHEN** polling detects service status changed to DOWN
- **THEN** system SHALL show a destructive toast with message "Service is DOWN"
- **AND** the toast SHALL auto-dismiss after 6 seconds

### Requirement: Toast when service goes UP (after being DOWN)
System SHALL show a success toast when a service that was DOWN transitions back to UP.

#### Scenario: Service recovers from DOWN to UP
- **WHEN** polling detects service status changed from DOWN to UP
- **THEN** system SHALL show a success toast with message "Service is UP again"
- **AND** the toast SHALL auto-dismiss after 4 seconds

### Requirement: Toast on action errors
System SHALL show a destructive toast when a user action fails.

#### Scenario: Service creation fails
- **WHEN** user submits a service creation form and it fails
- **THEN** system SHALL show a destructive toast with the failure message
- **AND** the toast SHALL auto-dismiss after 6 seconds

### Requirement: Toast on manual run trigger
System SHALL show a success toast when a manual run is triggered.

#### Scenario: Manual run triggered
- **WHEN** user clicks "Run now" button and the run is triggered
- **THEN** system SHALL show a success toast with message "Manual run triggered"
- **AND** the toast SHALL auto-dismiss after 4 seconds

### Requirement: Toasts are non-blocking
System SHALL ensure toasts do not block user interaction or navigation.

#### Scenario: User navigates while toast is visible
- **WHEN** a toast is visible and user navigates to another page
- **THEN** the toast SHALL remain visible briefly then auto-dismiss
- **AND** navigation SHALL complete without waiting for toast