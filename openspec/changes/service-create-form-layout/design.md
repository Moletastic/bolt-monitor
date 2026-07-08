## Context

The current service creation route renders `ServiceForm` beside a “Create flow notes” card. The side card repeats implementation context instead of helping operators complete the task, and it prevents the form from using the full page width. The form already handles creation and update modes, notification route selection, business-hours state, and server-action error display.

This change keeps `/services/new` as the dedicated creation surface. It does not add monitor creation to the service form because the existing API creates services and monitors through separate endpoints, and monitors require an existing service ID.

## Goals / Non-Goals

**Goals:**

- Make service creation feel like a focused setup page instead of a card plus notes layout.
- Use full-width form content on the create page.
- Group fields by operator intent: service identity first, notification behavior second.
- Show actual service-category icons during selection instead of only category names.
- Keep the business-hours control understandable as an optional part of notification behavior.
- Preserve the existing service creation behavior and redirect to service detail after success.

**Non-Goals:**

- Add initial monitor creation to the service creation form.
- Add a create-service-with-monitors API endpoint.
- Redesign monitor creation or edit flows.
- Change service update behavior beyond safely sharing reusable form section structure where appropriate.
- Introduce modal or drawer creation.

## Decisions

### Keep service creation on a dedicated page

The current form contains enough setup complexity that a modal or drawer would make validation recovery and mobile usability worse. A dedicated route also preserves refresh/back behavior and works naturally with server actions.

### Remove side notes and use full width

The create page should remove the notes card and give the form the full content area:

```txt
Create service
Define the identity and alert routing for a new service.

[ Service identity ]
[ Notifications    ]

                                      [Create service]
```

The note that monitors are created after the service exists can be omitted from the primary form until a future monitor-onboarding change.

### Split form content into two icon-labeled sections

Except for the page title/description, each major form section should have a section icon and section name:

```txt
[service icon] Service identity
  [icon/category selector] [service name]
  [description]

[bell icon] Notifications
  [notification route]
  [business hours switch]
  [business hours details when enabled]
```

This makes the create flow easier to scan and leaves room for future sections without reintroducing a side rail.

### Represent service category through icons

The persisted field remains `serviceCategory`; the UI should present it as an icon/category choice. The selected value still submits through the existing `serviceCategory` field so no API contract changes are needed.

### Keep health monitoring out of this change

The service API currently creates a service first, then monitors under `/api/v1/services/{serviceId}/monitors`. Adding monitor drafts to the service form would require dashboard orchestration or a new backend endpoint. This change intentionally avoids that complexity and leaves monitor onboarding for a later, explicit change.

## Risks / Trade-offs

- **Icon selector may become visually noisy** -> Keep labels visible with icons so category names remain understandable and accessible.
- **Shared create/edit form may make create-only layout awkward** -> Allow conditional rendering for create-specific page framing while preserving update behavior.
- **Removing notes may hide useful context** -> Keep any essential guidance inline near the relevant control instead of in a separate notes card.
- **Business-hours copy may imply more routing control than exists** -> Write copy that matches current payload behavior and avoids unsupported promises.
