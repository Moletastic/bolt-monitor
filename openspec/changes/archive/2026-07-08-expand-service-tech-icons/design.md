## Context

Service category is a persisted string on services and is validated by `shared/monitorconfig`. The dashboard mirrors the category union in TypeScript and maps each category to an inline SVG through `TechIcon`. Current categories are `server`, `database`, `cache`, `http`, `queue`, `container`, and `function`.

The service creation/edit UI and service cards use these categories as generic technology identity signals. Operators need a broader catalog for common service shapes without introducing brand-specific logos or external icon dependencies.

## Goals / Non-Goals

**Goals:**

- Expand the supported service category catalog with common generic technology/service types.
- Keep backend validation, dashboard TypeScript types, and dashboard icon rendering aligned.
- Preserve all existing category values and persisted data.
- Provide accessible labels and deterministic fallback rendering for unknown or missing categories.
- Use product-owned inline icons consistent with the existing monitoring-console visual language.

**Non-Goals:**

- Add vendor-specific logos such as AWS, Stripe, Postgres, Redis, or Kubernetes marks.
- Add an external icon package.
- Add custom per-service uploaded icons.
- Change monitor protocol icons.
- Migrate existing service records to new categories.

## Decisions

### Expand generic categories end-to-end

Add the following technical category values everywhere categories are validated or typed:

```txt
web
api
worker
scheduler
storage
search
auth
payments
analytics
observability
ai
integration
```

Add the following service-purpose category values everywhere categories are validated or typed:

```txt
media
content
finance
learning
gaming
commerce
messaging
support
marketing
admin
security
location
social
```

This provides enough range for common operator mental models while avoiding a long vendor catalog. The catalog intentionally mixes technical shape categories and product-purpose categories because operators may think about a service by either how it runs or what business function it owns.

### Keep icons inline and generic

`TechIcon` should continue using inline SVG glyphs. This avoids dependency and licensing churn and keeps rendering consistent across cards, detail headers, and form controls.

Purpose-category glyphs should be semantically recognizable without becoming decorative brand art. For example, `media` can use a play/image motif, `content` can use a document/editorial motif, `finance` can use ledger/coin motifs, and `learning` can use a book/cap motif.

### Preserve string storage and no migration

No storage migration is needed because service category is already stored as a string. Existing category values remain supported. Existing services without a category still render the fallback service/server icon.

### Maintain a single dashboard catalog

Dashboard category values, labels, and icons should come from one local catalog or tightly coupled definitions so the service form selector, service card rendering, and detail rendering do not drift.

## Risks / Trade-offs

- **Backend/frontend category drift** -> Add or update tests that compare the Go-supported values against the dashboard TypeScript catalog if practical, or at minimum test both lists when adding categories.
- **Too many choices make creation slower** -> Use clear labels and a stable ordering or grouping that keeps broad/common categories discoverable.
- **Generic icons may be less recognizable than brand logos** -> Favor consistent product-owned symbols and labels over brand specificity.
- **Unknown future categories could break rendering** -> Preserve fallback icon behavior for missing or unrecognized category values.
