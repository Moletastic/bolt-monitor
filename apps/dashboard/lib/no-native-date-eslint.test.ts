import { ESLint } from 'eslint'
import { describe, expect, it } from 'vitest'

describe('native Date lint rule', () => {
  it('rejects new Date in dashboard source', async () => {
    const eslint = new ESLint({ cwd: process.cwd() })
    const [result] = await eslint.lintText('const value = new Date()\n', {
      filePath: 'app/page.tsx',
    })

    expect(result.messages).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ message: 'Use date-fns instead of native Date' }),
      ])
    )
  })
})
