## Action Consumer Inventory

| Consumer | Actions | Classification | Rationale |
| --- | --- | --- | --- |
| `components/escalation-policy-form.tsx` | create/update escalation policy | navigation-first | Success returns to policies; validation errors currently route back to edit/new page. |
| `components/monitor-table.tsx` | toggle monitor | mixed | Row toggle stays in context but currently refreshes route after mutation. |
| `app/(monitoring)/services/[serviceId]/page.tsx` | delete service | navigation-first | Destructive delete leaves service detail and returns to service list. |
| `components/archive-service-button.tsx` | archive service | navigation-first | Success changes service lifecycle and redirects to detail with status flag. |
| `components/service-form.tsx` | create/update service | navigation-first | Success navigates to service detail. |
| `app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx` | trigger run, toggle monitor, toggle maintenance, delete monitor | mixed | Inline controls stay on detail; delete navigates away. |
| `app/admin/scheduler/page.tsx` | update scheduler config | inline-feedback | Same-page admin setting; no route transition needed. Converted as reference flow. |
| `components/monitor-form.tsx` | create/update monitor | navigation-first | Success navigates to monitor detail. |
| `components/notification-channel-form.tsx` | create/update channel | navigation-first | Success navigates to channel list/detail. |
| `app/(monitoring)/integrations/channels/[channelId]/page.tsx` | delete channel | navigation-first | Destructive delete leaves channel detail and returns to list. |
| `app/(monitoring)/incidents/[id]/page.tsx` | acknowledge/resolve incident | mixed | Detail actions stay in context but currently refresh route after mutation. |

## Action-State Contract

Converted flows use `ActionState<T>` from `apps/dashboard/lib/action-state.ts`:

- `idle` before submission.
- `success` with optional message and serializable data.
- `error` with `code: ApiErrorCode`, `details: Record<string, unknown>`, and optional `message`.

Server actions serialize `ApiError` through `actionErr(error)` so client components never receive class instances. UI copy calls `actionErrorMessage(error)`, which displays `error.message` when present and falls back to `humanize(error.code)`.

## Redirect vs Inline Guidance

Keep redirect-based server actions when success changes resource context, leaves the current route, or must preserve the router convention from `AGENTS.md`. Convert to returned action state when the form is same-page, needs inline success/error feedback, and the component can branch on `status` without query-string error transport.

## Converted Flow UX

Scheduler config no longer redirects to `/admin/scheduler?updated=1` on success. It now revalidates `/admin/scheduler` and renders inline success/error state via `useActionState`, preserving same-page behavior while removing query-string feedback transport.
