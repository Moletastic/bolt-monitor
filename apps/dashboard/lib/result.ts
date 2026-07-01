/**
 * Result<T, E> — a discriminated union for fallible operations.
 *
 * TypeScript lacks a native (T, error) tuple, so this utility is the
 * runtime expression of the constitution's "no try/catch for control flow"
 * rule. The companion {@link tryCatch} helper turns thrown exceptions at the
 * I/O boundary into a `Result`. Business logic, server actions, and UI
 * components branch on {@link isOk} / {@link isErr} instead of catching.
 *
 * Error contract: a `Result` always carries the discriminant literal `ok`.
 * Type guards narrow the union so consumers see `value` on the `Ok` arm and
 * `error` on the `Err` arm — no runtime tag check is required.
 */

export type Ok<T> = { readonly ok: true; readonly value: T }
export type Err<E> = { readonly ok: false; readonly error: E }

export type Result<T, E> = Ok<T> | Err<E>

/**
 * Wrap a successful value.
 *
 * @param value The value to carry.
 * @returns An `Ok<T>` ready for return.
 */
export function ok<T>(value: T): Ok<T> {
  return { ok: true, value }
}

/**
 * Wrap an error value.
 *
 * @param error The error to carry.
 * @returns An `Err<E>` ready for return.
 */
export function err<E>(error: E): Err<E> {
  return { ok: false, error }
}

/**
 * Narrow a `Result` to its `Ok` arm.
 *
 * @param r The result to test.
 * @returns `true` when `r` is `Ok<T>`; the guard narrows the type.
 */
export function isOk<T, E>(r: Result<T, E>): r is Ok<T> {
  return r.ok
}

/**
 * Narrow a `Result` to its `Err` arm.
 *
 * @param r The result to test.
 * @returns `true` when `r` is `Err<E>`; the guard narrows the type.
 */
export function isErr<T, E>(r: Result<T, E>): r is Err<E> {
  return !r.ok
}

/**
 * Transform the success value of a `Result`. Short-circuits on `Err`.
 *
 * @param r The result to map over.
 * @param f The function applied to the value when `r` is `Ok`.
 * @returns A `Result<U, E>` — `Ok(f(value))` on success, `Err` unchanged.
 */
export function map<T, U, E>(r: Result<T, E>, f: (value: T) => U): Result<U, E> {
  return r.ok ? ok(f(r.value)) : r
}

/**
 * Chain another fallible operation. Short-circuits on `Err`.
 *
 * @param r The result to chain from.
 * @param f The function returning a new `Result` from the value.
 * @returns The chained `Result<U, E>`.
 */
export function flatMap<T, U, E>(r: Result<T, E>, f: (value: T) => Result<U, E>): Result<U, E> {
  return r.ok ? f(r.value) : r
}

/**
 * Exhaustively fold a `Result` into a single value.
 *
 * @param r The result to fold.
 * @param onOk Handler invoked when `r` is `Ok`.
 * @param onErr Handler invoked when `r` is `Err`.
 * @returns The value returned by the chosen handler.
 */
export function match<T, E, U>(r: Result<T, E>, onOk: (value: T) => U, onErr: (error: E) => U): U {
  return r.ok ? onOk(r.value) : onErr(r.error)
}

/**
 * Extract the success value or throw the carried error.
 *
 * `unwrap` is appropriate only at the very edges of a program: in tests, or
 * after an exhaustive {@link match} that has already established the
 * discriminant. Server actions and UI components MUST branch on
 * {@link isOk} / {@link isErr} instead of calling `unwrap`.
 *
 * @param r The result to unwrap.
 * @returns The carried value when `r` is `Ok`.
 * @throws The carried error when `r` is `Err`.
 */
export function unwrap<T, E>(r: Result<T, E>): T {
  if (r.ok) {
    return r.value
  }
  throw r.error
}
