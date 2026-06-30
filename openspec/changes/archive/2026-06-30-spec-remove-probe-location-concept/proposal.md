## Why

Bolt Monitor no longer plans to support multi-regional health checking. Keeping probe locations as a first-class product concept creates implementation and infrastructure complexity for a future that is intentionally out of scope: monitor payloads carry `probeLocations`, the scheduler fans out by location, results store location identifiers, and the dashboard presents a single-region preview as if it may become a selector later.

There is no production environment or persisted data compatibility requirement, so the product contract can be simplified directly instead of preserving deprecated location fields.

## What Changes

- Remove probe locations as an operator-facing and API-facing product concept.
- Define monitor execution as running from the system's execution environment, not from selected regions.
- Update specs so monitors no longer require `probeLocations` and schedulers no longer create monitor-location fan-out work.
- Retire the probe-location catalog capability and stop promising future multi-region selection.

## Capabilities

### Modified Capabilities

- `monitor-configuration`: Monitor configuration no longer includes selected probe locations or catalog validation.
- `scheduler-eventbridge-trigger`: Scheduler creates one execution request per enabled monitor instead of one per monitor-location pair.
- `check-result-status-model`: Raw results and status snapshots no longer require probe location identity.

### Removed Capabilities

- `probe-location-catalog`: The canonical probe-location catalog is removed from the product model.

## Impact

- Establishes the source-of-truth decision for follow-on backend and dashboard changes.
- Enables removal of `probeLocations`, `probeLocationId`, `lastProbeLocationId`, `GET /api/v1/probe-locations`, the dashboard locations page, and related OpenAPI examples in later phases.
- This change is spec-only and should be completed before implementation changes.
