## Context

The Services landing page currently renders an `sr-only` page title, a large “Service operations” card, uneven summary cards, and the service card grid. The operations card repeats service identity/status information at lower density than the card list below it. The desired refactor keeps the page focused on scanning and acting on services: health indicators, list controls, and service cards.

The dashboard uses Next App Router server pages and the repo convention prefers `Link` navigation for route changes. The page already fetches all services server-side with `listServices()`, so list search and filters can be applied in the dashboard without requiring a backend API change.

## Goals / Non-Goals

**Goals:**

- Make the Services landing page hierarchy clearer by removing duplicated service preview content from the top card.
- Give `Active`, `Drafts`, and `Down now` equal visual weight.
- Add a dedicated controls row above the service cards for search and filters.
- Keep service creation prominent on desktop and reachable on mobile through a floating action button.
- Preserve existing empty, unavailable, deletion-feedback, and service-card behaviors.
- Keep loading skeletons aligned with the final layout to avoid visual mismatch.

**Non-Goals:**

- Change service, monitor, or status API behavior.
- Add API-backed search.
- Make summary indicators clickable filters in this change.
- Redesign service overview cards themselves.
- Introduce a shared page-header component across other dashboard modules.

## Decisions

### Use a two-band desktop layout

Desktop should show a top identity/health band followed by a list-controls band:

```txt
Services                                  [ Active ] [ Drafts ] [ Down now ]
Track service health, lifecycle, and
monitor coverage from one place.

[ Search services...                         ] [Filter] [Create service]

[ Service card ] [ Service card ] [ Service card ] [ Service card ]
```

This separates page meaning from list manipulation. The old operations card combined page identity, preview content, and navigation shortcuts in one area, which made the page feel busier without adding distinct information.

### Use a compact mobile-first service-list layout

On narrow screens, visible title/description should be removed from the content flow. The top visible content should be the three health indicators in one row, followed by controls, then one service card per row:

```txt
[ Active ] [ Drafts ] [ Down now ]

[ Search services... ]
[ Filter ]

[ Service card ]
[ Service card ]

                         [ + ]
```

The page still needs a semantic accessible heading so screen reader users have page context even when the visual mobile layout omits the desktop title and description.

### Keep create service responsive by placement

Desktop creation belongs in the controls row because the operator is already deciding how to act on the list. Mobile creation should use a fixed bottom-right floating action button because horizontal controls are constrained and creation should remain reachable while scanning a long list.

The floating button should expose an accessible name such as `Create service`, respect safe-area spacing, and avoid covering primary service card content.

### Scope search and filters to dashboard behavior

Search should filter the already-loaded service collection by human-readable service attributes. The first implementation can keep filtering local to the Services page while preserving room for URL-backed filters later. Filter affordances should be present as part of the controls row, but advanced filter behavior can remain minimal unless explicitly covered by follow-up requirements.

## Risks / Trade-offs

- **Mobile stat row may become cramped** -> Use compact labels and equal flexible cards; prefer `Down now` over longer wording in constrained layouts.
- **Floating create button may obscure content** -> Add bottom padding to the list area and position the action with safe-area-aware spacing.
- **Search/filter scope may feel ambiguous** -> Keep placeholder/help text focused on services and filter only visible service-list data in this change.
- **Desktop/mobile divergence may hurt accessibility** -> Maintain semantic headings and accessible names even when visual content differs by breakpoint.
- **Loading skeleton mismatch could regress perceived polish** -> Update `loading.tsx` in the same change to mirror header, indicators, controls, and card-list layout.
