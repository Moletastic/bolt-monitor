## Why

The dashboard currently shows stale data. Server components only re-fetch when the user navigates away or submits an action. Users monitoring services and incidents need near-real-time updates (within 5 seconds) without manually refreshing the page.

## What Changes

- **New**: `PollingProvider` client component that refreshes server component data every 5 seconds
- **New**: Wrap dashboard pages that display live status data with polling provider
- **New**: Polling is selective - only pages that need real-time data are polled

## Capabilities

### New Capabilities
- `dashboard-short-polling`: Dashboard pages automatically refresh data every 5 seconds using `router.refresh()` without full page reload

## Impact

- **Code**: `apps/dashboard/components/polling-provider.tsx` - new client component
- **Code**: Dashboard layout and pages wrapped with polling provider