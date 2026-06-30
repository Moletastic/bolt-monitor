## Why

Users currently receive no feedback when:
- A service is created successfully
- A monitor is created successfully
- A service status changes (UP/DOWN)

The dashboard shows stale data via polling, but doesn't proactively notify users of important events. Users must manually refresh or watch the dashboard to notice changes.

## What Changes

- **New**: Toast notification system using shadcn/ui toast component
- **New**: Success toasts for service/monitor creation via URL search params
- **New**: Status change toasts when services go UP or DOWN via polling
- **New**: Error toasts when operations fail

## Known Issues from v1

The first attempt at dashboard-toast-notifications failed due to:
1. `redirect()` inside `try-catch` blocks causing NEXT_REDIRECT errors
2. `useToast` hook having `useEffect(() => {...}, [state])` causing listener instability
3. `ToastWatcher` component not properly detecting search params on navigation

## Capabilities

### New Capabilities
- `dashboard-toast-feedback-v2`: Dashboard shows toast notifications for user actions and status changes

## Impact

- **Code**: `hooks/use-toast.ts` - fix dependency array bug
- **Code**: `lib/actions.ts` - move redirect() outside try-catch
- **Code**: `components/toast-watcher.tsx` - use window.location.search directly
- **Code**: `components/service-list-status-toast.tsx` - status change detection
- **Code**: `app/layout.tsx` - add Toaster component