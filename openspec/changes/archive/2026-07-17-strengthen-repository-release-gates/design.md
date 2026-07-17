## Context

The repository already has GitHub CI, Makefile validation targets, a static Bruno route checker, a checked-in OpenAPI document, and SST route declarations in `infra/stacks/bootstrap.ts`. These surfaces have drifted: ordinary CI does not build the production dashboard or run Bruno governance, Bruno only compares itself with SST, OpenAPI describes legacy non-service-scoped routes and pre-envelope payloads, README health and architecture details are stale, and `services/monitor-api/handler.go` contains route behavior not wired in SST. This known handler/SST drift is a cutover blocker because the separate authentication change will refactor route registration through a protected-v1 helper.

The implementation crosses CI, infrastructure metadata, Go routing, API assets, tests, and contributor documentation. It must retain immutable pnpm installs and explicit install-script trust, keep cloud credentials away from untrusted pull requests, and avoid production deployment automation.

## Goals / Non-Goals

**Goals:**

- Make ordinary CI exercise the repository's verified Go, dashboard, infrastructure, Bruno, and API-contract checks, including the dashboard production build.
- Fail deterministically when SST, handler routing, Bruno, or OpenAPI disagree about a statically representable route.
- Carry public/protected classification through contract assets when infrastructure auth metadata becomes available.
- Land local deterministic release gates before authentication changes route registration, then provide an isolated, explicit staging deploy smoke after authentication exists.
- Align README/OpenAPI/OpenSpec maintenance guidance with deployed architecture and the standard response envelope.
- Keep failures actionable and local CI cost bounded.
- Support portable, non-interactive SST profile and stage configuration without weakening production-stage guards or the documented local default.

**Non-Goals:**

- Automating production deployment or changing production promotion policy.
- Redesigning GitHub Actions broadly, introducing a new CI platform, or requiring AWS credentials in ordinary CI.
- Implementing operator authentication itself; the separate authentication change owns authorizers, identity, and token issuance.
- Migrating `openapi/` from npm or relaxing dependency install-script allowlists.
- Proving arbitrary dynamic Go control flow through a general-purpose static analyzer.
- Treating OpenSpec archive state as application behavior or replacing implementation verification with archival checks.

## Decisions

### 1. Keep SST as deployed-route authority and compare independent representations

The contract validator will normalize method/path pairs from SST, Bruno, and OpenAPI and require exact set equality for deployed routes. It will also perform a one-way check that explicit method/path requirements in merged `openspec/specs/` are wired in SST; active changes are excluded because they can describe future behavior. Path parameters normalize to `{parameterName}` and query strings are excluded from route identity. Diagnostics name the source, file or operation, method, and normalized path.

SST remains authoritative because it determines API Gateway deployment. Bruno and OpenAPI remain independent consumer-facing assets so the gate can detect omissions rather than generating both silently from one source.

Alternative considered: generate Bruno and OpenAPI directly from SST. This would reduce maintenance but could faithfully generate an incomplete contract from an accidental infrastructure edit and would not validate payload documentation.

### 2. Make handler routes explicitly enumerable and block auth refactoring on current drift

The monitor handler will expose a declarative method/path inventory adjacent to its dispatch behavior. Handler matching can continue to call existing functions, but tests will establish that the inventory and dispatch cases stay aligned. The repository validator compares that inventory only with SST routes targeting the monitor handler; the dedicated health handler is checked through its SST/OpenAPI/Bruno contract and existing handler tests.

Any remaining dynamic branch that cannot map safely to one route pattern is surfaced by tests or an explicit unsupported-pattern diagnostic, not guessed. This catches current handler-only archive, reactivate, maintenance, and escalation-state behavior while keeping the analysis bounded.

The first full comparison is allowed to fail on the repository's current drift. That failure is useful evidence, not a reason to weaken or defer the gate. Each discrepancy must be explicitly reconciled against merged or active OpenSpec requirements, and the handler/SST comparison must be green before `add-single-tenant-operator-authentication` task 3.1 refactors registrations through the protected-v1 helper. This ordering prevents the helper refactor from accidentally defining the route truth by omission.

Alternative considered: parse the current `switch` with regular expressions. Suffix checks and concatenated identifiers make this brittle and likely to produce false confidence. A full Go AST control-flow analyzer is disproportionate for this repository.

### 3. Extend the existing Bruno guard into one repository contract gate

The existing parser and fixture tests will be reused or factored into a deterministic script invoked by a root Makefile target. `make check-bruno` remains supported for Bruno conventions and is run explicitly in CI; the broader contract target adds OpenAPI, handler, and auth-metadata comparisons. Parsing uses repository-owned code and existing runtime facilities unless a small parser dependency is demonstrably necessary and approved under the applicable dependency policy.

Alternative considered: add unrelated validators per source. One normalized comparison model yields clearer diagnostics and avoids subtly different path rules.

### 4. Establish auth-ready metadata locally, then require it at cutover

The normalized route model supports `public` and `protected`. Before SST exposes auth classification, route and payload-contract coverage checks run and the validator makes no claim for non-health routes. Once infrastructure route metadata exists, every route must be classified and equivalent Bruno/OpenAPI metadata is required. Health is always asserted public. Protected OpenAPI operations use the configured security scheme; public operations explicitly override inherited security. Bruno requests record whether auth is absent or uses the collection's token variable without embedding a token.

Alternative considered: classify current routes as protected in advance. That would conflict with deployed behavior and the separate auth change. Ignoring auth entirely would allow documentation and smoke tests to drift as soon as auth lands.

### 5. Treat the local deterministic foundation as Phase 0

CI workflow steps call root Makefile targets rather than duplicate module lists. Before authentication cutover, the required command grouping covers Go format/vet/lint/tests/build as appropriate, dashboard format/lint/typecheck/tests/production build, infrastructure format/typecheck, `make check-bruno`, full SST/OpenAPI/Bruno/handler route drift, health envelope/documentation checks, and deterministic tests for portable profile/stage selection and production-stage rejection where those settings are consumed. Jobs may remain split by toolchain for useful parallelism, with dependency caches keyed by committed lockfiles and module sums.

This local foundation is the first internal milestone of this change and has no dependency on authentication. The roadmap can therefore place it in Phase 0 before the security cutover. The credentialed staging-auth smoke is a later milestone and must retain explicit dependencies on both `add-single-tenant-operator-authentication` and `standardize-stage-resource-lifecycle`.

Frozen pnpm installation and each root's `onlyBuiltDependencies` policy remain unchanged. OpenAPI's existing npm exception remains out of scope unless validation can parse YAML without installing that package root.

Alternative considered: one serial all-checks job. It simplifies YAML but increases feedback time and couples independent toolchains. Fine-grained path skipping is deferred because incorrect filters can skip required cross-surface drift checks.

### 6. Run cloud smoke locally only after auth and under an approved stage lifecycle

A local operator runs the smoke helper only after authentication and protected route metadata exist. The operator deliberately deploys the declared long-lived persistent staging environment through the lifecycle guard, then the helper reads the API URL from SST output. The helper never deploys or removes infrastructure.

A unique or per-run stage with any retained resource is forbidden because normal stack removal would accumulate orphaned, billable resources. Inputs and guards reject production stage names, and repository CI has no credential-bearing job or AWS credentials.

The smoke uses quiet HTTP requests and status/body assertions without shell tracing or header output. GitHub secrets are masked, token values are never written to artifacts, and the selected protected operation is read-only. Missing-token rejection is generated by API Gateway before Lambda invocation, so the smoke requires HTTP 401 but does not require the application response envelope or `reason.code`. Public health still requires the standard success envelope. A valid staging access token must pass Gateway authentication and reach the selected route's documented non-authentication outcome.

Alternative considered: deploy from GitHub Actions. This couples the repository to AWS identity and exposes credentials to a wider trigger surface. A shared long-lived staging environment is acceptable only when explicitly declared persistent and deliberately updated before assertions; it is not treated as disposable isolation.

### 7. Correct contracts and guidance as part of the gate rollout

OpenAPI will be reconciled to every SST route, service-scoped parameter names, standard envelopes, and auth metadata. README architecture will include both queues and the escalation runtime, and health examples will use the shared envelope. Contributor guidance will describe the synchronized route artifacts and commands. Completed-change archival is documented after implementation validation, never asserted as runtime behavior.

The validator may be introduced with a focused failing regression test that captures known drift, but required API assets and handler/SST wiring are reconciled in the same Phase 0 milestone before the gate becomes required CI and before auth route-helper refactoring. This preserves an honest red-to-green sequence without leaving required CI permanently red.

## Risks / Trade-offs

- [Independent route inventories can drift internally] -> Add focused fixture/unit tests that compare handler inventory with dispatch behavior and fail with source-oriented diagnostics.
- [OpenAPI reconciliation is large and may expose payload drift beyond routes] -> First establish exact route/security/envelope coverage, validate the document, and add representative schema tests without claiming complete semantic equivalence that is not checked.
- [Conditional auth support could become a permanent skip] -> Detect the presence of any SST auth metadata; once present, require complete classification and fail partial adoption.
- [SST deployment output may change] -> Encapsulate output discovery in repository scripts, test argument validation locally, and retain actionable logs without secrets.
- [Cleanup failure leaves ephemeral cloud resources] -> Permit ephemeral smoke only when lifecycle configuration disables retention, run always-on cleanup, verify zero residual resources, and report bounded manual cleanup by stage identifier.
- [Unique retained stages accumulate billable or security-sensitive resources] -> Forbid that combination; use either fully disposable ephemeral stages or one declared persistent staging environment with ownership and lifecycle controls.
- [Building all targets increases CI duration] -> Parallelize independent toolchains, use safe caches, avoid cloud work in ordinary CI, and do not duplicate the dashboard build across jobs.
- [Static handler coverage cannot prove runtime semantics] -> Scope the gate to route reachability and retain Go handler tests for behavior and envelopes.

## Migration Plan

1. Phase 0 foundation: add route normalization/parsing tests and a monitor-handler route inventory, then expose deterministic Makefile contract targets.
2. Reconcile current handler/SST routes and update Bruno and OpenAPI until the complete drift gate passes; do this before the authentication change refactors route registration.
3. Correct health envelopes and architecture/maintenance documentation, add portable profile/stage configuration checks, and run the dashboard production build and `make check-bruno`.
4. Expand ordinary pull-request and `main` CI to invoke all local release gates using frozen, trust-controlled installs, no AWS credentials, and bounded parallel work. This completes the Phase 0 milestone.
5. After authentication exists and `standardize-stage-resource-lifecycle` defines the approved non-production lifecycle, add the local staging auth smoke helper with production guards and secret-safe edge-aware assertions.
6. Run all repository checks locally, then manually exercise staging smoke with protected credentials against the current revision and verify the selected lifecycle invariant.

Rollback is workflow-safe: the local cloud smoke helper can be removed independently without changing production, while deterministic local gates remain. If a parser causes false positives, revert that parser/gate change rather than weakening source contracts or dependency trust controls.

## Open Questions

- Which read-only protected route and token acquisition mechanism will the authentication change expose for staging smoke? Resolve after that change supplies the route and direct-client flow; do not run a credentialed smoke early, hard-code credentials, or invent an auth contract here.
- Will `standardize-stage-resource-lifecycle` select fully disposable ephemeral smoke or declared persistent staging for this repository? Do not implement unique retained smoke stages while that dependency is unresolved.
- Does SST provide stable structured output for the deployed API URL in the pinned version, or should a small repository wrapper parse its JSON output? Confirm against the installed SST CLI before relying on the local helper.
