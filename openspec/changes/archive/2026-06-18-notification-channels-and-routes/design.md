## Context

The escalation-rules change (archived 2026-06-17) shipped channel credentials inline on each step of each route. Operators who reuse the same Telegram bot across many routes re-paste the bot token on every step. Rotation requires editing every route by hand. Secrets get duplicated across policy records, increasing blast radius if any one leaks.

The dashboard label "Escalation policies" reads as jargon. Operators route notifications — the noun they care about is a route, not the underlying state machine.

This change introduces a reusable channel registry, refactors route steps to reference channels by ID, renames the user-facing surface to "Notification routes", and writes UX copy against the Chipax guide.

## Goals / Non-Goals

**Goals:**
- One place to configure each transport/recipient pair. Same channel reusable across N routes.
- One place to rotate a credential. Deleting/updating a channel propagates to every route that uses it.
- Block deletion when the channel is in use; surface the conflicting routes in the error.
- Rename "Escalation policies" → "Notification routes" in the UI without breaking API contracts.
- Migrate existing inline-config routes without data loss.

**Non-Goals:**
- Per-step target override (a channel is fully self-contained).
- Secrets manager integration (env-only references). Out of scope; revisit later.
- Channel-level routing rules (e.g. rate limits per channel). Out of scope.
- Tenant-level channel sharing. Single-tenant (`DEFAULT`) like the rest of the system.

## Decisions

### D1: Channel owns both target and config

A channel is `{ type, name, target, config }`. `target` is the destination per channel type (chat ID, email, phone, webhook URL, PagerDuty routing key). `config` is the credential blob. Routes no longer carry target or config.

**Why:** One entity, one source of truth. If we kept target on the step and config on the channel, operators would have to update two places to change recipients. A channel that already encodes "where to send" plus "how to authenticate" is the smallest unit an operator reasons about.

**Alternative:** `target` per step, `config` shared via channel. More flexible (same bot to multiple chats) but creates a confusing UX: the channel becomes "a credential bundle", and operators have to re-enter the chat ID every time they reuse a bot. Reject.

### D2: ChannelId reference, not channel inline

Route step = `{ channelId, delayMinutes }`. No `type`, `target`, or `config` on the step.

**Why:** Forces reuse. Step is a pointer. Lookup happens at dispatch time inside escalation-runtime, where we already have the `MergeChannelTarget` logic — it now resolves from a channel record instead of the step itself.

**Alternative:** Allow either inline OR referenced. Two ways to do the same thing. Reject.

### D3: Migration on read, not in batch

When the new code reads a route that still has inline config (legacy data), it lazily backfills: for each step with `config != nil`, create a new channel (auto-named "Migrated channel {n}") and rewrite the step to `{ channelId, delayMinutes }`.

**Why:** Zero-downtime deploy. Existing routes keep working through the transition. No separate migration script.

**Alternative:** One-shot script during deploy. Cleaner but requires orchestration. Reject for v1.

### D4: API path stays `/api/v1/escalation-policies` for routes

The route CRUD endpoints keep their existing path. Only the dashboard renames the surface.

**Why:** Zero breaking API change for any external client. Internal entity stays `EscalationPolicy`; the dashboard calls them "Notification routes".

**Alternative:** Add `/api/v1/routes` and proxy. Double maintenance. Reject.

### D5: Channel CRUD endpoint under `/api/v1/notification-channels`

New path, new entity, clear separation.

**Why:** Aligns with the entity name and the dashboard URL `/integrations/channels`. Matches the proposed mental model.

### D6: Deletion blocked with conflict list

`DELETE /api/v1/notification-channels/{id}` returns `409 Conflict` with `{ referencingRoutes: [{ policyId, name }, ...] }` when in use.

**Why:** Operators need to know exactly which routes to update before they can delete the channel.

**Alternative:** Cascade delete. Reject — too destructive.

### D7: Channel name is the operator-facing label

Display name like "Payments on-call (telegram)". Same channel name shown wherever the channel is referenced (in route steps).

**Why:** Operators reason in names. The channelId is opaque to them.

## Risks / Trade-offs

- **Migration adds channels without operator consent** → Migration names are obviously auto-generated ("Migrated channel 1") so operators know what to rename. UI surfaces a one-time banner after first deploy telling them to rename migrated channels.
- **Renamed surface may confuse operators who already memorized "Escalation policies"** → Mitigation: route URLs stay `/policies` for now (UI label changes only). Operators see the new label in the sidebar immediately. No silent URL churn.
- **Secrets in channel records** → Same blast radius as today (DDB at-rest encryption is the system's existing posture). Out of scope to introduce a secrets manager in this change.
- **One channel = one recipient** → Operator creating 3 escalation steps to 3 different Telegram chats needs 3 channels. Acceptable — that is the cost of reuse.
- **Auto-naming migrated channels may collide on rerun** → Migration uses `{tenantId}#{policyId}#{stepIndex}` as deterministic IDs so re-running the migration is idempotent.

## Migration Plan

1. Deploy schema change (add `EntityNotificationChannel`) with no behavior change yet.
2. Deploy code that supports reading both inline-config and channel-ref steps. On read, migrate inline → channel + reference.
3. Deploy code that refuses to write inline-config routes (validation: `step.config` must be empty).
4. UI banner: "We migrated your channel settings. Review them under Integrations → Channels." Lives for one release cycle.
5. Rollback: redeploy prior code. Inline data still in records (migration is lazy on read, but if a write happens through new code, inline is stripped). Worst case, a route becomes unreadable; re-deploy old code to recover.

## Open Questions

- Should channels live in the same tenant partition as routes, or a separate one? Same partition for now — same `PK=TenantPK`, `SK=NOTIFICATION_CHANNEL#…`.
- Should the dashboard use the same `/policies` path or rename to `/routes`? Keeping `/policies` to avoid breaking bookmarks. URL is not part of the UX writing pass.
- Do we need a "test channel" action (sends a sample notification to one channel)? Nice-to-have, not in v1.