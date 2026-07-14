# Constitution

Engineering principles for bolt-monitor. Principles only — concrete tooling, commands, and library picks live in `AGENTS.md`. Compliance is achieved through OpenSpec changes, not enforced inline.

Tag legend: `[G]` general · `[F]` frontend · `[B]` backend · `[A]` API contracts.

## General

1. **Spec-driven.** Behavior lives in OpenSpec first. No implementation without an active change. `[G]`
2. **Tests accompany change.** Behavior without coverage is incomplete. `[G]`
3. **Conventional commits, structured history.** Commit messages follow Conventional Commits. Header ≤ 80 chars. No body or description — terse one-liner only. `[G]`
4. **Docs mirror code.** Constitution, `AGENTS.md`, `README.md` reflect current state. Stale docs rot. `[G]`
5. **Smallest viable surface.** Build only what the active spec demands. No speculative features. `[G]`
6. **Service = source of truth.** The API owns the domain model. Other layers adapt, never redefine. `[G]`
7. **One language per concern.** Go for domain and services. TypeScript for UI and infra. Each stays in its lane. `[G]`
8. **Convention over configuration.** New code follows existing neighbors. `[G]`
9. **Immutable by default.** No mutation after construction. Readonly fields, value types, copies over edits. `[G]`
10. **Pure domain.** Business logic = pure functions. Side effects live at edges only. `[G]`
11. **Enums over magic values.** All domain literals named. No raw strings or numbers in conditionals. `[G]`
12. **Result over exceptions.** Return typed `Result<T, E>`. No try/catch for control flow; TypeScript catch sites live only in documented I/O boundaries. See `openspec/specs/ts-result-and-no-any/spec.md`. `[G]`
13. **Facade over direct dependency.** Domain code talks to internal interfaces, never raw libraries. Implemented for AWS access by `openspec/changes/code-patterns-foundation` through `shared/aws`. `[G]`
14. **Safety by default.** Destructive actions confirm. Secrets never in code or logs. `[G]`
15. **Input validation at every boundary.** Reject early with explicit, structured detail. `[G]`
16. **Fail loud, fail early.** Missing config = explicit error. No silent fallbacks. `[G]`
17. **Reproducible deploy.** Single tool, single profile, single stage per environment. `[G]`
18. **Explicit dependencies.** Lockfiles committed. Manager pinned. Allowlists visible. `[G]`
19. **Automated gates.** Lint, typecheck, test, format, build run in CI. Manual review is the exception. `[G]`
20. **Public APIs documented.** JSDoc in TypeScript, godoc in Go. `[G]`
21. **Manual API contracts stay synchronized.** Bruno collections cover every exposed API route and remain aligned with the service-owned contract. `[G]`

## Frontend

21. **Server-side truth.** UI reads authoritative state via server fetch. Client never mirrors canonical data. `[F]`
22. **No native `Date`.** Use an immutable date utility. Predictable, tree-shakeable, testable. Implemented by `openspec/changes/code-patterns-foundation` through `date-fns` and the dashboard clock wrapper. `[F]`
23. **No `any` in TypeScript.** Use `unknown` and narrow. Strict types at every boundary; dashboard ESLint enforces this. See `openspec/specs/ts-result-and-no-any/spec.md`. `[F]`

## Backend

24. **Shared domain, no duplication.** Canonical Go modules live once in `shared/`. Services consume, never fork. `[B]`
25. **Rules pattern in Go backend.** Business rules = composable predicates. No scattered branching. Implemented by `openspec/changes/code-patterns-foundation` through `shared/rules`. `[B]`
26. **Centralized error codes.** Go = registry. Stable, documented, machine-readable. `[B]`
27. **Side-effectful boundaries explicit.** Lambda handlers are edges. Domain stays pure. `[B]`

## API Contracts

28. **Uniform response envelope.** All entry points return `{ status, data?, reason?, message?, pagination? }`. Go = struct, TypeScript = class. `[A]`
29. **`status` enum.** `success | error`. No string literals. `[A]`
30. **`message` is success-path only.** Failures surface detail through `reason`. `[A]`
31. **Machine-readable error codes.** `reason = { code, details }`. `details: Record<string, unknown>` in TypeScript, `map[string]any` in Go. `[A]`
32. **Centralized error codes (API).** Go = registry. TypeScript = enum. Shared vocabulary across runtimes. `[A]`
33. **Pagination object.** `{ page, size, total, items }`. Present only when applicable. `[A]`
34. **API versioning.** Routes carry a version (`/api/v1/...`). Breaking change = new version, never mutation. `[A]`
35. **Idempotent mutations.** `POST`, `PUT`, `DELETE` safe to retry. Server tracks or client supplies a key. `[A]`
36. **Backwards compatibility by default.** Additive changes only. Removal = new spec, new version. `[A]`

## Operational Principles — FinOps (AWS)

37. **Right-size by default.** Lambda memory and timeout measured, not maxed. DynamoDB capacity mode matches traffic pattern. Document the choice in spec. `[G]`
38. **No idle resources.** Stage stacks cleaned up. No orphaned volumes, ENIs, or NAT gateways. Dev environments torn down on idle. `[G]`
39. **Tag and attribute.** All SST resources tagged (`service`, `stage`, `owner`). Cost queryable per service in Cost Explorer. `[G]`
40. **Data transfer is real cost.** Cross-AZ and cross-region traffic budgeted. Single-region by default; multi-region requires spec justification. `[G]`
41. **Cost in spec review.** Material AWS additions (new table, new schedule, new function) carry a cost note in OpenSpec. `[G]`
