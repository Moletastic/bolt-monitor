import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const DASHBOARD_ROOT = resolve(__dirname, '..')
const LIB_DIR = resolve(DASHBOARD_ROOT, 'lib')
const IO_DIR = resolve(LIB_DIR, 'io')

/**
 * Recursively collect every `.ts`/`.tsx` file under `dir` except those under
 * `excludeDir` (matched by absolute path prefix).
 */
function collectSources(
  dir: string,
  excludeDir: string,
  acc: string[] = []
): string[] {
  let entries: string[]
  try {
    entries = readdirSync(dir)
  } catch {
    return acc
  }
  for (const entry of entries) {
    const abs = join(dir, entry)
    let st
    try {
      st = statSync(abs)
    } catch {
      continue
    }
    if (st.isDirectory()) {
      if (abs === excludeDir) continue
      collectSources(abs, excludeDir, acc)
      continue
    }
    if (!st.isFile()) continue
    if (!/\.(ts|tsx)$/.test(entry)) continue
    if (/\.test\.(ts|tsx)$/.test(entry)) continue
    if (entry.endsWith('.d.ts')) continue
    acc.push(abs)
  }
  return acc
}

/**
 * Strip block comments and line comments before checking for `try` so a
 * `try` mentioned in JSDoc doesn't trigger the guard.
 */
function stripComments(source: string): string {
  return source
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/\/\/.*$/gm, '')
}

/**
 * Find the byte index of the first `try` keyword outside a string. We use a
 * coarse regex after stripping comments; this is a guard, not a parser.
 */
function firstTryIndex(source: string): number {
  const stripped = stripComments(source)
  const match = /\btry\s*\{/.exec(stripped)
  return match ? match.index : -1
}

describe('try/catch boundary guard', () => {
  const sources = collectSources(LIB_DIR, IO_DIR)

  it('finds lib sources to scan', () => {
    expect(sources.length).toBeGreaterThan(0)
  })

  it('no source under lib/ except lib/io/ uses try { ... }', () => {
    const offenders: Array<{ file: string; index: number }> = []
    for (const file of sources) {
      const text = readFileSync(file, 'utf8')
      const idx = firstTryIndex(text)
      if (idx >= 0) {
        offenders.push({ file, index: idx })
      }
    }
    if (offenders.length > 0) {
      const message = offenders
        .map((o) => `  ${o.file.replace(DASHBOARD_ROOT + '/', '')}`)
        .join('\n')
      throw new Error(
        `try/catch is not allowed in lib/** except lib/io/**. Offenders:\n${message}`
      )
    }
  })
})
