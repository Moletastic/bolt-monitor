## 1. Contract Model And Validator Foundations

- [ ] 1.1 Define one normalized route record for method, parameter-preserving path, target handler, source location, and optional public/protected classification, with unit fixtures for path and query normalization.
- [ ] 1.2 Refactor the existing Bruno parser and convention checks into reusable validator modules without weakening `make check-bruno` behavior or diagnostics.
- [ ] 1.3 Add OpenAPI operation extraction and tests for methods, paths, operation locations, security inheritance, explicit public overrides, and protected security schemes.
- [ ] 1.4 Add SST route extraction and tests for handler target and optional authentication metadata, including failure on partial auth classification once any route metadata is present.
- [ ] 1.5 Add merged OpenSpec route extraction limited to `openspec/specs/`, preserve source locations, and test that active change routes are excluded.

## 2. Handler Route Reachability

- [x] 2.1 Introduce an explicit method/path inventory for `services/monitor-api` that represents every statically supported dispatch branch with exact route parameter names.
- [ ] 2.2 Refactor or test monitor API dispatch so every inventory route reaches its intended handler and every statically identifiable dispatch branch is represented.
- [ ] 2.3 Add validator comparison between monitor-handler SST routes and the Go route inventory, with fixtures for handler-only, SST-only, and unsupported dynamic patterns.
- [x] 2.4 Capture the current handler-only and SST-only discrepancies as failing validator fixtures or an initial gate failure, then explicitly reconcile each against merged or active OpenSpec requirements by wiring required archive, reactivate, maintenance, escalation-state, or other routes and removing only behavior with no requirement.
- [ ] 2.5 Make the reconciled handler/SST comparison a required green prerequisite before `add-single-tenant-operator-authentication` refactors registrations through its protected-v1 route helper.

## 3. API Asset Reconciliation

- [x] 3.1 Extend the contract validator to fail on missing or stale SST/Bruno/OpenAPI routes and on explicit merged OpenSpec routes absent from SST, with source-specific actionable diagnostics.
- [ ] 3.2 Extend Bruno route metadata and validation for public/protected classification when SST auth metadata exists, without storing credentials in collection files.
- [x] 3.3 Update Bruno requests for every reconciled SST route while preserving verb-resource names, exact route variables, domain/operation tags, and Purpose/Setup/Expected result docs.
- [ ] 3.4 Rebuild `openapi/openapi.yaml` route operations around the deployed service-first API surface, exact parameter names, response envelopes, and current request/response schemas.
- [ ] 3.5 Add OpenAPI security metadata that keeps `GET /api/health` public and mirrors protected SST routes when authentication metadata is available.
- [ ] 3.6 Add validator fixtures for missing, stale, parameter-mismatched, OpenSpec-only, handler-only, unhandled, and authentication-conflicting routes and assert the diagnostic text needed to fix each failure.

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
- [ ] 5.4 Keep dashboard and infrastructure installs frozen to committed pnpm lockfiles and verify each package root's explicit install-script allowlist remains enforced.
- [ ] 5.5 Bound ordinary CI cost with useful toolchain parallelism and safe lockfile/module caches, avoid duplicate production builds, and confirm no ordinary job deploys SST or requests AWS credentials.
- [ ] 5.6 Verify intentional validator failures identify the failed source, file or operation, normalized route, and expected remedy in GitHub Actions logs.
- [x] 5.7 Add deterministic checks for portable, non-interactive SST profile and stage selection where deployment tooling consumes them, preserving the documented local default and rejecting production smoke targets without requiring AWS credentials.
- [ ] 5.8 Verify the required pre-cutover gate includes dashboard production build, `make check-bruno`, SST/OpenAPI/Bruno/handler route drift, health envelope/documentation alignment, and portable profile/stage checks, and record it as complete before the authentication security cutover.
- [ ] 5.9 Update roadmap dependency language so this local deterministic foundation can move to Phase 0 independently of authentication while the staging auth-smoke milestone remains dependency-gated.

## 6. Post-Authentication Staging Smoke

- [ ] 6.1 After `add-single-tenant-operator-authentication` provides protected route metadata and a staging token flow, add locally testable smoke helpers that reject production stage names, validate SST API output, select a read-only protected route, require the health envelope, and assert API Gateway missing-token HTTP 401 without requiring an application envelope.
- [ ] 6.2 Make SST credential/profile selection work non-interactively in protected CI while preserving the documented `bolt-monitor` local profile default.
- [ ] 6.3 Depend on `standardize-stage-resource-lifecycle` and select exactly one approved lifecycle: a clean ephemeral stage whose teardown leaves zero retained resources, or a declared long-lived persistent staging environment; prohibit unique or per-run stages containing retained resources.
- [ ] 6.4 Add a manually dispatched or protected credential-gated workflow that deploys the current revision under the selected non-production lifecycle, uses concurrency control, and obtains the API URL from structured SST output.
- [ ] 6.5 Configure the smoke workflow to mask secrets, disable request tracing, call public health without credentials, and never persist credentials or authorization headers in logs or artifacts.
- [ ] 6.6 Assert HTTP 401 for a missing token and valid-token acceptance against the same read-only protected route; treat the missing-token result as an API Gateway edge response with no application-envelope guarantee and fail clearly if required staging token material is unavailable.
- [ ] 6.7 For ephemeral lifecycle, run always-on cleanup, verify zero retained resources remain, preserve the original validation result, and report cleanup failures; for persistent lifecycle, retain the declared environment and verify no unique stage was created.
- [ ] 6.8 Ensure untrusted pull-request events cannot receive AWS/token secrets, assume deployment roles, or invoke the credentialed smoke job, and ensure no workflow path targets production.

## 7. End-To-End Verification

- [x] 7.1 Run validator unit/fixture tests, `make check-bruno`, and the new API-contract Makefile target against the reconciled repository.
- [x] 7.2 Run the verified Go format, vet, lint, test, and build targets from the repository root.
- [x] 7.3 Run dashboard format check, lint, typecheck, tests, and `make build-dashboard` from the repository root.
- [x] 7.4 Run infrastructure format check and typecheck from the repository root and validate workflow syntax without AWS credentials.
- [ ] 7.5 After authentication and stage-lifecycle dependencies are complete, manually invoke staging smoke with protected credentials, verify the health envelope, Gateway edge 401, valid-token acceptance, and secret-safe logs, then confirm zero residual ephemeral resources or the declared persistent staging identity.
- [x] 7.6 Review the final diff to confirm no production deploy trigger, broad CI platform redesign, dependency trust relaxation, or archive-state behavior test was introduced.
