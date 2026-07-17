## 1. Contract Model And Validator Foundations

- [x] 1.1 Define one normalized route record for method, parameter-preserving path, target handler, source location, and optional public/protected classification, with unit fixtures for path and query normalization.
- [x] 1.2 Refactor the existing Bruno parser and convention checks into reusable validator modules without weakening `make check-bruno` behavior or diagnostics.
- [x] 1.3 Add OpenAPI operation extraction and tests for methods, paths, operation locations, security inheritance, explicit public overrides, and protected security schemes.
- [x] 1.4 Add SST route extraction and tests for handler target and optional authentication metadata, including failure on partial auth classification once any route metadata is present.
- [x] 1.5 Add merged OpenSpec route extraction limited to `openspec/specs/`, preserve source locations, and test that active change routes are excluded.

## 2. Handler Route Reachability

- [x] 2.1 Introduce an explicit method/path inventory for `services/monitor-api` that represents every statically supported dispatch branch with exact route parameter names.
- [x] 2.2 Refactor or test monitor API dispatch so every inventory route reaches its intended handler and every statically identifiable dispatch branch is represented.
- [x] 2.3 Add validator comparison between monitor-handler SST routes and the Go route inventory, with fixtures for handler-only, SST-only, and unsupported dynamic patterns.
- [x] 2.4 Capture the current handler-only and SST-only discrepancies as failing validator fixtures or an initial gate failure, then explicitly reconcile each against merged or active OpenSpec requirements by wiring required archive, reactivate, maintenance, escalation-state, or other routes and removing only behavior with no requirement.
- [x] 2.5 Make the reconciled handler/SST comparison a required green prerequisite before `add-single-tenant-operator-authentication` refactors registrations through its protected-v1 route helper.

## 3. API Asset Reconciliation

- [x] 3.1 Extend the contract validator to fail on missing or stale SST/Bruno/OpenAPI routes and on explicit merged OpenSpec routes absent from SST, with source-specific actionable diagnostics.
- [x] 3.2 Extend Bruno route metadata and validation for public/protected classification when SST auth metadata exists, without storing credentials in collection files.
- [x] 3.3 Update Bruno requests for every reconciled SST route while preserving verb-resource names, exact route variables, domain/operation tags, and Purpose/Setup/Expected result docs.
- [x] 3.4 Rebuild `openapi/openapi.yaml` route operations around the deployed service-first API surface, exact parameter names, response envelopes, and current request/response schemas.
- [x] 3.5 Add OpenAPI security metadata that keeps `GET /api/health` public and mirrors protected SST routes when authentication metadata is available.
- [x] 3.6 Add validator fixtures for missing, stale, parameter-mismatched, OpenSpec-only, handler-only, unhandled, and authentication-conflicting routes and assert the diagnostic text needed to fix each failure.

## 4. Health And Documentation Alignment

- [x] 4.1 Update the Go health Lambda to return the shared standard success envelope with a stable healthy data value, and update handler tests for status, data, and omitted error-only fields.
- [x] 4.2 Update the OpenAPI health operation and schemas to document the same success envelope and public authentication behavior.
- [x] 4.3 Correct README health examples, common validation commands, and authentication notes to match the deployed contract.
- [x] 4.4 Correct README architecture and repository-layout guidance to include the current SST dashboard, DynamoDB, execution and notification queues, scheduler/worker runtime, and escalation runtime data flow.
- [x] 4.5 Update contributor route-governance guidance to require synchronized SST, handler inventory, Bruno, OpenAPI, and merged OpenSpec updates and document the relevant Makefile checks.
- [x] 4.6 Document completed OpenSpec archival as post-verification maintenance guidance, explicitly separate from runtime/build/test gates.

## 5. Phase 0 Local Release Gates And Ordinary CI

- [x] 5.1 Add a root Makefile target for deterministic API-contract validation and its tests while retaining `make check-bruno` as the Bruno governance entry point.
- [x] 5.2 Add or consolidate non-mutating root Makefile targets for Go format/vet/lint/tests/build, dashboard format/lint/typecheck/tests/production build, and infrastructure format/typecheck so workflows do not duplicate command lists.
- [x] 5.3 Update GitHub pull-request and `main` CI to invoke the verified Makefile targets, including `make build-dashboard`, `make check-bruno`, and API-contract validation.
- [x] 5.4 Keep dashboard and infrastructure installs frozen to committed pnpm lockfiles and verify each package root's explicit install-script allowlist remains enforced.
- [x] 5.5 Bound ordinary CI cost with useful toolchain parallelism and safe lockfile/module caches, avoid duplicate production builds, and confirm no ordinary job deploys SST or requests AWS credentials.
- [x] 5.6 Verify intentional validator failures identify the failed source, file or operation, normalized route, and expected remedy in GitHub Actions logs.
- [x] 5.7 Add deterministic checks for portable, non-interactive SST profile and stage selection where deployment tooling consumes them, preserving the documented local default and rejecting production smoke targets without requiring AWS credentials.
- [x] 5.8 Verify the required pre-cutover gate includes dashboard production build, `make check-bruno`, SST/OpenAPI/Bruno/handler route drift, health envelope/documentation alignment, and portable profile/stage checks, and record it as complete before the authentication security cutover.
- [x] 5.9 Update roadmap dependency language so this local deterministic foundation can move to Phase 0 independently of authentication while the staging auth-smoke milestone remains dependency-gated.

## 6. Post-Authentication Staging Smoke

- [x] 6.1 After `add-single-tenant-operator-authentication` provides protected route metadata and a staging token flow, add locally testable smoke helpers that reject production stage names, validate SST API output, select a read-only protected route, require the health envelope, and assert API Gateway missing-token HTTP 401 without requiring an application envelope.
- [x] 6.2 Preserve the documented non-interactive `bolt-monitor` local profile default while keeping repository CI free of AWS credentials.
- [x] 6.3 Depend on `standardize-stage-resource-lifecycle` and select the declared long-lived persistent staging environment; prohibit unique or per-run stages containing retained resources.
- [x] 6.4 Document an explicit local operator procedure that deploys the current revision to declared staging and obtains the API URL from structured SST output.
- [x] 6.5 Configure the local smoke helper to avoid request tracing and never print credentials, tokens, or authorization headers.
- [x] 6.6 Assert HTTP 401 for a missing token and valid-token acceptance against the same read-only protected route; treat the missing-token result as an API Gateway edge response with no application-envelope guarantee and fail clearly if required local token material is unavailable.
- [x] 6.7 Retain the declared persistent staging environment and verify no unique stage was created.
- [x] 6.8 Ensure repository CI never receives AWS/token secrets, assumes deployment roles, deploys SST, or targets production.

## 7. End-To-End Verification

- [x] 7.1 Run validator unit/fixture tests, `make check-bruno`, and the new API-contract Makefile target against the reconciled repository.
- [x] 7.2 Run the verified Go format, vet, lint, test, and build targets from the repository root.
- [x] 7.3 Run dashboard format check, lint, typecheck, tests, and `make build-dashboard` from the repository root.
- [x] 7.4 Run infrastructure format check and typecheck from the repository root and validate workflow syntax without AWS credentials.
- [x] 7.5 Manually invoke the local staging smoke helper with a fresh access token after deliberately deploying the declared persistent staging environment; verify health envelope, Gateway edge 401, valid-token acceptance, and secret-safe logs, then confirm the declared persistent staging identity.
- [x] 7.6 Review the final diff to confirm no production deploy trigger, broad CI platform redesign, dependency trust relaxation, or archive-state behavior test was introduced.
