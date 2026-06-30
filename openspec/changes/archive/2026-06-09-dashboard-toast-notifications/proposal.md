## Why

Users currently receive no feedback when:
- A service is created successfully
- A monitor is created successfully
- A service status changes (UP/DOWN)

The dashboard shows stale data via polling, but doesn't proactively notify users of important events. Users must manually refresh or watch the dashboard to notice changes.

## What Changes

- **New**: Toast notification system using shadcn/ui toast component
- **New**: Success toasts for service/monitor creation
- **New**: Status change toasts when services go UP or DOWN
- **New**: Error toasts when operations fail

## Capabilities

### New Capabilities
- `dashboard-toast-notifications`: Dashboard shows toast notifications for user actions and status changes

## Impact

- **Code**: `hooks/use-toast.ts` - already exists from shadcn toast
- **Code**: Toast calls added to pages after successful operations
- **Code**: Status change polling triggers toasts when service status changes