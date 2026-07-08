## Why

The current service creation page uses a split layout with a “Create flow notes” card that consumes space without helping operators complete the form. The creation flow should use the full page width to make service identity and notification setup clearer while keeping monitor setup out of scope until the service exists.

## What Changes

- Keep `/services/new` as the canonical dedicated service creation route.
- Remove the “Create flow notes” side card from the service creation page.
- Use the full page width for the service creation form.
- Add a page title and description section above the form content.
- Organize the form into two named sections with icons: `Service identity` and `Notifications`.
- Replace the text-only service category selector with an icon-based service icon/category selector.
- Group service icon selection and service name in the same identity row, with description below.
- Keep notification route selection in the Notifications section.
- Place the business-hours switch below notification route selection with explanatory copy and preserve the existing business-hours details when enabled.
- Omit health monitoring from the service creation page for this change.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Update service creation requirements for the dedicated create page layout, icon-based service identity selection, and notification/business-hours organization.

## Impact

- Affected route: `apps/dashboard/app/services/new/page.tsx`.
- Affected form component: `apps/dashboard/components/service-form.tsx`.
- Potential supporting component impact for an icon-based service category selector.
- No backend API changes are expected.
- No create-with-monitors behavior is introduced.
