## Why

The Bruno collection has drifted from the SST API routes: it still uses the deprecated top-level monitor model while the API is service-first, and it omits several exposed endpoints. This change establishes one maintained manual-testing collection and a local guard so route and metadata drift is detected early.

## What Changes

- Replace deprecated top-level monitor requests with service-scoped requests.
- Add Bruno requests for every exposed method-and-path route, organized by API domain.
- Standardize request names, exact route-variable names, descriptions, and `domain`/`operation` tags.
- Add a static local `make check-bruno` guard that compares normalized SST routes with Bruno requests and validates required metadata.
- Remove deprecated requests from the primary collection.
- Document Bruno maintenance responsibilities in `AGENTS.md` and the governing principle in `CONSTITUTION.md`.
- Report OpenSpec routes missing from SST wiring separately; do not treat that report as Bruno coverage.

## Capabilities

### New Capabilities

- `bruno-api-collection-governance`: Maintained Bruno coverage and conventions for exposed API routes.

### Modified Capabilities

- None.

## Impact

- Affects `.bruno/collections/`, `.bruno/docs/`, `infra/stacks/bootstrap.ts` route extraction, `Makefile`, `AGENTS.md`, and `CONSTITUTION.md`.
- Adds local validation tooling; no production API behavior changes.
- Existing manual-testing workflows using deprecated top-level monitor URLs must migrate to service-scoped variables and requests.
