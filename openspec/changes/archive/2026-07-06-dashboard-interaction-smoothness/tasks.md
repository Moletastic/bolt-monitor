## 1. Audit And Baseline

- [x] 1.1 Inventory all dashboard mutation feedback paths, including query parameters, inline banners, toasts, and action-state messages.
- [x] 1.2 Identify which mutation events currently produce duplicate feedback and record the intended single feedback owner for each event.
- [x] 1.3 Browser-test the current polling refresh behavior to confirm whether interval refresh or visibility refresh causes visible interruption.
- [x] 1.4 Browser-test current detail-route loading behavior and list only routes whose placeholders are missing, mismatched, or layout unstable.
- [x] 1.5 Verify current destructive-delete focus behavior for services, monitors, policies, and notification channels.

## 2. Polling Smoothness

- [x] 2.1 Schedule polling-provider interval refreshes as non-urgent React work without adding client router usage outside the provider.
- [x] 2.2 Prevent a single visibility transition from issuing redundant refresh bursts or creating overlapping polling intervals.
- [x] 2.3 Add or update tests for polling interval setup, visibility handling, and cleanup behavior.

## 3. Feedback Surface Consistency

- [x] 3.1 Implement the feedback ownership map so each mutation event renders exactly one visible feedback surface.
- [x] 3.2 Preserve toast feedback only for events assigned to toast ownership and remove duplicate toast/banner cases.
- [x] 3.3 Ensure error feedback uses typed dashboard error messages and an accessible alert or equivalent announced surface.
- [x] 3.4 Add tests or guard coverage for duplicate-feedback prevention on routes with query-parameter feedback.

## 4. Same-Page Mutation Continuity

- [x] 4.1 Select one same-page mutation flow as the reference implementation without using `router.push` or `router.replace`.
- [x] 4.2 Add visible pending, success, and error feedback for the reference same-page mutation flow.
- [x] 4.3 Reconcile the completed state with server-rendered data while preserving required cache invalidation.
- [x] 4.4 Apply the reference pattern to monitor enable/disable where the operator remains on the same route.
- [x] 4.5 Apply the reference pattern to incident acknowledge/resolve where the operator remains on the same route.
- [x] 4.6 Add success and failure coverage for converted same-page mutation flows.

## 5. Interactive Semantics And Focus

- [x] 5.1 Refactor mobile monitor cards so navigation links and inline mutation forms are sibling interactive elements, not nested controls.
- [x] 5.2 Verify keyboard tab order and touch behavior for the refactored mobile monitor card.
- [x] 5.3 Fix any destructive-delete flow where focus falls back to `<body>` or an unrelated target after successful navigation.
- [x] 5.4 Add accessibility coverage for nested interactive prevention and post-delete focus behavior where practical.

## 6. Loading Continuity

- [x] 6.1 Add or adjust route-specific loading placeholders only for verified gaps from task 1.4.
- [x] 6.2 Ensure changed loading placeholders match the destination page shape and use existing skeleton styling tokens.
- [x] 6.3 Verify loading placeholders respect reduced-motion requirements and avoid introducing unnecessary skeleton noise.

## 7. Verification

- [x] 7.1 Run `make lint-dashboard`.
- [x] 7.2 Run `make check-dashboard`.
- [x] 7.3 Run `make test-dashboard`.
- [x] 7.4 Browser-test the high-frequency dashboard flows touched by this change on desktop and mobile widths.
- [x] 7.5 Confirm no new `useRouter`, `router.push`, or `router.replace` usage exists outside the polling provider.
