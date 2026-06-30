## Why

The `ts-result-and-no-any` change introduced the dashboard `Result<T, E>` utility, typed `ApiError`, strict no-`any` linting, and an I/O boundary that converts thrown API failures into `Result<T, ApiError>`. Existing dashboard mutation flows still use navigation-first server actions (`<form action={...}>`) that redirect on success or error. That pattern preserves the current router convention, but it means exported server actions do not directly return `Result<T, ApiError>` to UI consumers.

This follow-on change converts selected mutation flows to UI-state-driven server actions so components can branch on `isOk` / `isErr`, render typed `ApiError` details, and avoid redirect query-string error transport where inline feedback is better.

## What Changes

- Introduce a consistent action-state return shape for dashboard mutation forms.
- Convert server actions that should render inline feedback to return `Result<T, ApiError>` or a serializable action state derived from it.
- Update consuming components to use `useActionState` where navigation is not the primary outcome.
- Preserve navigation-first redirects for flows where success must change route, unless the spec explicitly replaces them.
- Update toast/alert components to render `error.message` when present, else `humanize(error.code)`.

## Capabilities

### Modified Capabilities

- `dashboard-web-app`: selected mutation forms consume returned action state instead of relying exclusively on redirect query parameters for errors.

## Impact

- Modified `apps/dashboard/lib/actions.ts` for returned action-state helpers.
- Modified form components that opt into inline action-state handling.
- Modified toast/alert/error presentation components to consume typed `ApiError` state.
- Additional tests covering successful and failed action-state branches.

## Out of Scope

- The foundational `Result<T, E>` utility, `ApiErrorCode`, and no-`any` lint rules. Those are owned by `ts-result-and-no-any`.
- Replacing every existing navigation-first flow in one pass. Each conversion must preserve the route behavior documented in `AGENTS.md` or explicitly justify the UX change.
