import { describe, expect, it } from 'vitest'

import { err, flatMap, isErr, isOk, map, match, ok, Result, unwrap } from './result'
import { tryCatch } from './io/server-action'

describe('result', () => {
  it('ok() builds an Ok<T> and isOk() narrows', () => {
    const r = ok(42)
    expect(r).toEqual({ ok: true, value: 42 })
    if (isOk(r)) {
      expect(r.value).toBe(42)
    } else {
      throw new Error('isOk should narrow')
    }
  })

  it('err() builds an Err<E> and isErr() narrows', () => {
    const r: Result<number, string> = err('boom')
    expect(r).toEqual({ ok: false, error: 'boom' })
    if (isErr(r)) {
      expect(r.error).toBe('boom')
    } else {
      throw new Error('isErr should narrow')
    }
  })

  it('map() transforms an Ok value and short-circuits on Err', () => {
    const doubled = map(ok(2), (n) => n * 2)
    expect(doubled).toEqual({ ok: true, value: 4 })

    const propagated = map(err<string>('bad'), (n: number) => n * 2)
    expect(propagated).toEqual({ ok: false, error: 'bad' })
  })

  it('flatMap() chains a fallible operation and short-circuits on Err', () => {
    const chained = flatMap(ok(2), (n) => (n > 0 ? ok(n * 3) : err('negative' as const)))
    expect(chained).toEqual({ ok: true, value: 6 })

    const bad = flatMap(err<string>('start'), (n: number) =>
      n > 0 ? ok(n * 3) : err('negative' as const)
    )
    expect(bad).toEqual({ ok: false, error: 'start' })

    const chainErr = flatMap(ok(-1), (n) => (n > 0 ? ok(n * 3) : err('negative' as const)))
    expect(chainErr).toEqual({ ok: false, error: 'negative' })
  })

  it('match() folds a Result into a single value', () => {
    const onOk = match(ok('hi'), (v) => v.toUpperCase(), () => 'NOPE')
    expect(onOk).toBe('HI')
    const onErr = match(err('bad'), () => 'ok', (e) => `err:${e}`)
    expect(onErr).toBe('err:bad')
  })

  it('unwrap() returns the value on Ok and throws on Err', () => {
    expect(unwrap(ok('v'))).toBe('v')

    const sentinel = new Error('sentinel')
    expect(() => unwrap(err(sentinel))).toThrow(sentinel)
  })

  it('tryCatch() wraps a resolved Promise into Ok', async () => {
    const r = await tryCatch(async () => 7)
    expect(r).toEqual({ ok: true, value: 7 })
  })

  it('tryCatch() wraps a rejected Promise into Err with the raw cause when no mapper', async () => {
    const cause = new Error('io blew up')
    const r = await tryCatch(async () => {
      throw cause
    })
    expect(r.ok).toBe(false)
    if (!r.ok) {
      expect(r.error).toBe(cause)
    }
  })

  it('tryCatch() runs the mapper on the thrown cause', async () => {
    const r = await tryCatch<string, { kind: 'wrapped'; message: string }>(
      async () => {
        throw new Error('raw')
      },
      (e) => {
        if (e instanceof Error) {
          return { kind: 'wrapped' as const, message: e.message }
        }
        return { kind: 'wrapped' as const, message: 'unknown' }
      }
    )
    expect(r.ok).toBe(false)
    if (!r.ok) {
      expect(r.error).toEqual({ kind: 'wrapped', message: 'raw' })
    }
  })
})
