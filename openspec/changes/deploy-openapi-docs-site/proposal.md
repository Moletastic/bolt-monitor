## Why

The repo now has local OpenAPI docs, but they are only available through a developer-run local server. Deploying those docs as an SST-managed static site makes the API reference easy to share and turns the docs workflow into part of the actual infrastructure toolbelt.

## What Changes

- Add an SST-managed static docs site that serves the existing OpenAPI documentation workspace.
- Wire the deployed docs site to use the stack's real API URL in the published OpenAPI contract so interactive docs target the live API.
- Expose the deployed docs URL as a stack output and document how to access it.

## Capabilities

### New Capabilities
- `api-documentation-site`: Hosted API documentation site deployed through SST, backed by the existing OpenAPI docs workspace and configured to target the deployed API URL.

### Modified Capabilities

## Impact

- Affects `infra/` stack wiring and outputs.
- Reuses the existing `openapi/` workspace as the site source.
- Adds a deployed documentation surface alongside the existing API Gateway and Lambda resources.
