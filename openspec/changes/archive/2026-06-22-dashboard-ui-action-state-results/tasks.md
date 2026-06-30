## 1. Design the action-state contract

- [x] 1.1 Inventory the 11 current `apps/dashboard/lib/actions.ts` consumers and classify each as navigation-first, inline-feedback, or mixed.
- [x] 1.2 Define a serializable action-state type that preserves `ApiErrorCode`, `details`, and `message` without passing class instances to client components.
- [x] 1.3 Document when to keep redirect-based server actions versus when to convert to returned action state.

## 2. Convert selected flows

- [x] 2.1 Convert one low-risk form flow to returned action state and `useActionState` as the reference implementation.
- [x] 2.2 Update the consuming component to branch on success/error state and render typed error feedback.
- [x] 2.3 Preserve route navigation behavior for success paths or document the UX change in this change.

## 3. Error presentation

- [x] 3.1 Update toast/alert helpers used by converted flows to render `error.message` when present, else humanize `error.code`.
- [x] 3.2 Ensure raw `ApiErrorCode` enum strings are not the primary user-facing copy in converted flows.

## 4. Tests and verification

- [x] 4.1 Add tests for the action-state serializer/deserializer or type guards.
- [x] 4.2 Add coverage for converted component success and error branches.
- [x] 4.3 Run `make check-dashboard`, `make lint-dashboard`, `make test-dashboard`, and `make build-dashboard`.
