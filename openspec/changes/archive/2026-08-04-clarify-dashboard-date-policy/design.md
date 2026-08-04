## Context

Native `Date` is the value type accepted by `date-fns`, but the current policy and lint rule reject every construction. This obscures the actual risks: unsafe string parsing, ad-hoc clock reads, and manual arithmetic.

## Goals / Non-Goals

**Goals:**
- Permit `new Date(epochMilliseconds)`.
- Preserve `now()` for current wall-clock reads.
- Keep date-fns as parsing, formatting, comparison, and arithmetic utility.

**Non-Goals:**
- Replace date-fns or alter timestamp APIs.

## Decisions

- ESLint permits `new Date(...)` but continues rejecting direct `Date()` calls and direct `Date` methods.
- `new Date(string)` remains a convention violation documented through `parseISO`; runtime lint cannot safely distinguish all string values.

## Risks / Trade-offs

- Permitting constructor calls can allow unsafe string parsing → Document `parseISO` requirement and retain code review/test coverage.
