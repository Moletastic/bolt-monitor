## ADDED Requirements

### Requirement: Route steps reference channels by id
A notification route step SHALL be `{ channelId, delayMinutes }`. Steps MUST NOT carry `target` or `config`.

#### Scenario: Create route with referenced step
- **WHEN** a `POST /api/v1/escalation-policies` request body contains a step with `{ channelId, delayMinutes }`
- **THEN** the system persists the route and returns `201 Created`

#### Scenario: Reject step with inline target or config
- **WHEN** a request body contains a step with `target` or `config` populated
- **THEN** the system returns `400 Bad Request` with `{ error: "steps must reference channels by channelId; remove target and config" }`

#### Scenario: Reject step with unknown channelId
- **WHEN** a request body references a `channelId` that does not exist for the tenant
- **THEN** the system returns `400 Bad Request` with `{ error: "channel not found", channelId }`

### Requirement: Channel resolution at dispatch
When the escalation-runtime fires a step, the system SHALL resolve the referenced channel and dispatch using the channel's `target` and `config`.

#### Scenario: Resolve channel for step dispatch
- **WHEN** escalation-runtime processes a step with `channelId`
- **THEN** the system loads the channel record, merges `target` into the appropriate per-type config field (`chatId` for telegram, `toEmail` for email, `toNumber` for sms, `url` for webhook, `routingKey` for pagerduty), and passes the merged config to the registered sender

#### Scenario: Channel deleted mid-escalation
- **WHEN** escalation-runtime attempts to resolve a `channelId` that was deleted between schedule and dispatch
- **THEN** the system logs the missing channel and stops the escalation for that step without retry

### Requirement: Migration of inline-config routes
The system SHALL lazily migrate any existing route whose steps still carry inline `config`.

#### Scenario: First read of legacy route
- **WHEN** a route step has `config != nil` AND `channelId == ""`
- **THEN** the system creates a channel named "Migrated channel {n}" for that step, rewrites the step to reference the new channel, persists both the new channel and the rewritten route, and returns the rewritten route to the caller

#### Scenario: Migration idempotency
- **WHEN** the same legacy route is read twice in succession
- **THEN** the system does NOT create duplicate channels (deterministic IDs derived from `{tenantId}#{policyId}#{stepIndex}`)

### Requirement: Route editor channel picker
The dashboard route editor SHALL list available channels and require the operator to select one per step. Inline credential inputs SHALL NOT appear in the route editor.

#### Scenario: Add step shows channel picker
- **WHEN** an operator adds a step in the route editor
- **THEN** the dashboard shows a dropdown listing every channel by name (with type and target as secondary text); selecting a channel stores its `channelId`

#### Scenario: Step cannot be saved without channel
- **WHEN** an operator attempts to save a route whose step has no channel selected
- **THEN** the dashboard shows an inline error "Pick a channel for step N"

### Requirement: Service binding surface renamed
The dashboard service detail page SHALL show the assigned notification route under the heading "Notification route" (not "Escalation policy"). The assigned route links to its edit page.

#### Scenario: Service detail shows route label
- **WHEN** an operator views a service detail page
- **THEN** the page shows a section titled "Notification route" with the assigned route's name and a link to edit it