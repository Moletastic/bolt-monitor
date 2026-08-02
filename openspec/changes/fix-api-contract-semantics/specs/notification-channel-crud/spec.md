## MODIFIED Requirements

### Requirement: Channel deletion blocked when referenced
The system SHALL refuse to delete a channel that any notification route references and SHALL return the standard error envelope with a stable typed conflict code and safe referencing-route details.

#### Scenario: Delete referenced channel
- **WHEN** a `DELETE /api/v1/notification-channels/{channelId}` request arrives AND at least one route step references `channelId`
- **THEN** the system returns `409 Conflict` with `status: "error"`
- **AND** `reason.code` identifies the channel-in-use conflict
- **AND** `reason.details` contains safe `referencingRoutes` entries with `policyId` and `name`

#### Scenario: Delete unreferenced channel
- **WHEN** a `DELETE /api/v1/notification-channels/{channelId}` request arrives AND no route references `channelId`
- **THEN** the system removes the channel and returns `204 No Content`
