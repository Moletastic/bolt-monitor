const path = require('path')

const noNativeDateRules = [
  {
    selector: "NewExpression[callee.name='Date']",
    message: 'Use date-fns instead of native Date',
  },
  {
    selector: "CallExpression[callee.name='Date']",
    message: 'Use date-fns instead of native Date',
  },
  {
    selector: "CallExpression[callee.object.name='Date']",
    message: 'Use the clock wrapper or date-fns instead of native Date methods',
  },
  {
    selector:
      'CallExpression[callee.property.name=/^(get(Date|Day|FullYear|Hours|Milliseconds|Minutes|Month|Seconds|Time|TimezoneOffset|UTCDate|UTCDay|UTCFullYear|UTCHours|UTCMilliseconds|UTCMinutes|UTCMonth|UTCSeconds)|set(Date|FullYear|Hours|Milliseconds|Minutes|Month|Seconds|Time|UTCDate|UTCFullYear|UTCHours|UTCMilliseconds|UTCMinutes|UTCMonth|UTCSeconds)|toISOString|toJSON|toLocaleString|toLocaleDateString|toLocaleTimeString|toUTCString)$/]',
    message: 'Use date-fns instead of native Date methods',
  },
]

/** @type {import('eslint').Linter.Config} */
module.exports = {
  root: true,
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    project: path.join(__dirname, 'tsconfig.json'),
    tsconfigRootDir: __dirname,
  },
  plugins: ['@typescript-eslint'],
  extends: ['next/core-web-vitals', 'plugin:@typescript-eslint/recommended'],
  ignorePatterns: ['.next', 'node_modules', 'next.config.ts', 'next-env.d.ts', 'vitest.config.ts'],
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/no-unsafe-argument': 'error',
    '@typescript-eslint/no-unsafe-assignment': 'error',
    '@typescript-eslint/no-unsafe-call': 'error',
    '@typescript-eslint/no-unsafe-member-access': 'error',
    '@typescript-eslint/no-unsafe-return': 'error',
    'no-restricted-syntax': ['error', ...noNativeDateRules],
  },
  overrides: [
    {
      files: ['lib/**/*.{ts,tsx}'],
      rules: {
        'no-restricted-syntax': [
          'error',
          ...noNativeDateRules,
          {
            selector: 'TryStatement',
            message:
              'try/catch is only allowed in apps/dashboard/lib/io/**. Use the I/O boundary helpers (@/lib/io/server-action) or the Result utility (@/lib/result) instead.',
          },
        ],
      },
    },
    {
      files: ['lib/io/**/*.{ts,tsx}'],
      rules: {
        'no-restricted-syntax': ['error', ...noNativeDateRules],
      },
    },
    {
      // Native Date is allowed only in the clock wrapper and test/setup files.
      files: ['lib/clock.ts', 'lib/**/*.test.ts', 'lib/io/**/*.test.ts'],
      rules: {
        'no-restricted-syntax': 'off',
      },
    },
  ],
}
