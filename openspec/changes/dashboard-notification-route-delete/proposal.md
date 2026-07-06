## Why

Operators can create and edit notification routes, but the dashboard route detail page does not expose the existing delete capability. This leaves stale or misconfigured routes visible unless deletion is performed outside the UI.

## What Changes

- Add a destructive delete control to the notification route edit page.
- Reuse the existing dashboard server action and monitor API `DELETE /api/v1/escalation-policies/{policyId}` endpoint.
- Surface successful deletes on the routes list and referenced-route failures on the route detail page.

## Capabilities

### New Capabilities

### Modified Capabilities
- `escalation-policy-crud`: Dashboard users can delete unreferenced notification routes from the route detail page and receive inline feedback when deletion is blocked.

## Impact

- Affected dashboard route: `apps/dashboard/app/(monitoring)/policies/[policyId]/page.tsx`
- Affected dashboard server action: existing `deleteEscalationPolicyAction`
- Affected tests: dashboard action or page tests covering delete form wiring and feedback
- No API, data model, dependency, or infrastructure changes are required.
