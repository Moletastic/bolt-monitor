## Why

Repository checks do not currently prove that the production dashboard builds or that the deployed SST routes, Go handler behavior, Bruno collection, OpenAPI contract, and operator documentation agree. Known handler/SST route drift also makes an authentication route-helper refactor unsafe: the refactor could preserve only the infrastructure inventory and silently strand handler behavior. Strengthening local release gates before the security cutover prevents undocumented or unreachable API behavior and catches authentication and response-envelope regressions without turning CI into a deployment pipeline.

## What Changes

- Split delivery within this change: land a Phase 0 local deterministic CI foundation before the authentication/security cutover, then enable credentialed staging authentication smoke only after authentication and its protected route metadata exist.
- Expand pull-request and `main` CI to run the repository's verified Go, dashboard, infrastructure, dashboard production-build, Bruno, contract-drift, health-contract, documentation, and portable profile/stage configuration checks with actionable failures and bounded cost.
- Add deterministic drift checks that compare normalized method/path routes across SST, Bruno, OpenAPI, explicit merged OpenSpec route requirements, and statically enumerable handler behavior. The known handler/SST drift must fail the gate or be explicitly reconciled before the authentication change refactors route registration through its protected-route helper.
- Validate public/protected authentication metadata across route and API-documentation sources when the infrastructure exposes that metadata.
- Keep the checked-in API contract and repository guidance aligned with the deployed architecture, including the standard health response envelope.
- Add an opt-in local staging authentication smoke helper only after authentication exists. It runs from an operator workstation after an explicit deploy and uses a declared long-lived persistent staging environment, never a unique stage containing retained resources.
- Assert API Gateway missing-token rejection as an HTTP 401 edge response without requiring the application JSON envelope; continue to require the standard application envelope for public health.
- Preserve frozen dependency installs and the existing explicit install-script trust policy.
- Document completed OpenSpec archival as contributor maintenance guidance rather than treating archived-change state as an application behavior or CI correctness test.
- Exclude production deployment automation and broad CI-platform redesign.

## Capabilities

### New Capabilities
- `api-contract-release-gates`: Deterministic SST, Bruno, OpenAPI, and feasible Go handler route/metadata drift detection with actionable diagnostics.
- `staging-release-smoke`: Post-authentication, opt-in local staging authentication smoke validation with lifecycle-safe resources, edge-aware assertions, no secret disclosure, and no production deployment.

### Modified Capabilities
- `repository-ci`: Require the complete verified command set, including dashboard production build and Bruno/API contract checks, while keeping CI reproducible and cost-bounded.
- `api-documentation`: Require the checked-in OpenAPI contract and repository guidance to reflect the deployed route surface, authentication metadata, architecture, and maintenance workflow.
- `api-health-endpoint`: Require the public health endpoint and its documentation/smoke assertion to use the standard response envelope.

## Impact

- Affected repository surfaces include `.github/workflows/`, the root `Makefile`, route-governance scripts and tests, `infra/stacks/bootstrap.ts`, `services/monitor-api` route declarations or route metadata, `.bruno/collections/`, `openapi/openapi.yaml`, `README.md`, and contributor/OpenSpec guidance.
- The Phase 0 local foundation has no authentication dependency and is suitable for moving to roadmap Phase 0 ahead of the security cutover. Only the staging authentication-smoke milestone depends on `add-single-tenant-operator-authentication` and `standardize-stage-resource-lifecycle`; roadmap dependency language should preserve that split rather than blocking all local gates on authentication.
- Pull-request CI gains local deterministic validation and a dashboard production build; CI never receives AWS credentials or triggers cloud work. AWS-backed staging smoke remains an explicit operator workstation procedure after authentication exists.
- No production deployment automation, application feature expansion, package-manager migration, or relaxation of dependency install-script trust is introduced.
