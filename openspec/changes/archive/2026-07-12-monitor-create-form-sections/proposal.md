## Why

The current monitor create/edit form is a flat grid and the create page still shows a “Create flow notes” side card. Operators need a clearer flow that separates monitor identity, protocol selection, request configuration, and validation expectations while keeping unsupported protocol work visibly deferred.

## What Changes

- Remove the “Create flow notes” side card from the monitor creation page.
- Use the full page width for monitor creation content.
- Split the monitor form into three sections: `Identity`, `Request`, and `Validation`.
- Place monitor name and check frequency in the `Identity` section.
- Add protocol tabs between `Identity` and `Request` with `HTTP` preselected and `TCP` / `gRPC` disabled with coming-soon affordance.
- Keep HTTP as the only supported submitted monitor type for this phase.
- Update the HTTP `Request` section to include method, target URL, timeout in milliseconds, and editable key/value header rows.
- Default headers to `Content-Type: application/json`.
- Replace the raw multiline headers textarea with key/value rows, delete controls, and a full-width add-header button.
- Update the `Validation` section to show expected status code selection as removable badge-style tags using common status code options, defaulting to `200`.
- Hide the existing expected-body-contains field and treat it as coming soon for this phase.
- Apply the redesigned form to both create and edit modes.
- Remove the monitor edit form from the monitor detail view for now; that edit surface will be refactored separately later.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Update monitor create/edit form layout, HTTP request configuration controls, validation controls, and monitor detail edit-surface requirements.

## Impact

- Affected route: `apps/dashboard/app/services/[serviceId]/monitors/new/page.tsx`.
- Affected route: `apps/dashboard/app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx`.
- Affected component: `apps/dashboard/components/monitor-form.tsx`.
- Potential supporting component impact for client-side header rows, protocol tabs, and status-code tag selection.
- No backend API changes are expected.
- TCP and gRPC creation remain out of scope.
