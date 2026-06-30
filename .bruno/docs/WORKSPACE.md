# Bruno Workspace

Workspace for Bolt Monitor API exploration and manual testing.

## Contents

- `collections/bolt-monitor-api/`: primary API collection
- `environments/`: shared workspace environments

## Current coverage

- health endpoint
- monitor CRUD endpoints

## Suggested flow

1. Open this workspace in Bruno.
2. Select the `development` environment.
3. Run `Create Monitor`.
4. Copy returned `monitorId` into collection/request variables.
5. Run list, get, patch, disable, and enable requests.

## Notes

- Current default probe location is `iad`.
- Current default tenant is `default`.
- Requests assume local SST or deployed dev API uses same base URL variable.
