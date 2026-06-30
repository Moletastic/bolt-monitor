## 1. Shared Domain

- [ ] 1.1 Remove `ProbeLocations` from monitor domain, create/update requests, summaries, records, and validation.
- [ ] 1.2 Remove `ProbeLocationCatalog` / `ValidateWithCatalog` dependencies from monitor validation.
- [ ] 1.3 Remove `shared/probelocationcatalog` and tests when no longer referenced.
- [ ] 1.4 Remove probe-location fields from execution request/result/work models.
- [ ] 1.5 Remove probe-location fields from result status and check run models.

## 2. Runtime and API

- [ ] 2.1 Simplify scheduler request construction to one request per enabled monitor.
- [ ] 2.2 Simplify worker request rebuilding to validate monitor enabled/config only.
- [ ] 2.3 Simplify manual run path to execute without selecting the first probe location.
- [ ] 2.4 Remove incident summary phrasing that says failures occurred from a location.
- [ ] 2.5 Remove `GET /api/v1/probe-locations` handler and infra route.
- [ ] 2.6 Remove location fields from monitor, status, run, and manual-run responses.

## 3. Tests and Verification

- [ ] 3.1 Update Go tests to remove `iad`, `probeLocations`, `probeLocationId`, and `lastProbeLocationId` expectations.
- [ ] 3.2 Run `make test-go-all`.
- [ ] 3.3 Run `make lint-go`.
- [ ] 3.4 Run `openspec validate backend-single-execution-environment --strict`.
