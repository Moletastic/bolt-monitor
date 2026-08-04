## Context

Root validation scripts provide deterministic API-contract, Bruno, OpenAPI, install-trust, Makefile safety, and deployment-lifecycle checks. The current Makefile calls selected root tests from several targets, leaving parser tests unexecuted while running install-trust tests twice. A manual auth-route checker repeats assertions owned by infrastructure tests. The auth-cutover prerequisite checker also inspects implementation text with regular expressions rather than testing deployment behavior.

## Goals / Non-Goals

**Goals:**

- Preserve deterministic, local API-contract and release-gate coverage while reducing duplicate validation work.
- Make every remaining root script test run once from the infrastructure test target.
- Remove redundant auth-route validation in favor of existing infrastructure behavior tests.
- Prefer behavior-level lifecycle tests over source-text matching when equivalent coverage can be established.
- Reduce repeated filesystem traversal, repository file reads, and CLI error reporting without obscuring rule ownership.

**Non-Goals:**

- Replace literal route inventories or handwritten route parsers with a new parser dependency.
- Change API routes, OpenAPI semantics, Bruno collection conventions, deployment behavior, or CI job topology.
- Remove a lifecycle invariant without a behavior-level test proving it.

## Decisions

### Run root script tests as one explicit glob

`test-infra` will call `node --test scripts/*.test.mjs` once. This makes test discovery visible at the root command surface and removes hand-maintained, overlapping test lists. Production validators remain individual Makefile targets because contributors need focused commands and CI diagnostics need named validation surfaces.

Alternative: retain selective lists and add missing files. Rejected because future tests can become dormant again and list maintenance caused the current duplication.

### Delete duplicated auth-route checker

Remove `check-auth-routes` and its test and Make target. `infra/auth-infrastructure.test.ts` already verifies the JWT authorizer, required scope, protected versioned registration, and public health route from the actual stack source.

Alternative: add the checker to CI. Rejected because it duplicates coverage and adds another source-shape contract to maintain.

### Gate lifecycle-check removal on behavior coverage

Map each prerequisite asserted by `check-auth-cutover-prerequisites` to a behavior-level test in infrastructure. Add narrowly scoped infrastructure tests for any missing invariant. Delete the checker and its test only when the mapping is complete; otherwise retain only unmatched assertions with negative fixtures.

Alternative: retain all regex assertions. Rejected because benign structural refactors can fail validation without changing lifecycle behavior.

### Share only mechanical helpers

Extract shared helpers for repository-root file reads, recursive matching file discovery, and standard CLI error reporting. Keep route parsers and validation rules in their current domain modules so their source formats and diagnostics remain clear.

Alternative: create a generic validation framework. Rejected because it would add indirection without reducing domain complexity.

## Risks / Trade-offs

- Glob test discovery can run an unintended test file → Root `scripts/` remains repository-owned and all tests must stay deterministic and network-free.
- Removing source-shape checks can leave an invariant uncovered → Delete only after behavior-test mapping and negative tests pass.
- Shared helpers can hide useful diagnostics → Helpers accept errors and paths only; rule messages remain in validators.
- Literal route parsers remain format-sensitive → Preserve fixture coverage and documented literal declaration conventions.

## Migration Plan

1. Establish behavior-test coverage for every lifecycle invariant and remove duplicated auth-route validation.
2. Consolidate root script test execution and remove repeated lists.
3. Extract mechanical helpers without changing validator output.
4. Run focused root validators and the complete infrastructure test target locally and in existing CI.
5. Rollback consists of restoring removed target/script files; no deployed-state migration exists.
