# Monitor Configuration Model

Shared canonical monitor contract for future API, persistence, scheduler, and prober code.

Persistence assumption: monitor records should fit a single-table DynamoDB design with explicit key metadata, instead of separate tables per model.

Storage contract reference: `shared/dynamodbschema` defines canonical item families, PK/SK patterns, GSIs, and retention expectations that monitor records should map into.

## Required top-level fields

- `monitorId`: stable monitor identity
- `tenantId`: ownership boundary
- `name`: human-readable monitor name
- `type`: monitor kind; v1 supports `http`
- `intervalSeconds`: check cadence, constrained to supported minute-based presets: 60, 120, 180, 300, 600, 900, 1800, or 3600
- `probeLocations`: selected execution locations from the system-owned probe-location catalog
- `enabled`: lifecycle toggle for scheduling and execution

## HTTP monitor v1 fields

- `http.target`: absolute URL
- `http.method`: allowed HTTP verb
- `http.timeoutMs`: positive timeout
- `http.expectedStatusCodes`: optional expected status list
- `http.expectedBodyContains`: optional substring assertion

Disabled monitors remain valid configuration records. Downstream schedulers and probers must skip execution when `enabled` is `false`.

Monitors do not invent arbitrary location strings. `probeLocations` should be validated against the system-owned catalog defined in `shared/probelocationcatalog`.
