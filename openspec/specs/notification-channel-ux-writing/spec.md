# notification-channel-ux-writing Specification

## Purpose
TBD - created by archiving change notification-channels-and-routes. Update Purpose after archive.
## Requirements
### Requirement: User-facing copy uses operator language
The system SHALL use the following user-facing copy on the notification-channel and notification-route pages.

#### Scenario: Page headings
- **WHEN** an operator visits the channels list page
- **THEN** the H1 reads "Notification channels" and the page subtitle reads "Reusable destinations for alerts. Configure once, share across routes."

#### Scenario: Page headings for routes
- **WHEN** an operator visits the routes list page
- **THEN** the H1 reads "Notification routes" and the page subtitle reads "Order the channels that fire when an incident opens. Each step waits for the previous one."

#### Scenario: Empty state for channels
- **WHEN** the channel list is empty
- **THEN** the empty state shows the title "No channels yet" and the description "Create a Telegram bot, email sender, or webhook before assigning one to a route." with a primary button "Create your first channel"

#### Scenario: Empty state for routes
- **WHEN** the route list is empty
- **THEN** the empty state shows the title "No routes yet" and the description "Routes decide who hears about incidents and when. Start with one that pages the on-call engineer." with a primary button "Create your first route"

#### Scenario: Empty state for service binding
- **WHEN** a service has no route assigned
- **THEN** the service detail shows "No notification route assigned" with a secondary button "Assign a route"

### Requirement: Button labels use action verbs
Primary buttons SHALL use imperative verbs and the exact label shown below. No "Submit", "OK", "Save" alone.

#### Scenario: Channel buttons
- **WHEN** an operator is on a channel form
- **THEN** the primary button reads "Create channel" (create view) or "Save changes" (edit view); the destructive action reads "Delete channel"

#### Scenario: Route buttons
- **WHEN** an operator is on a route form
- **THEN** the primary button reads "Create route" (create view) or "Save changes" (edit view); the destructive action reads "Delete route"

### Requirement: Error messages explain the next step
Error messages SHALL tell the operator what to do, not what went wrong alone.

#### Scenario: Channel in use
- **WHEN** an operator tries to delete a channel that is referenced by routes
- **THEN** the error reads "This channel is used by 2 routes. Remove it from those routes before deleting."

#### Scenario: Missing channel in step
- **WHEN** an operator tries to save a route whose step has no channel
- **THEN** the inline error reads "Pick a channel for step 1"

#### Scenario: Channel not found at dispatch
- **WHEN** the runtime cannot resolve a `channelId` referenced by a route step
- **THEN** the operator-facing incident timeline entry reads "Route step skipped: channel was deleted. Update the route to continue."

### Requirement: Field labels are short and concrete
Form labels SHALL use the strings below. No placeholder-only labels.

#### Scenario: Channel field labels
- **WHEN** the channel form is rendered
- **THEN** the labels read: `Name` (with helper "What you'll recognize it as"), `Type`, `Target` (with helper "Where this channel delivers to"), and per-type credential labels as defined in the channel type metadata requirement

#### Scenario: Route field labels
- **WHEN** the route form is rendered
- **THEN** the labels read: `Name`, `Business hours path`, `Off-hours path`, `Business hours` (collapsible), `Channel` (per step), `Delay before firing` (per step)

### Requirement: Confirmation dialogs name the object
Delete confirmations SHALL quote the object's name and the impact.

#### Scenario: Channel delete confirmation
- **WHEN** an operator clicks "Delete channel"
- **THEN** the dialog text reads "Delete {name}? Routes using this channel will stop firing. This cannot be undone."

#### Scenario: Route delete confirmation
- **WHEN** an operator clicks "Delete route"
- **THEN** the dialog text reads "Delete {name}? Services using this route will fall back to their off-hours path with no escalation."
