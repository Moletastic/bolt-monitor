# Bruno Workspace

Workspace for Bolt Monitor API exploration and manual testing.

## Contents

- `collections/bolt-monitor-api/`: primary API collection
- `environments/`: shared workspace environments

## Current coverage

- Every method-and-path route wired in `infra/stacks/bootstrap.ts`
- Health, search, channels, policies, services, service-scoped monitors, incidents, and admin scheduler config

## Suggested flow

1. Open this workspace in Bruno.
2. Select the `development` environment.
3. Run `Create Service` or `List Services`.
4. Copy returned `serviceId` into the environment.
5. Run `Create Service Monitor`.
6. Copy returned `monitorId` into the environment.
7. Run monitor read, run, patch, disable, and enable requests.

## Notes

- Current default tenant is `DEFAULT` in the API.
- Configure `apiUrl` in ignored `.bruno/environments/development.local.yml`.
- Requests assume local SST or deployed staging API uses the same `apiUrl` variable.
- Run `make check-bruno` after changing API routes or Bruno requests.
