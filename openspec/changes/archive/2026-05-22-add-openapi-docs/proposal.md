## Why

The repo already exposes a small HTTP API through API Gateway and Go Lambda handlers, but there is no explicit machine-readable contract or local documentation surface for it. Adding OpenAPI-backed local docs makes the API easier to inspect, demo, and evolve without forcing a larger infrastructure rewrite.

## What Changes

- Add a source-controlled OpenAPI document that describes the current public API routes and payloads.
- Add a local Swagger UI workflow for interactive API exploration against the OpenAPI document.
- Add a local Redoc workflow for cleaner static API reference rendering from the same OpenAPI document.
- Add a single documented developer command surface for viewing API docs locally.

## Capabilities

### New Capabilities
- `api-documentation`: Local OpenAPI-based API documentation, including interactive Swagger UI and static Redoc views backed by the same contract file.

### Modified Capabilities

## Impact

- Adds a new `openapi/` documentation workspace and local docs tooling.
- Describes existing API routes implemented in `infra/` and `services/` without changing their runtime behavior.
- Affects developer workflow documentation so contributors can open local API docs consistently.
