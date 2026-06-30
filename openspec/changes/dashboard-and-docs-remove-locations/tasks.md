## 1. Dashboard

- [ ] 1.1 Remove `listProbeLocations` API client and probe-location TypeScript types.
- [ ] 1.2 Remove `getMonitorLocationField` helper and associated tests.
- [ ] 1.3 Remove probe-location field rendering and hidden inputs from monitor forms.
- [ ] 1.4 Remove probe-location resolution from create/update monitor server actions.
- [ ] 1.5 Remove probe-location columns, labels, chips, and status fallbacks from monitor table/detail views.
- [ ] 1.6 Remove `/locations` page and navigation links.
- [ ] 1.7 Remove probe-location settings summary from `/config`.

## 2. Documentation and API Docs

- [ ] 2.1 Remove `GET /api/v1/probe-locations` from `openapi/openapi.yaml`.
- [ ] 2.2 Remove `probeLocations`, `probeLocationId`, `lastProbeLocationId`, and `iad` examples from OpenAPI schemas/examples.
- [ ] 2.3 Update dashboard README and repo docs to remove single-region preview/catalog language.
- [ ] 2.4 Update smoke-test/checklist references that include `/locations` or probe-location chips.

## 3. Verification

- [ ] 3.1 Run `make lint-dashboard`.
- [ ] 3.2 Run `make check-dashboard`.
- [ ] 3.3 Run `make test-dashboard`.
- [ ] 3.4 Run `make build-dashboard`.
- [ ] 3.5 Run `openspec validate dashboard-and-docs-remove-locations --strict`.
