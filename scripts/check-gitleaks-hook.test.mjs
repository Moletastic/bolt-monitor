import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const root = new URL('..', import.meta.url)
const lefthook = readFileSync(new URL('lefthook.yml', root), 'utf8')

test('pre-commit scans staged content with redacted Gitleaks output', () => {
  assert.match(lefthook, /name: scan staged credentials/)
  assert.match(lefthook, /gitleaks protect --staged --redact --no-banner/)
})

test('pre-commit fails when Gitleaks is unavailable with installation guidance', () => {
  assert.match(lefthook, /command -v gitleaks/)
  assert.match(
    lefthook,
    /No credential leak analysis ran: install Gitleaks and retry\./
  )
  assert.match(lefthook, /command -v gitleaks[\s\S]*?exit 1/)
})
