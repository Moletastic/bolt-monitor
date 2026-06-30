## Context

Current dashboard implementation already has service-first routing, nested monitor detail, incident reads, notification-channel APIs, scheduler config, and a shared shell. The gaps are mostly presentation and arrival-state gaps:

- `/services` cards only make the service name clickable, while card shape visually implies the whole card is actionable.
- Technology icons render without explicit size, so some devicons appear too small or indistinct.
- Service and monitor detail headers show raw IDs prominently, which is low-value for operators.
- Service summary currently reads like metadata, not monitoring health.
- Monitor overview lacks a protocol column on desktop and hides protocol only inside the name cell.
- Monitor detail current status does not surface latest error or enough triage context in one place.
- `/integrations` starts with an empty channel list until the operator manually refreshes.
- `/incidents` can show an empty table state, but the module should still explain incident lifecycle and available filters/actions.
- `/config` is a placeholder while scheduler/admin and setup assumptions already exist elsewhere.

## Mental Model

Operator surfaces should move from storage identity to operational identity:

```text
Storage-shaped UI                      Operator-shaped UI
─────────────────                     ───────────────────────────
serviceId                             service name + lifecycle
monitorId                             monitor name + protocol + target
raw status alone                      status + last check + error
placeholder settings                  scheduler + probe + environment state
manual refresh integrations           arrival data + loading/error state
```

## Proposed UX Shape

### Services View

```text
┌─────────────────────────────────────────────────────────────┐
│ [Go icon] API Gateway                              DOWN      │  whole card clickable
│ Public edge API for web traffic                              │
│                                                             │
│ Lifecycle active     Technology Go     Coverage 3/3 enabled │
│ Updated 2026-06-11                                      →   │
└─────────────────────────────────────────────────────────────┘
```

Service card click should navigate to service detail from any non-interactive card area. Any nested controls introduced later must not accidentally trigger navigation.

### Service Detail Summary

Service detail should answer: "What is this service, is it healthy, and is coverage complete?"

```text
┌──────────────────────── Service Summary ────────────────────────┐
│ [Tech icon] API Gateway                         [DOWN] [active] │
│ Public edge API for web traffic                                  │
│                                                                 │
│ Rollup        DOWN        Coverage        2/3 enabled            │
│ Monitors      3 total     Technology      Go                     │
│ Last update   Jun 11     Setup signal    1 disabled monitor      │
└─────────────────────────────────────────────────────────────────┘
```

Raw service ID should not be primary header text. If retained for debugging, place behind a low-emphasis metadata affordance such as "Copy service ID" or a collapsed details row.

### Monitor Overview

Desktop monitor overview should expose protocol as its own scannable column:

```text
Name             Protocol   Status   Enabled   Last check   Duration   Probe   Action
API /health      HTTP       UP       Enabled   2m ago       124ms      IAD     Disable
```

Mobile cards already show the protocol badge; remove raw monitor ID from primary body and use target/probe/status context instead.

### Monitor Detail Current Status

Monitor current status should be triage-first:

```text
┌──────────────────────── Current Status ─────────────────────┐
│ [HTTP] API /health                                  [DOWN]  │
│ GET https://api.example.com/health                         │
│                                                            │
│ Last outcome  timeout      Last check  Jun 11, 10:42       │
│ Duration      5000ms       Probe       IAD                 │
│ Cadence       every 60s    Enabled     Yes                 │
│                                                            │
│ Error: context deadline exceeded                            │
└────────────────────────────────────────────────────────────┘
```

This keeps configuration nearby but separates "what happened" from "how configured".

### Integrations

Integrations should preload channels on page arrival:

```text
mount → loading skeleton/spinner → channels table | empty state | error state
```

Manual refresh remains useful, but first meaningful data must not depend on pressing refresh.

### Incidents

Incident list can legitimately be empty. Empty should feel intentional:

- No open incidents: positive operational state.
- No closed incidents: no history in selected filter.
- API unavailable: explicit unavailable state.

If incident rows cannot show monitor/service names from current incident API, avoid making IDs the main visual label. Use summary, status, opened time, origin, and an action/link label like "Open monitor". A future API improvement may enrich incidents with service/monitor names.

### Settings

Settings should become control-plane overview, not placeholder copy:

- Scheduler recurring execution state and link/control to scheduler admin.
- Probe location catalog summary.
- Runtime/API configuration state if safely knowable from frontend.
- Bootstrap assumptions from shell can move here or be duplicated as setup context.

## Decisions

### Keep scope mostly dashboard-only

Most requested improvements are presentational and can use existing API shapes. Do not add backend work unless a human-readable incident label is required and cannot be derived without expensive fan-out.

### Hide IDs from primary operator surfaces

IDs remain valid route keys and form fields. They should not be primary display content in service cards, service summaries, monitor tables, dashboard matrices, recent incidents, or incident rows.

### Use status plus context, not status alone

Monitoring UI should pair every status with useful context: when checked, what target/protocol, which probe, how long it took, and latest error if failing.

## Risks / Trade-offs

- Full-card links can create invalid nested-interactive markup if buttons/forms are added inside cards. Keep card body as one link or use click handler carefully with semantic anchors.
- Larger devicons can dominate dense cards. Normalize with a fixed square container and responsive sizes.
- Incident rows may still lack human names. Avoid API fan-out unless necessary; consider backend enrichment later if incident list remains ambiguous.
- Settings can become a dumping ground. Keep first version limited to scheduler, probe locations, and environment/setup status.

## Open Questions

- Should raw IDs be completely hidden, or available through a low-emphasis "copy ID" affordance for support/debugging?
- Should Settings include editable scheduler controls, or only link to `/admin/scheduler` in this change?
- Should Incidents list enrich rows with service/monitor names now via existing service data, or wait for an API contract that returns names with incidents?
