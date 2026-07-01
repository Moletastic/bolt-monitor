# `lib/io/` — the I/O boundary

This directory is the **only** place in `apps/dashboard/lib/` that may use
`try { ... } catch { ... }`. The ESLint rule `no-restricted-syntax` blocks
`TryStatement` everywhere else under `lib/**`; the `lib/io/**` glob is
explicitly allowlisted.

## Why an I/O boundary?

TypeScript's `try`/`catch` is the only way to handle exceptions thrown by
`fetch`, third-party SDKs, and the Node.js runtime. Catching those thrown
exceptions and converting them into a `Result<T, E>` is the seam between
"the world can throw" and "our code does not". Once the boundary is crossed,
business logic, server actions, and UI components branch on
`isOk` / `isErr` instead of catching.

## Files in this directory

| File               | Purpose                                                                                                                                          |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `server-action.ts` | `runServerAction` — wraps an async function in `tryCatch` and maps thrown values to `ApiError` so a server action returns `Result<T, ApiError>`. |

## Allowed patterns

```ts
// lib/io/server-action.ts (allowed)
import { runServerAction } from '@/lib/io/server-action'

export async function myAction(): Promise<Result<MyValue, ApiError>> {
  return runServerAction(async () => {
    return await apiRequest<MyValue>('/api/v1/example')
  })
}
```

```ts
// lib/actions.ts (NOT allowed — use runServerAction instead)
try {
  await apiRequest<MyValue>('/api/v1/example')
} catch (error) {
  redirect(`?error=${(error as Error).message}`)
}
```

The second example is exactly what the lint rule rejects. `actions.ts` must
go through `runServerAction` (or the lower-level `tryCatch` from
`@/lib/result`) so the boundary stays in one place.

## Why not let `try`/`catch` be free?

Constitution §12 (TypeScript half) requires that fallible operations return
`Result<T, E>` rather than throw. Allowing `try`/`catch` everywhere would
let exceptions leak past the boundary as control flow — the very thing
`Result` exists to prevent. The lint rule is the structural enforcement.

## Adding a new I/O helper

If you need a new `try`/`catch` site, add the file to this directory. The
glob is `apps/dashboard/lib/io/**`, so any new file at any depth is covered.
Do **not** add `try`/`catch` to a file outside this directory even if the
catch is small — it will fail CI.
