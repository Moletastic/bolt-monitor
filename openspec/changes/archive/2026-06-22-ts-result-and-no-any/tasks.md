## 1. Result utility

- [x] 1.1 Create `apps/dashboard/lib/result.ts` with discriminated-union `Result<T, E> = { ok: true; value: T } | { ok: false; error: E }`.
- [x] 1.2 Add factory functions `ok<T>(value: T): Ok<T>` and `err<E>(error: E): Err<E>`.
- [x] 1.3 Add `isOk<T, E>(r: Result<T, E>): r is Ok<T>` and `isErr<T, E>(r: Result<T, E>): r is Err<E>` type guards.
- [x] 1.4 Add `map<T, U, E>(r: Result<T, E>, f: (v: T) => U): Result<U, E>`, `flatMap<T, U, E>(r, f)`, `match<T, E, U>(r, onOk, onErr)`, `unwrap<T, E>(r): T` (throws on `Err`).
- [x] 1.5 Add `tryCatch<T, E>(fn: () => Promise<T>, mapErr?: (e: unknown) => E): Promise<Result<T, E>>` — the I/O boundary helper, only used in `apps/dashboard/lib/io/*`.
- [x] 1.6 Document each helper with JSDoc explaining intent, error narrow contract, and when `unwrap` is appropriate (tests, exhaustive match).

## 2. Error code enum

- [x] 2.1 Create `apps/dashboard/lib/errors.ts` with `ApiErrorCode` enum mirroring the 21-code Go registry in `shared/errors/code.go`.
- [x] 2.2 Add `ApiError` class with `code: ApiErrorCode`, `details: Record<string, unknown>`, `message?: string`. Implements `Error` interface for stack-trace compatibility.
- [x] 2.3 Add `fromEnvelope(reason: { code: string; details: Record<string, unknown> }): ApiError` factory that maps string code → enum (throws if code is unknown — signals drift).
- [x] 2.4 Add a test asserting every Go code in `shared/errors/code.go` has a matching `ApiErrorCode` entry (sync test, fails CI if drift).

## 3. ESLint configuration

- [x] 3.1 Add `@typescript-eslint/no-explicit-any: error` to `apps/dashboard/.eslintrc.js`.
- [x] 3.2 Add `@typescript-eslint/no-unsafe-argument`, `no-unsafe-assignment`, `no-unsafe-call`, `no-unsafe-member-access`, `no-unsafe-return` at `error` level to enforce `unknown` narrowing.
- [x] 3.3 Add `no-restricted-syntax` rule flagging `TryStatement` AST nodes with a `TryStatement > CatchClause > BlockStatement` pattern, scoped to all files except `apps/dashboard/lib/io/**`.
- [x] 3.4 Document the I/O boundary convention in `apps/dashboard/lib/io/README.md` with one example per allowed file.
- [x] 3.5 Run `pnpm lint` and fix every existing violation. Track and resolve before merging.

## 4. Refactor dashboard actions

- [x] 4.1 Refactor `try`/`catch` sites in `apps/dashboard/lib/actions.ts` to call `runServerAction` at API boundaries and branch on `Result<T, ApiError>` before redirecting.
- [x] 4.2 Preserve the existing navigation-first `<form action={...}>` server-action convention; defer returned action state and `useActionState` component branching to `dashboard-ui-action-state-results`.
- [x] 4.3 Update server-action error redirects to use `messageFor(error)`, rendering `error.message` if present, else humanized `error.code`.

## 5. Tests

- [x] 5.1 Add `apps/dashboard/lib/result.test.ts` covering all helpers, including `unwrap` throwing, `map` short-circuiting on `Err`, `flatMap` chain.
- [x] 5.2 Add `apps/dashboard/lib/errors.test.ts` covering `fromEnvelope` happy path, unknown-code throw, code-drift sync.
- [x] 5.3 Add a server-action test asserting `try`/`catch` is not used outside `lib/io/*` (lint + grep guard).
- [x] 5.4 Run `make check-dashboard`, `make lint-dashboard`, `make build-dashboard`. All pass.

## 6. Documentation

- [x] 6.1 Update `AGENTS.md` TypeScript error-handling subsection with `Result` usage examples.
- [x] 6.2 Document the I/O boundary convention in `apps/dashboard/lib/io/README.md`.
- [x] 6.3 Update `CONSTITUTION.md` §12 (TS half) and §23 (`no any`) cross-references to point at this change's spec.
