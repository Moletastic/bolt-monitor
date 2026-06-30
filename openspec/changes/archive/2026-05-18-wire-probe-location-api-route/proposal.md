## Why

Dashboard v1 can ship with current built-in `iad` probe-location assumption, but runtime probe-location discovery is now defined in source-of-truth specs and implemented in the monitor API service. Bootstrap infrastructure still does not route `GET /api/v1/probe-locations` through API Gateway, so frontend consumers cannot rely on that endpoint in local SST or deployed environments.

## What Changes

- Wire `GET /api/v1/probe-locations` through SST bootstrap API to the existing monitor API handler.
- Verify local and deployed API surface exposes the route consistently with the probe-location read spec.
- Keep scope limited to infrastructure and verification; do not redesign dashboard forms in this change.

## Impact

- Closes gap between `probe-location-read-api` spec and live API routing.
- Enables future dashboard forms to switch from hardcoded bootstrap assumptions to runtime location discovery.
- Keeps dashboard-v1 from silently expanding scope while preserving explicit follow-on path.
