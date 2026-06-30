import { err, ok, type Result } from '@/lib/result'
import { ApiError, fromEnvelope, type ApiReasonPayload } from '@/lib/errors'

export { ApiError, fromEnvelope, type Result, type ApiReasonPayload, err, ok }

/**
 * Run a possibly-throwing async function and convert the outcome to a
 * `Result`. This is the only place in the dashboard where thrown exceptions
 * are caught — it is the I/O boundary helper and MUST only be used from
 * files under `apps/dashboard/lib/io/*`.
 *
 * If the function throws and `mapErr` is supplied, the thrown value is
 * narrowed by `mapErr` before becoming the `Err` payload. If `mapErr` is
 * omitted, the raw thrown value is carried as `unknown`.
 *
 * @param fn The async function to run.
 * @param mapErr Optional mapper from the thrown value to a typed error.
 * @returns `Ok(value)` on resolve, `Err(mapped)` on throw.
 */
export async function tryCatch<T, E = unknown>(
  fn: () => Promise<T>,
  mapErr?: (cause: unknown) => E
): Promise<Result<T, E>> {
  try {
    const value = await fn()
    return ok(value)
  } catch (cause) {
    return err(mapErr ? mapErr(cause) : (cause as E))
  }
}

/**
 * Server-action I/O boundary.
 *
 * The module that wraps a `fetch`/`apiRequest` call in `tryCatch` so the
 * action body can branch on `isOk`/`isErr` instead of catching. `apiRequest`
 * already throws an `ApiError` on non-2xx; we map the thrown error through
 * `fromEnvelope` re-classification (in case the I/O layer surfaces a string
 * code that didn't originate from the response envelope).
 */

export async function runServerAction<T>(
  fn: () => Promise<T>
): Promise<Result<T, ApiError>> {
  return tryCatch(fn, (cause) => {
    if (cause instanceof ApiError) {
      return cause
    }
    if (cause instanceof Error) {
      return new ApiError(
        // Fall back to INTERNAL for non-API errors; the user-visible message
        // is the original cause.
        'INTERNAL' as ApiError['code'],
        500,
        { message: cause.message }
      )
    }
    return new ApiError(
      'INTERNAL' as ApiError['code'],
      500,
      { message: String(cause) }
    )
  })
}

/**
 * Parse a JSON string into a `Result`. The I/O boundary is the only place
 * in the dashboard that may invoke `JSON.parse` inside a `try`; the
 * alternative — a raw try/catch in `actions.ts` — would violate the
 * `lib/**` rule in `lib/io/README.md`.
 */
export function parseJson<T>(raw: string): Result<T, string> {
  try {
    return ok(JSON.parse(raw) as T)
  } catch (cause) {
    return err(cause instanceof Error ? cause.message : String(cause))
  }
}
