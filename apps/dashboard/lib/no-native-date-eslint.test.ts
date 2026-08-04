import { Linter } from 'eslint'
import { createRequire } from 'node:module'
import { describe, expect, it } from 'vitest'

const require = createRequire(import.meta.url)
const eslintConfig = require('../.eslintrc.js') as {
  rules: Record<string, Linter.RuleEntry>
}

describe('native Date lint rule', () => {
  it('allows Date construction but rejects native time operations', async () => {
    const linter = new Linter()
    const messages = linter.verify(
      [
        'var created = new Date(epochMilliseconds)',
        'var called = Date()',
        'var timestamp = Date.now()',
        'created.setHours(0)',
        'created.getTime()',
        'created.toISOString()',
      ].join('\n'),
      {
        parserOptions: {
          ecmaVersion: 2022,
          sourceType: 'module',
        },
        rules: {
          'no-restricted-syntax': eslintConfig.rules['no-restricted-syntax'],
        },
      }
    )

    expect(messages).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          message: 'Use the clock wrapper or date-fns instead of native Date methods',
        }),
        expect.objectContaining({ message: 'Use date-fns instead of native Date methods' }),
      ])
    )
    expect(messages).toHaveLength(5)
  })
})
