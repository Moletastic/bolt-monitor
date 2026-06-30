# Probe Location Catalog

System-owned catalog of valid probe locations.

## Ownership model

- system defines available probe locations
- tenant or user selects from that catalog
- monitors reference selected `probeLocations`

## Required fields

- `locationId`: stable vendor-neutral identifier such as `iad` or `dub`
- `displayName`: human-readable label
- `executionTarget`: scheduler or worker routing target
- `enabled`: whether location is selectable and routable

Users do not invent arbitrary location strings. Monitor validation must reference this catalog.
