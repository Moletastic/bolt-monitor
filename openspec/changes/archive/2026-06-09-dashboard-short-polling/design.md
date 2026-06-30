## Context

The dashboard currently relies on:
- Initial server-side data fetch (static at page load)
- `revalidatePath()` after mutations to refresh cache
- No automatic refresh for real-time monitoring

Users viewing service status, monitor runs, and incidents need to see updates within seconds of new data being available.

## Goals / Non-Goals

**Goals:**
- Provide near-real-time updates (5 second interval)
- Minimize unnecessary API calls
- Use Next.js `router.refresh()` for seamless updates without full page reload
- Apply polling selectively to pages that need live data

**Non-Goals:**
- WebSockets or long-polling (overkill for this use case)
- Polling on every page (only dashboard monitoring pages)
- Polling during inactive tabs (should pause when tab is not visible)

## Decisions

### Decision 1: Use `router.refresh()` with setInterval

**Choice**: Client component uses `setInterval` with `router.refresh()` every 5 seconds.

**Rationale**: 
- `router.refresh()` re-executes the server component and re-fetches data
- No full page reload - seamless user experience
- Native Next.js approach for App Router
- Much simpler than WebSocket or SSE implementation

**Pattern:**
```typescript
'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export function PollingProvider({ intervalMs = 5000 }: { intervalMs?: number }) {
  const router = useRouter()

  useEffect(() => {
    // Don't poll in inactive tabs
    if (document.visibilityState === 'hidden') {
      return
    }

    const id = setInterval(() => {
      router.refresh()
    }, intervalMs)

    return () => clearInterval(id)
  }, [router, intervalMs])

  return null // Renders nothing, just provides polling side effect
}
```

### Decision 2: Visibility API to Pause Polling

**Choice**: Use Page Visibility API to pause polling when tab is not active.

**Rationale**: Saves resources and prevents unnecessary API calls when user is not looking at the dashboard.

```typescript
useEffect(() => {
  const handleVisibility = () => {
    if (document.visibilityState === 'visible') {
      router.refresh() // Immediate refresh when tab becomes visible
    }
  }

  document.addEventListener('visibilitychange', handleVisibility)
  return () => document.removeEventListener('visibilitychange', handleVisibility)
}, [router])
```

### Decision 3: Selective Polling (Not Global)

**Choice**: Only wrap pages that display live monitoring data, not the entire dashboard.

**Rationale**: Polling every page is wasteful. Static pages (settings, create forms) don't need it.

**Pages to poll:**
- `/services` - service status rollups
- `/services/:id` - service detail with monitors
- `/services/:id/monitors/:id` - monitor detail with runs
- `/incidents` - incident list
- `/incidents/:id` - incident detail

**Pages to NOT poll:**
- `/services/new` - static form
- `/services/:id/monitors/new` - static form
- `/admin/scheduler` - config page
- `/audit-trail` - historical data

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    POLLING ARCHITECTURE                          │
└─────────────────────────────────────────────────────────────────┘

  ┌─────────────────────────────────────────────────────────────┐
  │  PollingProvider (Client Component)                          │
  │                                                              │
  │  useEffect: setInterval(5000ms)                             │
  │      │                                                       │
  │      ├──► router.refresh()                                  │
  │      │       │                                               │
  │      │       ▼                                               │
  │      │  Server Component re-executes                          │
  │      │       │                                               │
  │      │       ▼                                               │
  │      │  API functions called (getMonitor, listServices...)   │
  │      │       │                                               │
  │      │       ▼                                               │
  │      │  Fresh data rendered                                  │
  │      │                                                       │
  │  visibilitychange listener:                                  │
  │      │                                                       │
  │      └──► Pause when hidden, refresh when visible            │
  └─────────────────────────────────────────────────────────────┘
```

## Component Structure

```
apps/dashboard/
├── components/
│   └── polling-provider.tsx  ← NEW
├── app/
│   ├── layout.tsx             ← Wrap children with provider
│   ├── services/
│   │   ├── page.tsx          ← Gets polled (via layout or direct)
│   │   └── [serviceId]/
│   │       └── monitors/
│   │           └── [monitorId]/page.tsx ← Gets polled
│   └── incidents/
│       ├── page.tsx          ← Gets polled
│       └── [id]/page.tsx     ← Gets polled
```

## Cost Analysis

**Per user per day:**
- 5s interval = 17,280 refreshes/day
- 10s interval = 8,640 refreshes/day

**10 users, 8-hour day, 10s interval:**
- ~28,800 API calls/day
- ~864,000/month
- Cost: ~$0.17/month (well within free tier)

**DynamoDB reads per refresh:**
- ListServices: ~5 read units
- GetMonitorStatus: ~1 read unit

**Even with 100 users, costs are negligible.**

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Too many API calls | 5s is reasonable; visibility API pauses when tab hidden |
| Server load | Lambda + DynamoDB scale automatically; negligible cost |
| Stale data during heavy load | 5s max staleness is acceptable for monitoring |
| Polling while form is open | Forms are on separate pages; no conflict |

## Open Questions

1. **Should we add a visual indicator** (like "last updated X seconds ago")?
2. **Should users be able to disable polling** in settings?
3. **Should we debounce rapid successive navigations** to avoid duplicate refreshes?