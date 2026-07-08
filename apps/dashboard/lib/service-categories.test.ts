import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

import { SERVICE_CATEGORIES, SERVICE_CATEGORY_LABELS, isServiceCategory } from './types'

const REPO_ROOT = resolve(__dirname, '..', '..', '..')

function parseGoServiceCategories(source: string): string[] {
  const categories: string[] = []
  const re = /ServiceCategory\w*\s+ServiceCategory\s*=\s*"([^"]+)"/g
  let match: RegExpExecArray | null
  while ((match = re.exec(source)) !== null) {
    categories.push(match[1] as string)
  }
  return categories
}

describe('service category catalog', () => {
  const goSource = readFileSync(
    resolve(REPO_ROOT, 'shared', 'monitorconfig', 'model.go'),
    'utf8'
  )
  const goCategories = parseGoServiceCategories(goSource)

  it('matches the Go service category registry', () => {
    expect([...SERVICE_CATEGORIES].sort()).toEqual([...goCategories].sort())
  })

  it('has labels for every dashboard category', () => {
    for (const category of SERVICE_CATEGORIES) {
      expect(SERVICE_CATEGORY_LABELS[category].length).toBeGreaterThan(0)
    }
  })

  it('narrows known service categories', () => {
    expect(isServiceCategory('payments')).toBe(true)
    expect(isServiceCategory('not-a-category')).toBe(false)
  })
})
