## Overview

Dashboard mutation flows that need inline feedback express server-action failures through typed action state derived from `Result<T, ApiError>`. Navigation-first forms may continue to redirect, but forms converted by this change SHALL branch on typed success/error state in the component.

## Requirements

### Requirement: Action state uses typed errors

Converted server actions SHALL expose a serializable action state whose error arm carries `code: ApiErrorCode`, `details: Record<string, unknown>`, and optional `message: string`. Components SHALL NOT parse free-form query strings when rendering these returned errors.

### Requirement: Components branch on result state

Converted components SHALL branch on `isOk` / `isErr` or an equivalent serializable discriminator derived from `Result<T, ApiError>`. Error UI SHALL render `messageFor(error)` or the serialized equivalent, and MAY expose `error.code` / `error.details` for operator debugging when useful.

### Requirement: Redirect behavior is preserved unless explicitly changed

Flows that currently navigate after success SHALL preserve that navigation unless the task explicitly changes the UX. `redirect()` MUST remain outside `runServerAction` / `tryCatch` so Next.js redirect signals are not captured as errors.

### Requirement: Toast and alert copy uses typed error messages

Toast and alert components used by converted flows SHALL display `error.message` when present, else a humanized form of `error.code`. Components SHALL NOT surface raw enum strings as primary user-facing copy.

### Requirement: Coverage for success and error branches

Converted flows SHALL have tests or guard coverage for both successful and failed action-state branches, including unknown API error-code handling at the `fromEnvelope` seam.
