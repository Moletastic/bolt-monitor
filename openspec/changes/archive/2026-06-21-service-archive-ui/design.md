# Service Archive UI

## Summary

Add UI to archive services from the dashboard.

## Goals

- Allow users to manually archive active services from the UI
- Provide clear visual indication that archived services are read-only
- Prevent editing of archived services

## Requirements

### Archive Action

- Users can trigger archive from service detail page
- Archive action calls `updateServiceAction` with `lifecycleState: "archived"`
- After archiving, service is displayed as read-only
- Confirmation dialog before archiving

### Archived Service Display

- Archived services show "Archived" badge/label
- Edit form is disabled/hidden for archived services
- Monitor list is still visible but disabled
- No enable/disable actions available on monitors

## Technical Approach

- Add archive button to service detail page (near the "Create monitor" button or in a menu)
- Create `archiveServiceAction` server action or reuse `updateServiceAction`
- Add UI state to show service is archived and read-only
- Prevent navigation to monitor creation for archived services