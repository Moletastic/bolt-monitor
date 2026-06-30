import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

import { ApiError, ApiErrorCode, fromEnvelope, humanize, isApiErrorCode } from './errors'

const REPO_ROOT = resolve(__dirname, '..', '..', '..')

/**
 * Parse the `Code = "..."` literals out of `shared/errors/code.go`.
 *
 * The Go file is the source of truth for the registry; this parser is
 * intentionally tolerant — it matches both `const (...)` and single-line
 * declarations so a future refactor of the Go side doesn't silently
 * disable the drift test.
 */
function parseGoCodeLiterals(source: string): string[] {
  const literals: string[] = []
  const re = /Code\w*\s+Code\s*=\s*"([^"]+)"/g
  let match: RegExpExecArray | null
  while ((match = re.exec(source)) !== null) {
    literals.push(match[1] as string)
  }
  return literals
}

describe('errors', () => {
  describe('ApiErrorCode', () => {
    it('isApiErrorCode() narrows known codes', () => {
      expect(isApiErrorCode('SERVICE_NOT_FOUND')).toBe(true)
      expect(isApiErrorCode('NOT_A_REAL_CODE')).toBe(false)
    })

    it('humanize() returns a non-empty string for every code', () => {
      for (const code of Object.values(ApiErrorCode)) {
        expect(humanize(code).length).toBeGreaterThan(0)
      }
    })
  })

  describe('fromEnvelope()', () => {
    it('builds a typed ApiError from a known reason', () => {
      const err = fromEnvelope(
        { code: ApiErrorCode.ServiceNotFound, details: { id: 'svc-1' } },
        404
      )
      expect(err).toBeInstanceOf(ApiError)
      expect(err.code).toBe('SERVICE_NOT_FOUND')
      expect(err.status).toBe(404)
      expect(err.httpStatus).toBe(404)
      expect(err.details).toEqual({ id: 'svc-1' })
      expect(err.message).toBe('SERVICE_NOT_FOUND')
    })

    it('attaches the optional human-readable message when present', () => {
      const err = fromEnvelope(
        { code: ApiErrorCode.ValidationFailed, details: { detail: 'bad' } },
        400
      )
      const withMessage = new ApiError(err.code, err.status, err.details, 'name is required')
      expect(withMessage.message).toBe('name is required')
    })

    it('throws on an unknown code so drift surfaces immediately', () => {
      expect(() =>
        fromEnvelope({ code: 'TOTALLY_NEW_CODE', details: {} }, 500)
      ).toThrow(/TOTALLY_NEW_CODE/)
    })

    it('uses the fallback status when not provided', () => {
      const err = fromEnvelope({
        code: ApiErrorCode.Internal,
        details: {},
      })
      expect(err.status).toBe(500)
    })
  })

  describe('drift sync with shared/errors/code.go', () => {
    const goSource = readFileSync(
      resolve(REPO_ROOT, 'shared', 'errors', 'code.go'),
      'utf8'
    )
    const goCodes = parseGoCodeLiterals(goSource)
    const tsCodes = Object.values(ApiErrorCode) as string[]

    it('TS enum values exactly match the Go registry', () => {
      expect(tsCodes.sort()).toEqual([...goCodes].sort())
    })

    it('every Go code is mapped to a known TS code', () => {
      for (const code of goCodes) {
        expect(isApiErrorCode(code)).toBe(true)
      }
    })

    it('every TS code is present in the Go registry', () => {
      for (const code of tsCodes) {
        expect(goCodes).toContain(code)
      }
    })
  })
})
