## Why

The constitution's principle 23 forbids `any` in TypeScript, principle 12 (TS half) requires a `Result<T, E>` style over `try`/`catch` for control flow, and the dashboard had repeated `try { ... } catch (e) { redirect('?error=' + e.message) }` blocks in `apps/dashboard/lib/actions.ts`. TypeScript has no native `(T, error)` tuple, so a `Result` is the runtime-appropriate expression of the principle. This change introduces a typed `Result<T, E>` utility, bans `any` via ESLint, mirrors the Go error-code registry as a TypeScript enum, and introduces a dashboard I/O boundary that converts thrown API failures into `Result<T, ApiError>` while preserving the current navigation-first form convention.

## What Changes

- Add a TypeScript `Result<T, E>` utility in `apps/dashboard/lib/result.ts` with `ok`, `err`, `map`, `flatMap`, `match`, `unwrap`, `isOk`, `isErr` helpers.
- Add a TypeScript error-code enum in `apps/dashboard/lib/errors.ts` mirroring the codes registered in `shared/errors/code.go`.
- Add ESLint rules: ban `any` (and `as any`); ban `try`/`catch` under `apps/dashboard/lib/**` except files explicitly marked as I/O boundaries under `apps/dashboard/lib/io/**`.
- Refactor `apps/dashboard/lib/actions.ts` to call `runServerAction` at API boundaries and branch on `Result<T, ApiError>` before redirecting.

## Capabilities

### New Capabilities

- `ts-result-and-no-any`: TypeScript `Result<T, E>` utility, error-code enum mirroring Go registry, ESLint rules banning `any` and `try`/`catch` outside I/O boundaries, and the dashboard I/O-boundary refactor that adopts both.

### Modified Capabilities

- `dashboard-web-app`: server actions preserve the existing navigation-first form convention but convert API failures through `runServerAction` and surface typed `ApiError` messages before redirecting. Full returned-action-state UI conversion is covered by `dashboard-ui-action-state-results`.

## Impact

- New TypeScript file `apps/dashboard/lib/result.ts` with `Result<T, E>`, `Ok<T>`, `Err<E>`, and helper functions.
- New TypeScript file `apps/dashboard/lib/errors.ts` with `ApiErrorCode` enum mirroring `shared/errors/code.go`.
- Modified `apps/dashboard/.eslintrc.js` to add `@typescript-eslint/no-explicit-any: error`, the `no-unsafe-*` family at `error`, and a `no-restricted-syntax` AST pattern flagging `try`/`catch` under `apps/dashboard/lib/**` except `apps/dashboard/lib/io/**`.
- New directory `apps/dashboard/lib/io/` marking the I/O boundary files (`fetch.ts`, `server-action.ts`) where `try`/`catch` is allowed.
- Modified `apps/dashboard/lib/actions.ts` to use `runServerAction` and remove local `try`/`catch` repetitions under `lib/**`.
- Modified `apps/dashboard/lib/api.ts` to map `ApiResponse.reason.code` to `ApiErrorCode` and throw a typed `ApiError` (or return `Err`).
- New tests in `apps/dashboard/lib/result.test.ts` and `apps/dashboard/lib/errors.test.ts`.
- ESLint config updated; existing `any` usages in dashboard code base fixed (search + replace with `unknown` + narrowing).

## Out of Scope

- Go error handling and code registry (covered by `go-error-handling-typed-codes`).
- Response envelope struct/class (covered by `api-response-envelope`).
- Facade pattern, Rules pattern, date-fns (covered by `code-patterns-foundation`).
- FinOps tagging and right-sizing (deferred).
- Returned-action-state UI conversion for server actions and client components (covered by `dashboard-ui-action-state-results`).
