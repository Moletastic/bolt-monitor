## 1. Schema and shared types

- [x] 1.1 Add `EntityNotificationChannel` + `NotificationChannelItem(tenantID, channelID)` to `shared/dynamodbschema/schema.go`
- [x] 1.2 Add `NotificationChannel` type to `shared/escalation/model.go` (or new `shared/notificationchannel` module) with `ChannelID`, `TenantID`, `Name`, `Type`, `Target`, `Config`, `CreatedAt`, `UpdatedAt`
- [x] 1.3 Update `EscalationStep` to drop `Channels` slice of `ChannelConfig` and replace with `ChannelID` (single channel per step) per design D1/D2
- [x] 1.4 Add `EscalationStep.ChannelID` resolution helper in `shared/escalation` that the runtime calls before dispatch
- [x] 1.5 Wire `shared/notificationchannel` into `go.work`

## 2. Monitor API — channel CRUD

- [x] 2.1 Add `monitorRepository` interface methods: `CreateNotificationChannel`, `ListNotificationChannels`, `GetNotificationChannel`, `UpdateNotificationChannel`, `DeleteNotificationChannel`, `ChannelsReferencedByRoutes(channelID)`
- [x] 2.2 Add `notificationChannelItemRecord` + marshaling helpers in `services/monitor-api/repository.go`
- [x] 2.3 Implement the five CRUD methods using `PK=TenantPK(tenantID)`, `SK=NOTIFICATION_CHANNEL#<channelID>`
- [x] 2.4 Add `POST /api/v1/notification-channels`, `GET /api/v1/notification-channels`, `GET /api/v1/notification-channels/{id}`, `PUT /api/v1/notification-channels/{id}`, `DELETE /api/v1/notification-channels/{id}` route dispatch
- [x] 2.5 Implement validation: required fields, allowed types, type-specific credential checks
- [x] 2.6 Implement redaction in responses: secret fields → `***REDACTED***`
- [x] 2.7 Implement `409 Conflict` on delete when channel is referenced, with `referencingRoutes` array
- [x] 2.8 Update `fakeMonitorRepository` in `services/monitor-api/main_test.go` with the new methods
- [x] 2.9 Add unit tests for create/list/get/update/delete + validation + redaction + delete-blocking
- [x] 2.10 Run `make test-go-all`, `make lint-go` until green

## 3. Route validation — strip inline target/config

- [x] 3.1 Update `policyFromRequest` / `validateEscalationPolicy` in `services/monitor-api/handler.go` to reject steps that include `target` or `config`
- [x] 3.2 Update `EscalationPolicyRequest` and response types so step payload is `{ channelId, delayMinutes }` only
- [x] 3.3 Update `cloneEscalationPath` and existing repository helpers in `services/monitor-api/repository.go` to drop inline `config`
- [x] 3.4 Update existing tests in `services/monitor-api/main_test.go` that construct policies with inline `config` (replace with channel fixtures)
- [x] 3.5 Run `make test-go-all`, `make lint-go` until green

## 4. Migration of legacy inline-config routes

- [x] 4.1 Add `MigrateRouteInlineChannels(ctx, policy *escalation.EscalationPolicy)` helper in `services/monitor-api/repository.go` that creates channels for any step with `config != nil`, rewrites the step to reference the new channel, and persists both
- [x] 4.2 Deterministic IDs: `{tenantID}#{policyID}#{stepIndex}` → channel ID; idempotent on re-read
- [x] 4.3 Call migration lazily inside `GetEscalationPolicy` and `ListEscalationPolicies` so existing routes migrate on first read
- [x] 4.4 Add unit test: feed a legacy route, read it twice, assert one channel created and step rewritten
- [x] 4.5 Run `make test-go-all`

## 5. Escalation runtime — channel resolution at dispatch

- [x] 5.1 Update `mergeChannelTarget` in `services/escalation-runtime/handler.go` to resolve `channelId` from the policy's `ChannelResolver` interface
- [x] 5.2 Add `ChannelResolver` repo method to escalation-runtime: `GetChannel(ctx, tenantID, channelID) (*escalation.NotificationChannel, error)`
- [x] 5.3 Wire `ChannelResolver` into `escalationHandler` constructor; keep the existing multi-sender registry intact
- [x] 5.4 Handle missing channel gracefully: log + skip step without retry (no `nil` panic)
- [x] 5.5 Update existing tests in `services/escalation-runtime/handler_test.go` to construct channels and reference them; verify dispatch path
- [x] 5.6 Add new test: dispatch with deleted channel → step skipped, log line emitted
- [x] 5.7 Run `make test-go-all`, `make lint-go`

## 6. Infra wiring

- [x] 6.1 Update `infra/stacks/bootstrap.ts` SST stack to wire the five new monitor API routes (`/api/v1/notification-channels[...]`)
- [x] 6.2 Update `go.work` and any SST build commands to include `services/monitor-api` changes (no new service)
- [x] 6.3 Run `make check-infra` + `make format-infra`

## 7. Dashboard — channels module

- [x] 7.1 Create `apps/dashboard/app/(monitoring)/integrations/channels/page.tsx` (list) with columns `Name`, `Type`, `Target`, `Updated`
- [x] 7.2 Create `apps/dashboard/app/(monitoring)/integrations/channels/new/page.tsx` (create form)
- [x] 7.3 Create `apps/dashboard/app/(monitoring)/integrations/channels/[channelId]/page.tsx` (edit form + delete button)
- [x] 7.4 Create `apps/dashboard/components/notification-channel-form.tsx` (client component) with type-conditional credential inputs (telegram bot token, email apiKey/fromEmail, SMS twilio creds, etc.)
- [x] 7.5 Add `NotificationChannel`, `CreateNotificationChannelPayload`, `UpdateNotificationChannelPayload` types to `apps/dashboard/lib/types.ts`
- [x] 7.6 Add `listNotificationChannels`, `getNotificationChannel`, `createNotificationChannel`, `updateNotificationChannel`, `deleteNotificationChannel` to `apps/dashboard/lib/api.ts`
- [x] 7.7 Add `createNotificationChannelAction`, `updateNotificationChannelAction`, `deleteNotificationChannelAction` server actions in `apps/dashboard/lib/actions.ts`
- [x] 7.8 Update `apps/dashboard/components/app-shell.tsx` nav: replace existing `Integrations` entry with `Channels` linking to `/integrations/channels`; rename `Escalation policies` to `Notification routes`
- [x] 7.9 Apply UX writing: empty states, button labels, field labels, delete confirm dialog per `notification-channel-ux-writing` spec
- [x] 7.10 Run `make lint-dashboard`, `make check-dashboard`, `make format-dashboard`

## 8. Dashboard — route editor uses channel picker

- [x] 8.1 Update `apps/dashboard/components/escalation-policy-form.tsx` to render a channel picker dropdown per step (no inline credential inputs)
- [x] 8.2 Replace per-channel credential rendering with a single `Channel` select that shows channel name + type + target
- [x] 8.3 Validate step has a selected channel before allowing save
- [x] 8.4 Update `apps/dashboard/lib/actions.ts` `parsePath` to reject steps with `target` or `config` populated server-side
- [x] 8.5 Pass available channels list to the route editor (new `policies/new` + `policies/[policyId]` pages)
- [x] 8.6 Run `make lint-dashboard`, `make check-dashboard`

## 9. Dashboard — rename to "Notification routes"

- [x] 9.1 Update page H1s + subtitles in `/policies`, `/policies/new`, `/policies/[policyId]` to use the new copy per UX spec
- [x] 9.2 Update service detail page section heading from "Escalation" to "Notification route" with link + "Assign a route" empty state button
- [x] 9.3 Update `apps/dashboard/components/service-form.tsx` to label the dropdown "Notification route" (not "Escalation policy")
- [x] 9.4 Verify all references to "Escalation policies" / "Escalation policy" replaced; run `rg "Escalation polic" apps/dashboard/` (excluding archive path) returns zero hits
- [x] 9.5 Run `make lint-dashboard`, `make check-dashboard`

## 10. Verification

- [x] 10.1 Run `make test-go-all`, `make lint-go`, `make build-go`
- [x] 10.2 Run `make check-dashboard`, `make lint-dashboard`, `make format-dashboard`, `make build-dashboard`
- [x] 10.3 Run `make check-infra`, `make format-infra`
- [x] 10.4 Run `make lint-all` to verify cross-cutting green
- [x] 10.5 Spot-check a legacy route through the API end-to-end (simulate via test): ensure migration creates a channel and rewrites the step

## 11. Cleanup

- [x] 11.1 Update `AGENTS.md` "Gotchas" block: replace any "channels hard-coded to iad" / "channels hard-coded" lines with "Route steps reference channels by channelId; configure channels under Integrations → Channels"
- [x] 11.2 Remove the one-time "We migrated your channel settings" banner code (or gate it behind a version flag) — not in scope for the spec, only flag for follow-up
- [x] 11.3 Verify no dead code: `rg "ChannelConfig" apps/dashboard/services/shared` shows only the type definition, no legacy inline-config consumers
