## 1. Inventory And Collection Migration

- [x] 1.1 Extract current SST bootstrap method-and-path routes and classify existing Bruno requests as valid, stale, or missing.
- [x] 1.2 Reorganize the primary Bruno collection into service, monitor, incident, notification-channel, escalation-policy, admin, search, and health domains.
- [x] 1.3 Replace deprecated top-level monitor requests with service-scoped requests covering all service and nested monitor routes.
- [x] 1.4 Add missing requests for search, notification channels, escalation policies, service operations, service audit, incident activities, and all other wired routes.
- [x] 1.5 Remove stale requests that reference routes no longer wired in bootstrap.

## 2. Bruno Conventions

- [x] 2.1 Rename requests to the verb-and-resource convention.
- [x] 2.2 Align Bruno route variables with bootstrap parameter names such as `serviceId`, `monitorId`, `incidentId`, `channelId`, and `policyId`.
- [x] 2.3 Apply exactly one `domain:<domain>` tag and one `operation:<operation>` tag to every covered request.
- [x] 2.4 Add request docs covering purpose, setup, and expected result.
- [x] 2.5 Update collection and workspace documentation with current domains, flow, variables, and maintenance instructions.

## 3. Local Validation

- [x] 3.1 Implement deterministic static extraction of method-and-path routes from `infra/stacks/bootstrap.ts`.
- [x] 3.2 Implement Bruno request discovery across every collection and normalization of Bruno variables, methods, paths, and query strings.
- [x] 3.3 Validate route coverage, stale requests, route variable names, request names, required tags, and required docs.
- [x] 3.4 Add separate diagnostic reporting for OpenSpec routes that are absent from bootstrap wiring.
- [x] 3.5 Add `make check-bruno` and document expected output and local usage.
- [x] 3.6 Add validator tests for matching routes, missing routes, stale routes, metadata violations, query normalization, and spec/bootstrap mismatch reporting.

## 4. Governance Documentation And Verification

- [x] 4.1 Add high-level Bruno/API-contract synchronization principle to `CONSTITUTION.md`.
- [x] 4.2 Add operational Bruno conventions and route-change checklist to `AGENTS.md`.
- [x] 4.3 Run `make check-bruno` and existing relevant repository checks.
- [x] 4.4 Verify no primary collection request references deprecated top-level monitor routes.
