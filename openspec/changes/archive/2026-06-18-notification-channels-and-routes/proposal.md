## Why

The escalation feature shipped with channel credentials inlined on every step of every route. That works for a single route, but real operators reuse the same Telegram bot, email sender, or PagerDuty integration across many services. Today there is no way to share a configured channel: each new route re-enters bot tokens and API keys, rotation requires editing every route by hand, and secrets sprawl across policy records.

We also need to rename "escalation policies" — that name describes the internal mechanic, not what an operator does with the page. Operators route notifications, and the resulting object is a **notification route**.

This change splits channel configuration into a reusable entity and refactors routes to reference it by ID.

## What Changes

- Add a new `notification-channel` entity that owns channel type, display name, destination target, and credential config. One channel = one recipient/transport pair (e.g. one Telegram bot to one chat, one SendGrid sender, one Twilio number).
- Recreate the `/integrations/channels` dashboard module: list, create, edit, delete. Channel names are operator-friendly labels; the page shows type, target, and a redacted preview of credentials.
- Rename the dashboard "Escalation policies" surface to "Notification routes". The data model keeps `escalationPolicy` entity names internally for migration cost, but the UI label and primary user-facing copy use "Notification routes" (route = a single ordered set of steps that fan out across configured channels).
- Routes reference channels by `channelId`. A route step is `{ channelId, delayMinutes }`. The runtime resolves `channelId` to the stored channel config before dispatching.
- Routes can no longer carry inline credentials. Existing routes with inline config are migrated on read: any step with `config` populated gets its own channel created during the migration, and the step is rewritten to reference the new channel.
- Channel deletion is blocked when a route references it, with a clear error pointing at the affected routes.
- Backend: new endpoints under `/api/v1/notification-channels`. Route endpoints stay at `/api/v1/escalation-policies` (internal name) but the dashboard calls them `/api/v1/routes` after the rename — or, simpler, keep the path and only rename the UI label so API contracts stay stable.
- UX writing pass: every screen in `/integrations/channels` and the routes editor follows the Chipax UX writing guide — operator-first, no jargon, action verbs on buttons, plain English for empty states.

## Capabilities

### New Capabilities

- `notification-channel-crud`: registry entity with list/create/get/update/delete endpoints and a dashboard CRUD module under `/integrations/channels`.
- `notification-route-channel-reference`: route steps reference channels by ID; runtime resolves config from the channel before dispatch; deletion of a referenced channel is blocked.
- `notification-channel-ux-writing`: copy patterns for the channel registry and route editor (labels, buttons, empty states, error messages) consistent with the Chipax UX writing guide.

### Modified Capabilities

- `escalation-policy-crud`: route steps lose inline `config`; required to be `channelId` only. This is the only spec-level behavior change.
- `service-escalation-binding`: no behavior change, but the UI label changes from "Escalation policy" to "Notification route".

## Impact

- **Backend Go**: `shared/dynamodbschema` adds `EntityNotificationChannel` + `NotificationChannelItem(...)`. `services/monitor-api` gains the channel CRUD endpoints and the route references resolver. `services/escalation-runtime` looks up the channel by ID before firing a step instead of merging inline `ChannelConfig.config`. Existing tests that build inline `ChannelConfig` need a helper that creates a channel first.
- **Frontend (Next.js)**: `/integrations` page restored under `/integrations/channels`. Routes editor (`/policies`) gains a channel-picker dropdown; the per-channel credential inputs are removed. Dashboard copy updates across `app-shell.tsx`, page headers, button labels.
- **Migration**: existing routes with inline config are auto-converted on first read into channels + route-step references. No data loss; one-time backfill in the same deploy.
- **Docs**: AGENTS.md gotcha block updates (no longer "channels hard-coded to iad", now "route steps reference channels").