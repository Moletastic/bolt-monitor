## Why

After the backend contract moves to single-environment execution, the dashboard and documentation should stop presenting probe locations as a visible product surface. Keeping a locations page, picker chip, region copy, and OpenAPI examples would mislead operators and preserve a multi-region promise that the project does not intend to support.

## What Changes

- Remove the dashboard locations page and any navigation entries pointing to it.
- Remove monitor form probe-location fields, hidden inputs, helper copy, and catalog fetches.
- Remove probe-location display from monitor tables, monitor detail status cards, settings, and run history.
- Remove dashboard client types/helpers/API calls for probe locations.
- Update OpenAPI examples and schemas to remove `iad`, `probeLocations`, `probeLocationId`, `lastProbeLocationId`, and `GET /api/v1/probe-locations`.
- Update README/AGENTS/docs that mention the previous single-region preview or catalog.

## Capabilities

### Modified Capabilities

- `dashboard-web-app`: Dashboard no longer presents probe-location selection, summaries, routes, or status fields.
- `api-documentation`: Public API documentation no longer documents probe-location contracts or examples.

## Impact

- `apps/dashboard`
- `openapi/openapi.yaml`
- Repository docs and OpenSpec references
- Dashboard lint/type/test/build verification
