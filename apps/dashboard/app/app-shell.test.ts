import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const shellSource = readFileSync(join(process.cwd(), 'components/app-shell.tsx'), 'utf8')

describe('AppShell source repository link', () => {
  it('renders a separated GitHub source link with external-tab semantics', () => {
    expect(shellSource).toContain('href="https://github.com/Moletastic/bolt-monitor"')
    expect(shellSource).toContain('>View source on GitHub</span>')
    expect(shellSource).toContain('aria-label="View source on GitHub (opens in a new tab)"')
    expect(shellSource).toContain('rel="noreferrer"')
    expect(shellSource).toContain('target="_blank"')
  })

  it('keeps the source link outside active module navigation', () => {
    expect(shellSource).toContain('<nav className="grid gap-2">')
    expect(shellSource).toContain(
      '</nav>\n            <div className="mt-auto border-t border-border pt-4">'
    )
    expect(shellSource).not.toContain("{ href: 'https://github.com/Moletastic/bolt-monitor'")
  })
})
