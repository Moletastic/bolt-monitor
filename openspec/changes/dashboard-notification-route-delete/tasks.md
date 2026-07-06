## 1. Dashboard UI

- [x] 1.1 Add a destructive delete section to the notification route edit page that posts `policyId` and `returnTo` to `deleteEscalationPolicyAction`.
- [x] 1.2 Add routes-list feedback for `deleted=1` so successful deletions show a confirmation message.
- [x] 1.3 Ensure delete failures return to the route edit page and display the typed error message.

## 2. Verification

- [x] 2.1 Add or update dashboard tests covering delete form wiring and delete success feedback.
- [x] 2.2 Run dashboard lint, type check, and tests relevant to the change.
