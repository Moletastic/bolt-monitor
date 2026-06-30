## Why

Repo has SST bootstrap but no deployed service path yet. Adding tiny Go Lambda behind API Gateway proves first end-to-end backend slice now, before probe logic and persistence add noise.

## What Changes

- Add first backend HTTP endpoint exposed through API Gateway.
- Add Go Lambda handler under `services/` that returns static health response.
- Wire SST stack to deploy API Gateway route to Go Lambda.
- Document local and deployed validation flow for endpoint.

## Capabilities

### New Capabilities
- `api-health-endpoint`: Public HTTP health endpoint backed by Go Lambda and API Gateway.

### Modified Capabilities

## Impact

- Affects `infra/` SST stack definitions and routing.
- Adds first Go service code under `services/`.
- Adds Go module/dependency setup for Lambda runtime.
- Establishes backend pattern future API and probe endpoints can follow.
