## Context

The dashboard shell currently has sidebar navigation but no global top bar. Operators must navigate to Services, service detail, Notification routes, or Channels before finding a specific resource. The monitor API already owns services, monitors, escalation policies, and notification channels. DynamoDB storage currently supports tenant-scoped service refs under `PK=TENANT#<tenant>` / `SK=SERVICE#<serviceId>`, monitor refs under `PK=SERVICE#<tenant>#<service>` / `SK=MONITOR#<monitorId>`, escalation policies under `PK=TENANT#<tenant>` / `SK=ESCALATION_POLICY#<policyId>`, and notification channels under `PK=TENANT#<tenant>` / `SK=NOTIFICATION_CHANNEL#<channelId>`.

Naively searching across services, monitors, policies, and channels by listing all resources would either scan too much or perform multiple broad partition reads. The new search surface needs a bounded tenant-scoped lookup pattern.

## Goals / Non-Goals

**Goals:**

- Add a global top-bar search control to the dashboard shell.
- Search services, monitors, escalation policies, and notification channels from one input.
- Return typed results that the dashboard can render with resource-specific iconography, labels, and links.
- Use an optimized DynamoDB access pattern that avoids scans and avoids reading full resource collections per keystroke.
- Keep search behavior predictable through normalization, debounce, min query length, result limits, and feedback states.

**Non-Goals:**

- Full-text ranking, fuzzy matching, typo tolerance, stemming, or semantic search.
- Searching incidents, audit events, runs, or settings in this change.
- Introducing OpenSearch, Algolia, or another external search system.
- Searching secret channel config fields or monitor headers.
- Cross-tenant search.

## Decisions

### Add `GET /api/v1/search`

The endpoint should accept query parameters:

```txt
GET /api/v1/search?q=<query>&limit=<n>&types=service,monitor,policy,channel
```

Defaults:

- `limit`: 8 or 10 total results.
- `types`: all supported searchable types.
- Minimum normalized query length: 2 characters.

Response data should be wrapped in the existing response envelope and include typed result objects:

```json
{
  "results": [
    {
      "type": "service",
      "id": "svc_...",
      "label": "Payments API",
      "description": "Service · active · 3 monitors",
      "href": "/services/svc_...",
      "iconKey": "service:api",
      "matchText": "payments api"
    },
    {
      "type": "monitor",
      "id": "public-http",
      "serviceId": "payments",
      "label": "Homepage availability",
      "description": "Monitor · Payments API · https://example.com",
      "href": "/services/payments/monitors/public-http",
      "iconKey": "monitor:http",
      "matchText": "homepage availability"
    },
    {
      "type": "policy",
      "id": "primary-route",
      "label": "Primary route",
      "description": "Notification route · 2 business-hours steps · 1 off-hours step",
      "href": "/policies/primary-route",
      "iconKey": "policy",
      "matchText": "primary route"
    },
    {
      "type": "channel",
      "id": "primary-on-call",
      "label": "Primary on-call",
      "description": "Channel · PagerDuty · escalation@example.com",
      "href": "/integrations/channels/primary-on-call",
      "iconKey": "channel:pagerduty",
      "matchText": "primary on-call"
    }
  ]
}
```

The API may compute `href` server-side so the dashboard does not need to replicate resource route rules. Returning `type` remains useful for icon/rendering and analytics.

### Use sparse search index records

Add one or more compact search-index records per searchable resource. These records live in the existing table and tenant partition:

```txt
PK = TENANT#<tenant>
SK = SEARCH#<normalized-prefix-or-term>#<resourceType>#<resourceStableKey>
EntityType = SearchIndex
ResourceType = service | monitor | policy | channel
ResourceID = ...
ServiceID = ...         // present for monitors and services where useful
Label = ...
Description = ...
Href = ...
IconKey = ...
MatchText = ...
UpdatedAt = ...
```

The query pattern is:

```txt
KeyCondition: PK = TENANT#DEFAULT AND begins_with(SK, SEARCH#<normalized-query>)
Limit: bounded, e.g. 25 before in-memory de-dupe/ranking
```

This uses the primary index, avoids scans, and avoids new GSI capacity/cost. It does increase write amplification because create/update/delete operations maintain search records.

### Index normalized prefixes from selected fields

Search should normalize text by trimming, lowercasing, removing repeated whitespace, stripping unsafe punctuation into spaces, and tokenizing into meaningful terms. To support prefix matching without scans, write bounded prefixes for selected tokens and phrase starts.

Recommended token prefix rules:

- Ignore tokens shorter than 2 characters.
- Store prefixes from length 2 through max 16 for each indexed token.
- Store a phrase prefix for the normalized primary label from length 2 through max 24.
- Cap indexed tokens per resource to avoid unbounded writes.

Field criteria beyond ULID/opaque IDs:

| Resource | Primary fields | Secondary fields | Excluded fields |
|---|---|---|---|
| Service | `name`, normalized `serviceId` slug | `description`, `serviceCategory`, `lifecycleState`, `rollupStatus` | business-hours internals, tenant ID |
| Monitor | `name`, normalized `monitorId` slug | parent service name, HTTP target hostname/path, monitor type, enabled state | request headers, expected body text, full secret-bearing URLs beyond safe target display |
| Escalation policy | `name`, normalized `policyId` slug | `description`, route structure summary, referenced channel IDs as safe references | inline channel configs, notification secrets |
| Channel | `name`, normalized `channelId` slug | channel type, safe target display such as email/domain/phone suffix | raw channel config JSON, webhook secrets/tokens |

The result record should carry safe display fields only. For channel targets and monitor targets, avoid indexing credentials or query-string secrets.

### Ranking and de-dupe

The repository should de-dupe records by `(resourceType, resourceStableKey)` because a query may match multiple prefix records for the same resource. Suggested ranking:

1. Primary label phrase prefix match.
2. Primary label token prefix match.
3. Stable slug/ID token prefix match.
4. Secondary field match.
5. Resource type priority: service, monitor, policy, channel.
6. Alphabetical label tie-break.

### Dashboard top-bar behavior

`AppShell` should grow a top content bar above `main`, containing a search input with a search icon. The search UI should be client-side and call the API through a dashboard helper.

Behavior:

- Debounce input, around 250-300ms.
- Do not call the API until the normalized query has at least 2 characters.
- Show feedback states: idle hint, loading, results, no results, and error.
- Render results as links with resource-specific icons and text.
- Keyboard support should include focusable input/results, arrow navigation if implemented, escape to close, and enter/click to navigate.
- On selection, navigate using `<Link>` or link semantics rather than imperative router calls unless unavoidable.

## Risks / Trade-offs

- **Write amplification from search records** -> Bound tokens and prefixes per resource and keep records compact.
- **Stale search records after updates/deletes** -> Maintain records transactionally where possible with resource writes; delete old search records before/with replacement.
- **DynamoDB transaction item limit** -> Keep indexed fields/prefixes bounded; if transaction limits become tight, use a single denormalized searchable text record plus GSI only in a future design.
- **Sensitive data leakage** -> Exclude channel config, inline channel config, monitor headers, expected body text, and secret URL query strings from indexed and returned fields.
- **No fuzzy matching** -> Make this explicit in copy and rely on prefix/token matching for fast jump behavior.
- **Top bar changes global layout** -> Keep the bar compact and responsive; mobile can collapse to a full-width search row above page content.

## Migration Plan

- New writes create search records for services, monitors, escalation policies, and notification channels.
- Existing resources need backfill. For the first implementation, add a repository helper or admin-safe migration path that reads tenant resource refs and writes search records in bounded batches.
- Until backfill runs, newly created/updated resources may appear in search before older resources.
