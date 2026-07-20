import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'

import { Feedback, Unavailable } from '@/components/ui/feedback'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'

const tailwindConfig = readFileSync(resolve(process.cwd(), 'tailwind.config.ts'), 'utf8')

describe('dashboard design-system contract', () => {
  it('maps documented compact radii, typography, spacing, and elevation tokens', () => {
    expect(tailwindConfig).toContain("sm: '0.125rem'")
    expect(tailwindConfig).toContain("DEFAULT: '0.25rem'")
    expect(tailwindConfig).toContain("md: '0.375rem'")
    expect(tailwindConfig).toContain("lg: '0.5rem'")
    expect(tailwindConfig).toContain("xl: '0.75rem'")
    expect(tailwindConfig).toContain("'display-lg': ['2rem'")
    expect(tailwindConfig).toContain("'dashboard-gutter': '1rem'")
    expect(tailwindConfig).toContain("panel: '0 12px 32px rgba(1, 15, 31, 0.28)'")
  })

  it('renders feedback with live semantics and a non-color icon cue', () => {
    const markup = renderToStaticMarkup(createElement(Feedback, { tone: 'success' }, 'Saved.'))

    expect(markup).toContain('role="status"')
    expect(markup).toContain('aria-hidden="true"')
    expect(markup).toContain('Saved.')
  })

  it('announces unavailable content without replacing surrounding content', () => {
    const markup = renderToStaticMarkup(
      createElement(Unavailable, { message: 'Try again later.', title: 'Data unavailable' })
    )

    expect(markup).toContain('aria-live="polite"')
    expect(markup).toContain('Data unavailable')
    expect(markup).toContain('Try again later.')
  })

  it('exposes invalid input and select semantics with a non-color border cue', () => {
    const inputMarkup = renderToStaticMarkup(createElement(Input, { 'aria-invalid': 'true' }))
    const selectMarkup = renderToStaticMarkup(
      createElement(
        Select,
        { 'aria-invalid': 'true' },
        createElement('option', undefined, 'Choose one')
      )
    )

    expect(inputMarkup).toContain('aria-invalid="true"')
    expect(inputMarkup).toContain('border-dashed')
    expect(selectMarkup).toContain('aria-invalid="true"')
    expect(selectMarkup).toContain('border-dashed')
  })
})
