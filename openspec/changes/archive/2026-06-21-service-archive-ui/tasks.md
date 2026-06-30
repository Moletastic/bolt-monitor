# Tasks: Service Archive UI

## 1. Add Archive Action to Service Detail Page

- [x] 1.1 Add archive button to service detail page
- [x] 1.2 Show confirmation dialog before archiving
- [x] 1.3 Call server action to set `lifecycleState: "archived"`

## 2. Handle Archived Service State in UI

- [x] 2.1 Show "Archived" badge on service detail
- [x] 2.2 Disable/hide edit form for archived services
- [x] 2.3 Disable "Create monitor" button for archived services
- [x] 2.4 Show read-only indicator on monitors list

## 3. Test

- [x] 3.1 Archive an active service via UI
- [x] 3.2 Verify service shows "Archived" status
- [x] 3.3 Verify edit form is disabled
- [x] 3.4 Verify cannot enable/disable monitors on archived service
- [x] 3.5 Verify can navigate back to services list and see archived status