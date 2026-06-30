## Context

The dashboard lacks user feedback for:
1. **Creation success**: When a service or monitor is created, user sees a redirect but no confirmation message
2. **Status changes**: When a service goes DOWN or UP, there's no notification - user must notice the status chip change

## Goals / Non-Goals

**Goals:**
- Show success toast when service/monitor is created
- Show status change toast when service goes UP or DOWN
- Show error toast when operations fail
- Non-intrusive - toasts should not block user workflow

**Non-Goals:**
- Notification center or notification history (just transient toasts)
- Email/SMS/push notifications (only in-dashboard toasts)
- Toast for every minor status change (only meaningful transitions)

## Decisions

### Decision 1: Use shadcn/ui Toast

**Choice**: Use shadcn/ui toast component that was already installed.

**Rationale**:
- Already part of the project (shadcn/ui)
- Consistent with existing UI components
- No additional dependencies
- Works well with Next.js App Router

### Decision 2: Toast Provider in Root Layout

**Choice**: `Toaster` component already added to root layout.

```typescript
// app/layout.tsx
import { Toaster } from '@/components/ui/toaster'

export default function RootLayout({ children }) {
  return (
    <html className="dark" lang="en">
      <body>
        {children}
        <Toaster />
      </body>
    </html>
  )
}
```

### Decision 3: Show Toasts from Page Components via Search Params

**Choice**: Show toasts in page components based on `?created`, `?updated`, `?error` search params using a client component.

**Rationale**:
- Server components can't call toast directly (it's a client-side hook)
- Search params already used for success/error messaging in pages
- Clean pattern: page reads param, shows toast, no duplication

```typescript
// components/toast-watcher.tsx
"use client"

import { useEffect } from "react"
import { useSearchParams } from "next/navigation"
import { toast } from "@/hooks/use-toast"

export function ToastWatcher() {
  const searchParams = useSearchParams()

  useEffect(() => {
    if (searchParams.get("created")) {
      toast.success("Created successfully")
    }
    if (searchParams.get("updated")) {
      toast.success("Updated successfully")
    }
    if (searchParams.get("error")) {
      toast.error(searchParams.get("error") || "An error occurred")
    }
  }, [searchParams])

  return null
}
```

### Decision 4: Status Change Toasts from Polling

**Choice**: Track previous status in component state and show toast when status changes during polling.

```typescript
// In monitoring page client component
const [prevStatus, setPrevStatus] = useState(service.rollupStatus)

useEffect(() => {
  if (prevStatus && prevStatus !== service.rollupStatus) {
    if (service.rollupStatus === 'DOWN') {
      toast.error("Service is DOWN!", {
        description: service.name,
      })
    } else if (service.rollupStatus === 'UP' && prevStatus === 'DOWN') {
      toast.success("Service is UP again!", {
        description: service.name,
      })
    }
  }
  setPrevStatus(service.rollupStatus)
}, [service.rollupStatus, prevStatus])
```

## Toast Types

| Event | Toast Type | Message | Duration |
|-------|------------|---------|----------|
| Service created | success | "Created successfully" | 4000ms |
| Service updated | success | "Updated successfully" | 4000ms |
| Monitor created | success | "Monitor created" | 4000ms |
| Monitor updated | success | "Monitor updated" | 4000ms |
| Service goes DOWN | error | "Service is DOWN" | 6000ms |
| Service goes UP (from DOWN) | success | "Service is UP again" | 4000ms |
| Action fails | error | Error message | 6000ms |

## Implementation Plan

1. Add `Toaster` to root layout ✓ (done)
2. Create `ToastWatcher` client component to handle search param toasts
3. Add `ToastWatcher` to monitoring layout
4. Create `StatusToast` client component for status change detection
5. Add status change toasts to services list and detail pages
6. Add toasts to incidents pages (acknowledge, resolve)

## Cost Analysis

- **No additional dependencies** - using existing shadcn toast
- **Runtime cost**: Negligible - only renders when triggered
- **No API costs**: Client-side only