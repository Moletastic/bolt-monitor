## Why

The dashboard policy says native `Date` is banned, which incorrectly implies `date-fns` replaces date objects. Clarify safe construction while retaining deterministic current-time and date-fns manipulation rules.

## What Changes

- Allow native `Date` construction from known epoch values.
- Keep `now()` as the dashboard current-time boundary.
- Require `parseISO` for external ISO strings and date-fns for manipulation, comparison, formatting, and duration calculations.
- Update lint enforcement and contributor guidance.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `code-patterns-foundation`: Dashboard date handling distinguishes date construction from parsing, current-time reads, and manipulation.

## Impact

- Affected dashboard ESLint configuration/tests, `AGENTS.md`, `CONSTITUTION.md`, and date policy spec.
- No API or runtime behavior change.
