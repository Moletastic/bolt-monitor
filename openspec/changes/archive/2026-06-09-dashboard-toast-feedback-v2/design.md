## Context

The first attempt at implementing toast notifications (dashboard-toast-notifications) failed. This document captures the issues and a revised approach.

## Issues from v1 and Fixes

### Issue 1: redirect() inside try-catch
**Problem**: `redirect()` throws `NEXT_REDIRECT` internally to stop execution. When inside a try-catch block, it gets caught and causes unexpected behavior.

**Fix**: Move `redirect()` OUTSIDE the try-catch. Only catch actual API errors:
```typescript
// BROKEN (v1)
try {
  await createService(payload)
  redirect(`/services/${id}?created=1`)  // throws NEXT_REDIRECT, caught!
} catch (error) {
  redirect(`${returnTo}?error=...`)  // conflict!
}

// FIXED (v2)
let service
try {
  service = await createService(payload)
} catch (error) {
  redirect(`${returnTo}?error=...`)  // only catches REAL errors
}
revalidatePath('/services')
redirect(`/services/${service.serviceId}?created=1`)  // success path
```

### Issue 2: useToast hook listener instability
**Problem**: `useEffect(() => {...}, [state])` re-ran on every state change, causing listeners to be constantly added/removed.

**Fix**: Remove dependency array so effect runs only once on mount:
```typescript
useEffect(() => {
  listeners.push(setState)
  return () => {
    const index = listeners.indexOf(setState)
    if (index > -1) listeners.splice(index, 1)
  }
}, [])  // empty dependency - runs once on mount
```

### Issue 3: ToastWatcher not detecting search params
**Problem**: Using `useSearchParams()` hook had issues with Next.js 15 Suspense requirements and reactivity.

**Fix**: Use `window.location.search` directly in useEffect with empty dependency (runs once on mount):
```typescript
useEffect(() => {
  const search = window.location.search
  const params = new URLSearchParams(search)
  // ... check params
}, [])  // empty dependency - runs once on mount
```

## Architecture

### ToastWatcher (Param-based Toasts)
```
User submits form
    ↓
Server action completes
    ↓
redirect() with ?created=1 (or ?error=...)
    ↓
Page navigates, ToastWatcher mounts
    ↓
useEffect reads window.location.search
    ↓
toast({ title: 'Created successfully' })
```

### ServiceListStatusToast (Polling-based Toasts)
```
PollingProvider triggers router.refresh()
    ↓
Page re-renders with fresh data
    ↓
ServiceListStatusToast effect runs
    ↓
Compares current status vs sessionStorage
    ↓
If status changed (DOWN→UP or UP→DOWN):
    toast() with appropriate message
```

## Toast Types

| Event | Toast Type | Message | Duration |
|-------|------------|---------|----------|
| Service created | success | "Created successfully" | 4000ms |
| Service updated | success | "Updated successfully" | 4000ms |
| Monitor created | success | "Created successfully" | 4000ms |
| Monitor updated | success | "Updated successfully" | 4000ms |
| Service goes DOWN | destructive | "Service is DOWN" | 6000ms |
| Service goes UP (from DOWN) | success | "Service is UP again" | 4000ms |
| Action fails | destructive | Error message | 6000ms |
| Manual run triggered | success | "Manual run triggered" | 4000ms |

## Component Structure

```
apps/dashboard/
├── app/
│   └── layout.tsx                    ← Toaster component
├── components/
│   ├── toast-watcher.tsx              ← Param-based toasts
│   └── service-list-status-toast.tsx ← Status change toasts
├── hooks/
│   └── use-toast.ts                  ← Fixed hook
└── lib/
    └── actions.ts                    ← Fixed redirects
```

## Implementation Plan

1. **Fix use-toast.ts**: Remove `state` from useEffect dependency array
2. **Fix actions.ts**: Move redirect() outside try-catch blocks
3. **Fix toast-watcher.tsx**: Use window.location.search directly
4. **Add service-list-status-toast.tsx**: For polling-based status toasts
5. **Test**: Create service/monitor, verify toast appears
6. **Test**: Wait for status change, verify toast appears