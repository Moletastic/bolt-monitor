## Overview

TypeScript expresses fallible operations through a discriminated-union `Result<T, E>`, never through `try`/`catch` for control flow. Errors carry a machine-readable code shared with the Go backend. ESLint enforces the rule at compile time and at lint time.

## Requirements

### Requirement: `Result<T, E>` utility

A `Result<T, E>` type SHALL be exported from `apps/dashboard/lib/result.ts`. The type SHALL be a discriminated union: `{ ok: true; value: T } | { ok: false; error: E }`. Helpers SHALL include `ok`, `err`, `isOk`, `isErr`, `map`, `flatMap`, `match`, and `unwrap`. The `tryCatch` helper SHALL live under `apps/dashboard/lib/io/` because it necessarily contains `try`/`catch`. Every helper SHALL have JSDoc explaining intent and contract.

### Requirement: `try`/`catch` is allowed only at I/O boundary

Inside `apps/dashboard/lib/**`, `try`/`catch` SHALL be permitted only inside files under `apps/dashboard/lib/io/`. ESLint SHALL enforce this via a `no-restricted-syntax` rule scoped to `lib/**`. The I/O boundary is the seam where thrown exceptions from `fetch`, `await`, `JSON.parse`, or third-party SDKs are caught and converted to `Result`. Business logic and server actions under `lib/**` SHALL NOT use `try`/`catch` for control flow.

### Requirement: No `any` in TypeScript

`any` SHALL be a lint error. The `@typescript-eslint/no-explicit-any` rule and the `no-unsafe-*` family SHALL be enabled at `error` level. Use of `any` requires an explicit `// eslint-disable-next-line` with a written reason; the suppression SHALL be reviewed on merge.

### Requirement: `unknown` is narrowed, not cast

Code that consumes `unknown` SHALL narrow it (typeof checks, type guards, `zod` parse, schema validation) before use. `as unknown as T` casts SHALL be flagged by lint; preferred alternative is a `zod` schema or a custom type guard.

### Requirement: Error codes mirror the Go registry

`ApiErrorCode` enum in `apps/dashboard/lib/errors.ts` SHALL have a one-to-one mapping with codes registered in `shared/errors/code.go`. A sync test SHALL fail CI if a Go code lacks a TS counterpart or vice versa. The enum value SHALL be the same string as the Go code.

### Requirement: `ApiError` is structured

The dashboard's error representation SHALL be an `ApiError` class with `code: ApiErrorCode`, `details: Record<string, unknown>`, optional `message: string`. It SHALL implement the `Error` interface for stack-trace and `instanceof` compatibility.

### Requirement: Server actions use `Result` at API boundaries

Server actions in `apps/dashboard/lib/actions.ts` SHALL call the I/O boundary (`runServerAction` / `tryCatch`) for API calls, branch on `isOk` / `isErr`, and surface typed `ApiError` messages before redirecting. The existing navigation-first form convention SHALL be preserved by this change. Returning action state directly to UI components is covered by `dashboard-ui-action-state-results`.

### Requirement: `tryCatch` is the only allowed escape

When a server action must call into the I/O boundary, it SHALL go through `tryCatch` (the helper that catches and maps to `Result`). Direct `try { ... } catch { ... }` in `actions.ts` SHALL be a lint error.
