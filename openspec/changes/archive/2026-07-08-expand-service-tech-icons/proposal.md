## Why

Service categories currently cover only a small set of generic infrastructure shapes, which makes many real services collapse into vague choices like `server` or `http`. Expanding the generic technology icon catalog will help operators recognize services faster in cards, details, and creation flows without relying on brand-specific logos.

## What Changes

- Add more supported service category values for common service types.
- Add matching dashboard technology icons and labels for every supported category.
- Keep existing service category values valid.
- Keep icons generic and product-owned rather than vendor or brand logos.
- Ensure dashboard TypeScript category definitions stay aligned with backend validation.
- Use the expanded catalog anywhere the dashboard lists or selects service categories.

Initial technical category expansion:
- `web`: browser-facing web application.
- `api`: API or backend interface.
- `worker`: background worker or job processor.
- `scheduler`: scheduled or cron-like workload.
- `storage`: object/file storage service.
- `search`: search or indexing service.
- `auth`: authentication or identity service.
- `payments`: billing, checkout, or payment-processing service.
- `analytics`: analytics, reporting, or event-processing service.
- `observability`: logging, metrics, tracing, or monitoring service.
- `ai`: AI, model, or inference service.
- `integration`: third-party integration or external connector.

Initial service-purpose category expansion:
- `media`: image, audio, video, streaming, or asset-processing service.
- `content`: blog, CMS, publishing, documentation, or editorial service.
- `finance`: accounting, ledger, invoicing, or financial operations service.
- `learning`: education, courses, training, or knowledge-delivery service.
- `gaming`: gameplay, matchmaking, achievements, or game platform service.
- `commerce`: storefront, catalog, cart, orders, or marketplace service.
- `messaging`: chat, email, notifications, or communication service.
- `support`: help desk, ticketing, customer success, or support workflow service.
- `marketing`: campaigns, attribution, landing pages, or growth tooling service.
- `admin`: internal administration, back office, or operator tooling service.
- `security`: fraud, permissions, compliance, or security-control service.
- `location`: maps, geospatial, routing, or place-data service.
- `social`: community, profiles, feeds, or social graph service.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `monitor-configuration`: Expand the supported service category enum accepted by service create/update validation.
- `dashboard-web-app`: Render and select the expanded service technology icon catalog across dashboard service surfaces.

## Impact

- Affected shared domain model: `shared/monitorconfig/model.go` and tests.
- Affected dashboard types: `apps/dashboard/lib/types.ts`.
- Affected dashboard rendering: `apps/dashboard/components/tech-icon.tsx`, `ServiceIcon` consumers, service cards, service detail, and service form selection.
- No storage migration is required because service category is stored as a string and existing values remain valid.
- No external icon dependency is expected.
