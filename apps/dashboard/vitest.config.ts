import { defineConfig } from 'vitest/config'
import path from 'node:path'

export default defineConfig({
  oxc: {
    jsx: {
      runtime: 'automatic',
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '.'),
    },
  },
  test: {
    environment: 'node',
    include: ['lib/**/*.test.ts', 'lib/**/*.test.tsx', 'app/**/*.test.ts', 'app/**/*.test.tsx'],
    exclude: ['node_modules', '.next', '**/node_modules/**'],
  },
})
