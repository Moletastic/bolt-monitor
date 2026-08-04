## 1. Remove Duplicate Validation

- [x] 1.1 Confirm infrastructure auth-boundary tests cover public health, versioned-route authorization, JWT authorizer, and required scope.
- [x] 1.2 Remove `check-auth-routes`, its unit test, and its Makefile target.
- [x] 1.3 Run the remaining focused auth and infrastructure tests after removal.

## 2. Consolidate Root Script Tests

- [x] 2.1 Change `test-infra` to execute every root `scripts/*.test.mjs` test once.
- [x] 2.2 Remove overlapping root-script test lists from focused Makefile targets while retaining their production checker commands.
- [x] 2.3 Verify CI continues to call only root Makefile targets and runs no duplicate root script tests.

## 3. Replace Brittle Lifecycle Checks

- [x] 3.1 Map each auth cutover prerequisite regex assertion to an existing behavior-level infrastructure test.
- [x] 3.2 Add focused positive and negative infrastructure tests for lifecycle invariants not already covered.
- [x] 3.3 Remove the auth cutover prerequisite checker, its test, and its Makefile/pre-cutover references after behavior coverage completed.

## 4. Extract Mechanical Helpers

- [x] 4.1 Extract shared repository file-read and recursive file-discovery helpers used by root validators.
- [x] 4.2 Extract shared CLI error reporting without changing validator diagnostics or exit behavior.
- [x] 4.3 Refactor Bruno and OpenSpec validators to use the helpers while retaining local parsing and policy rules.

## 5. Verify

- [x] 5.1 Run `make test-infra` and confirm every root script test passes once.
- [x] 5.2 Run `make check-api-contract check-bruno` and confirm unchanged contract diagnostics and route coverage.
- [x] 5.3 Run `make check-infra format-infra-check` and relevant focused infrastructure tests.
