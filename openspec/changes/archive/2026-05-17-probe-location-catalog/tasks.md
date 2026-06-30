## 1. Define Probe-Location Catalog Model

- [x] 1.1 Add shared probe-location catalog model with stable ID, display metadata, and enabled state.
- [x] 1.2 Encode vendor-neutral naming and validation rules for valid monitor probe-location selection.

## 2. Update Monitor Configuration Contract

- [x] 2.1 Rename monitor execution-location semantics from `regions` to `probeLocations` in shared contracts and validation.
- [x] 2.2 Validate monitor-selected probe locations against the system-defined catalog contract.

## 3. Document Integration Boundaries

- [x] 3.1 Document how system-owned probe locations differ from tenant/user selection.
- [x] 3.2 Run relevant checks to confirm shared contracts compile cleanly and reflect the updated probe-location semantics.
