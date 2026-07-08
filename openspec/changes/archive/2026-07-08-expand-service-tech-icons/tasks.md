## 1. Shared Category Catalog

- [x] 1.1 Add expanded service category constants to `shared/monitorconfig`.
- [x] 1.2 Include expanded categories in backend service category validation.
- [x] 1.3 Add or update Go tests for accepted expanded categories and rejected unsupported categories.

## 2. Dashboard Category Catalog

- [x] 2.1 Update dashboard `ServiceCategory` TypeScript union and `SERVICE_CATEGORIES` list with every backend-supported category.
- [x] 2.2 Add dashboard labels for expanded categories with stable operator-readable wording.
- [x] 2.3 Ensure dashboard category definitions remain aligned with backend-supported values.

## 3. Technology Icon Rendering

- [x] 3.1 Add inline generic SVG glyphs for technical categories: `web`, `api`, `worker`, `scheduler`, `storage`, `search`, `auth`, `payments`, `analytics`, `observability`, `ai`, and `integration`.
- [x] 3.2 Add inline generic SVG glyphs for purpose categories: `media`, `content`, `finance`, `learning`, `gaming`, `commerce`, `messaging`, `support`, `marketing`, `admin`, `security`, `location`, and `social`.
- [x] 3.3 Map every dashboard category value to a `TechIcon` glyph.
- [x] 3.4 Preserve fallback icon behavior for missing or unknown categories.
- [x] 3.5 Confirm service cards, service detail, and service form selection render expanded categories correctly.

## 4. Verification

- [x] 4.1 Run `make test-go-all`.
- [x] 4.2 Run `make lint-dashboard`.
- [x] 4.3 Run `make check-dashboard`.
- [x] 4.4 Run `make test-dashboard` if dashboard category tests are added or changed.
