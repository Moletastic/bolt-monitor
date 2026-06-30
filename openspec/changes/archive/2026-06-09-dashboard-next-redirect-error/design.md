## Context

`NEXT_REDIRECT` errors in Next.js 15 App Router occur when `redirect()` is used incorrectly. In Next.js 15, `redirect()` from `next/navigation` throws a special `NEXT_REDIRECT` error internally - this is expected behavior. The error only becomes problematic when:

1. It's wrapped in a try-catch block that catches it as a generic error
2. It's used in a context where the throw isn't properly propagated
3. It's called outside of a server action or route handler where it can't be properly handled

## Goals / Non-Goals

**Goals:**
- Fix all form submissions to use `redirect()` correctly
- Ensure redirects work as proper HTTP redirects (not errors)
- Verify all affected paths: service create/update, monitor create/update, enable/disable

**Non-Goals:**
- Change the redirect destinations (keep existing UX flow)
- Add new redirect paths

## Decisions

### Decision 1: Use `redirect()` from `next/navigation` Correctly

**Pattern for Server Actions:**
```typescript
// CORRECT - redirect is a return value, not a throw
'use server'
import { redirect } from 'next/navigation'

export async function createService(formData: FormData) {
  // ... validation and API call ...

  redirect(`/services/${serviceId}`)
}

// INCORRECT - wrapping redirect in try-catch
export async function createService(formData: FormData) {
  try {
    // ... logic ...
    redirect(`/services/${serviceId}`)
  } catch (error) {
    // This catches NEXT_REDIRECT and treats it as an error!
    throw error
  }
}
```

### Decision 2: Don't Catch Redirect Errors

`redirect()` throws `NEXT_REDIRECT` internally - this is intentional. Server actions should NOT catch this error type.

**Correct Pattern:**
```typescript
export async function submitForm(formData: FormData) {
  const result = await apiCall()

  if (!result.success) {
    return { error: result.error }
  }

  redirect(result.redirectUrl) // This throws - that's fine
}
```

### Decision 3: Use `revalidatePath()` When Data Changes

After mutations that affect list pages, use `revalidatePath()` to refresh cached data without full page reload.

```typescript
import { redirect, revalidatePath } from 'next/navigation'

export async function createMonitor(serviceId: string, formData: FormData) {
  const monitor = await createMonitorAPI(serviceId, formData)

  revalidatePath(`/services/${serviceId}`)
  revalidatePath('/services')

  redirect(`/services/${serviceId}/monitors/${monitor.monitorId}`)
}
```

## Affected Areas

| Area | File Pattern | Issue |
|------|--------------|-------|
| Service forms | `apps/dashboard/app/services/**/` | May catch redirect errors |
| Monitor forms | `apps/dashboard/app/**/monitors/**/` | May catch redirect errors |
| Server actions | `apps/dashboard/actions/**/` | May use incorrect redirect pattern |

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Missing some redirect paths | Thorough testing of all form submissions |
| Breaking existing redirects | Verify each redirect destination remains correct |

## Open Questions

1. **Should we add error boundary components** to gracefully handle any remaining edge cases?

2. **Should we log redirect events** for analytics on user navigation patterns?
