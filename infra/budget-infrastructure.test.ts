import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

import {
  hasBudgetConfig,
  parseTarget,
  validateDeploymentTarget,
  type DeploymentTarget,
} from './deployment-target.ts'

const stackSource = readFileSync(
  fileURLToPath(new URL('./stacks/bootstrap.ts', import.meta.url)),
  'utf8'
)

test('optional AWS Budget wiring is conditional and absent by default', () => {
  assert.match(
    stackSource,
    /if \(hasBudgetConfig\(target\)\)\s*\{[\s\S]*new aws\.budgets\.Budget\(/,
    'budget resource must be declared inside a hasBudgetConfig guard'
  )
  assert.match(stackSource, /new aws\.budgets\.Budget\(\s*`MonthlyBudget`/)
  assert.match(stackSource, /name: `\$\{target\.service\}-\$\{target\.stage\}-monthly`/)
  assert.match(stackSource, /budgetType: 'COST'/)
  assert.match(stackSource, /limitUnit: 'USD'/)
  assert.match(stackSource, /timeUnit: 'MONTHLY'/)
  assert.match(
    stackSource,
    /name: 'TagKeyValue',\s*values: \[`user:service\$\$\{target\.service\}`, `user:stage\$\$\{target\.stage\}`\]/
  )
  assert.match(stackSource, /notificationType: 'FORECASTED'/)
  assert.match(stackSource, /notificationType: 'ACTUAL'/)
  assert.match(stackSource, /threshold: 80/)
  assert.match(stackSource, /threshold: 100/)
  assert.match(stackSource, /subscriberEmailAddresses: target\.alertEmails/)
})

test('parseTarget accepts target without budget fields', () => {
  const raw = {
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
  }
  const target = parseTarget(raw)
  assert.equal(target.budgetAmountUsd, undefined)
  assert.equal(target.alertEmails, undefined)
  assert.equal(hasBudgetConfig(target), false)
  validateDeploymentTarget(target)
})

test('parseTarget accepts target with valid budget fields', () => {
  const raw = {
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
    budgetAmountUsd: 10,
    alertEmails: ['ops@example.com'],
  }
  const target = parseTarget(raw)
  assert.equal(target.budgetAmountUsd, 10)
  assert.deepEqual(target.alertEmails, ['ops@example.com'])
  assert.equal(hasBudgetConfig(target), true)
})

test('parseTarget rejects non-positive budgetAmountUsd', () => {
  const base = {
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
    alertEmails: ['ops@example.com'],
  }
  assert.equal(parseTarget({ ...base, budgetAmountUsd: 0 }).budgetAmountUsd, undefined)
  assert.equal(parseTarget({ ...base, budgetAmountUsd: -1 }).budgetAmountUsd, undefined)
  assert.equal(parseTarget({ ...base, budgetAmountUsd: Number.NaN }).budgetAmountUsd, undefined)
})

test('parseTarget rejects empty alertEmails array', () => {
  const raw = {
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
    budgetAmountUsd: 10,
    alertEmails: [],
  }
  const target = parseTarget(raw)
  assert.equal(target.alertEmails, undefined)
  assert.equal(hasBudgetConfig(target), false)
})

test('parseTarget trims whitespace inside alertEmails', () => {
  const raw = {
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
    budgetAmountUsd: 10,
    alertEmails: ['  ops@example.com  ', 'oncall@example.com'],
  }
  const target = parseTarget(raw)
  assert.deepEqual(target.alertEmails, ['ops@example.com', 'oncall@example.com'])
})

test('hasBudgetConfig is false for partially configured targets', () => {
  const amountOnly: DeploymentTarget = parseTarget({
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
    budgetAmountUsd: 10,
  })
  assert.equal(hasBudgetConfig(amountOnly), false)

  const emailsOnly: DeploymentTarget = parseTarget({
    stage: 'staging',
    profile: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    lifecycle: 'persistent',
    owner: 'team',
    service: 'bolt-monitor',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
    alertEmails: ['ops@example.com'],
  })
  assert.equal(hasBudgetConfig(emailsOnly), false)
})
