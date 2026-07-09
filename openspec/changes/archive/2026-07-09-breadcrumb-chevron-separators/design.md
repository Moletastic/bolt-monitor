## Context

Dashboard breadcrumbs already provide parent navigation. Separator choice is visual polish, but it affects scannability and accessibility noise.

## Goals / Non-Goals

**Goals:**

- Use a chevron separator between breadcrumb items.
- Keep separators visually muted.
- Prevent screen readers from announcing decorative separators.

**Non-Goals:**

- Change breadcrumb routes, labels, or visibility rules.
- Redesign top bar or page headers.

## Decisions

Preferred separator: right chevron icon.

Options:

| Option | Pros | Cons |
|---|---|---|
| Chevron icon | Clear hierarchy, compact, common app pattern | Needs `aria-hidden` |
| Slash `/` | Simple, text-only | Feels like file path |
| Dot `·` | Minimal | Weak hierarchy cue |

Use chevron icon, muted color, `aria-hidden="true"`.

## Risks / Trade-offs

- **Icon adds visual noise** -> Keep small and muted.
- **Screen reader noise** -> Mark separator `aria-hidden`.
