## Context

The monitor API already supports deleting escalation policies through `DELETE /api/v1/escalation-policies/{policyId}` and rejects policies that are still referenced by services. The dashboard already has `deleteEscalationPolicyAction`, but the edit notification route page only renders the edit form.

## Goals / Non-Goals

**Goals:**
- Expose route deletion from the notification route detail page.
- Preserve the existing backend reference guard so in-use routes cannot be deleted.
- Use the existing dashboard confirmation pattern for destructive actions.

**Non-Goals:**
- Add bulk deletion.
- Add a new API endpoint or change backend deletion semantics.
- Add service reassignment or detach flows for referenced routes.

## Decisions

- Reuse the existing delete server action and API helper instead of adding a page-specific delete action.
  - Rationale: deletion semantics already live at the API boundary and include the reference guard.
  - Alternative considered: create a new route-specific action wrapper; rejected because it adds indirection without new behavior.

- Render a destructive confirmation form on the edit page rather than the list page.
  - Rationale: the detail page gives operators context before destructive deletion and is the page already mentioned by the current gap.
  - Alternative considered: add delete buttons to every list row; rejected as a larger UX change.

- Redirect successful deletes to `/policies?deleted=1` and keep failures on the route detail page through the existing `returnTo` field.
  - Rationale: successful deletion invalidates the detail page, while failures need inline context.

## Risks / Trade-offs

- Referenced policies still show a delete control even though deletion will fail -> The backend remains authoritative and the page displays the typed error message.
- The route detail page gains another destructive control -> Use the existing confirmation dialog pattern and clear copy to reduce accidental deletion.
