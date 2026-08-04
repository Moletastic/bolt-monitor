## Why

Repository validation scripts protect important API and deployment contracts, but the current command surface contains a duplicate manual auth checker, dormant test files, repeated test execution, and source-shape checks that duplicate behavior-level infrastructure tests. Simplifying this suite reduces CI and maintenance overhead without weakening deterministic release confidence.

## What Changes

- Remove the redundant manual auth-route validator, its test, and its Makefile target; retain authentication boundary coverage in infrastructure tests.
- Run every root `scripts/*.test.mjs` test exactly once through the infrastructure validation surface so helper-parser tests are not dormant.
- Remove duplicate root-script test invocations while retaining separate production validation commands.
- Extract only small shared filesystem and command-reporting helpers from root validators; preserve validator-specific route, Bruno, OpenAPI, and lifecycle rules.
- Replace source-shape lifecycle assertions in the auth cutover prerequisite gate with behavior-level infrastructure tests where equivalent coverage exists, and remove the gate if that coverage is complete.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `repository-ci`: CI and local Makefile validation commands run deterministic checks without redundant execution and exercise all repository script tests.

## Impact

- Affected code: root `scripts/`, `Makefile`, infrastructure tests, and CI command grouping.
- No HTTP API, deployed infrastructure behavior, external dependencies, or cloud credentials change.
- API contract, Bruno, OpenAPI, and handler route-drift validation remain required release gates.
