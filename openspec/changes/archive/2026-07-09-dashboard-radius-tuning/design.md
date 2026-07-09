## Context

Current Tailwind radius tokens are:

```txt
sm      0.125rem
DEFAULT 0.25rem
md      0.375rem
lg      0.5rem
xl      0.75rem
```

Many dashboard surfaces use `rounded-lg` and cards use `rounded-xl`, making operational panels feel more consumer-app than monitoring console.

## Goals / Non-Goals

**Goals:**

- Make dashboard surfaces less curved overall.
- Keep a consistent radius scale through design tokens.
- Preserve fully rounded shapes where shape communicates status or affordance.

**Non-Goals:**

- Remove all rounding.
- Change color, spacing, typography, or shadows.
- Redesign component layouts.

## Decisions

Recommended radius scale:

```txt
sm      0.0625rem
DEFAULT 0.125rem
md      0.25rem
lg      0.375rem
xl      0.5rem
```

Component guidance:

- Cards/panels: `rounded-lg` or `rounded-xl` after token reduction.
- Inputs/buttons/selects/menus: `rounded-md` after token reduction.
- Status chips: keep `rounded-full`.
- Traffic lights/dots: keep `rounded-full`.
- Progress bars: keep `rounded-full` where pill shape improves readability.
- Icon tiles: reduce token if using `rounded-xl`; preserve square-ish tile feel.
- Floating create button: keep circular shape.

Options considered:

| Option | Pros | Cons |
|---|---|---|
| Reduce radius tokens | Broad consistency, minimal class churn | Affects many surfaces at once |
| Replace classes manually | Precise control | High churn, easy drift |
| Remove rounding almost entirely | Very technical look | Harsh, less approachable |

Recommendation: reduce tokens first, then clean up obvious outliers.

## Risks / Trade-offs

- **Global token shift changes many surfaces** -> Verify key pages visually after implementation.
- **Pills become too squared if changed manually** -> Preserve `rounded-full` semantic shapes.
- **Existing `rounded-xl` still too soft in some places** -> Token reduction should handle most cases; only adjust class names if still visibly too curved.
