## Context

The service detail page currently combines a large summary card, inline edit form, archive/create actions, and delete card in a two-column layout. This makes routine operations compete with destructive actions and pushes monitors lower. The monitor table is wrapped in an overview card and includes redundant columns: protocol already appears beside monitor name, and probe location is deprecated for the dashboard.

The page also needs recent alert context for incidents tied to monitors under the service. Existing dashboard helpers can list global incidents or monitor-specific incidents, but using those naively on service detail either reads unrelated data or creates N+1 calls across monitors.

## Goals / Non-Goals

**Goals:**

- Make service detail read as an operational overview first.
- Move primary service actions to the top-right: edit service, archive service, create monitor.
- Replace summary/edit/delete sections with a full-width service info banner and separate danger zone.
- Show uptime, P99 latency, and error rate as the main service-level metrics.
- Simplify monitor list by removing card shell and redundant/deprecated columns.
- Show recent alerts near monitors, with desktop side-by-side layout and mobile stacked layout.
- Add service-level incident read endpoint to avoid broad or N+1 incident fetches.

**Non-Goals:**

- Change service archive, delete, or monitor creation semantics.
- Add service edit modal/drawer behavior unless already available.
- Redesign monitor detail tabs.
- Add probe location back to dashboard monitor tables.
- Introduce new incident mutation behavior.

## Decisions

### Desktop Layout

Desktop layout should keep actions and service overview dense:

```txt
Services / Payments API                         [Edit] [Archive] [Create monitor]

┌────────────────────────────────────────────────────────────────────────────┐
│ [icon] Payments API      [UP]                    Uptime | P99 latency | Err │
│        Service description                        99.95% | 184 ms      | .12%│
└────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────┐ ┌──────────────────────────┐
│ Monitors                                     │ │ Recent alerts            │
│ plain table                                  │ │ incident links           │
└──────────────────────────────────────────────┘ └──────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────┐
│ Danger Zone                                                               │
│ Delete warning + delete action                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

### Mobile Layout

Mobile should stack content in operational priority order:

```txt
Actions
Service info card
Recent alerts
Monitors table/list
Danger Zone
```

### Metrics

Use these indicators in the service info banner:

- `Uptime`: recent uptime percentage from service card metrics where available.
- `P99 latency`: p99 latency from service card metrics where available.
- `Error rate`: derived as failed samples / total samples where available.

Fallbacks should be explicit: `No data`, `-`, or equivalent low-emphasis unavailable state. Avoid fake zeros when no samples exist.

### Actions

Top-right action buttons use outlined styling so routine operations stay visible without overpowering the service info banner:

- `Edit service`: outlined link or button with pencil icon.
- `Archive service`: outlined existing archive action with archive icon; hide or disable for archived services.
- `Create monitor`: outlined link to `/services/{serviceId}/monitors/new` with plus/radio icon; disable for archived services.

Destructive delete moves to Danger Zone only.

### Recent Alerts Data Source

Prefer a service-scoped incidents endpoint:

```txt
GET /api/v1/services/{serviceId}/incidents?limit=5
```

The endpoint should return incidents tied to any monitor under the service, sorted newest first, bounded by limit. This reduces dashboard cost versus listing global incidents and filtering, and avoids calling monitor incident endpoint once per monitor.

Recent alert links should take operators to incident context. Preferred link:

```txt
/services/{serviceId}/monitors/{monitorId}?tab=incidents
```

If incident detail is more actionable for a result, the result can also expose `/incidents/{incidentId}` as secondary or future behavior.

### Monitor Table

Remove columns:

- `Probe location`, because probe location is deprecated in dashboard UI.
- `Protocol`, because protocol badge already appears in the name cell.

Keep name, status, enabled state, last check/update context, and actions.

## Risks / Trade-offs

- **Metrics unavailable** -> Show explicit unavailable state; do not imply zero errors or perfect uptime without samples.
- **Desktop row too dense** -> Keep recent alerts width constrained and monitor table as primary column.
- **Service incidents query cost** -> Use existing incident GSI or bounded service/monitor lookup; enforce limit.
- **Archive/delete action confusion** -> Top action handles reversible archive, Danger Zone handles permanent delete.
- **Archived service actions** -> Disable create monitor and archive action consistently while keeping read-only context visible.
