import { readFileSync } from 'node:fs'

export type LifecycleClass = 'persistent' | 'ephemeral'

export interface DeploymentTarget {
  stage: string
  profile: string
  lifecycle: LifecycleClass
  owner: string
  service: string
  accountId: string
  region: string
  dashboardOrigin: string
  approved?: boolean
  disposable?: boolean
  expiresAt?: string
  budgetAmountUsd?: number
  alertEmails?: string[]
}

export interface LifecyclePolicy {
  appRemoval: 'remove'
  appProtect: boolean
  retainDurableResources: boolean
  tags: Record<string, string>
}

export function normalizeStageName(stage: string) {
  return stage.toLowerCase().replace(/[^a-z0-9]/g, '')
}

function requireText(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error(`deployment target requires ${field}`)
  }
  return value.trim()
}

export function parseTarget(value: unknown): DeploymentTarget {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    throw new Error('deployment target must be an object')
  }

  const target = value as Record<string, unknown>
  const lifecycle = requireText(target.lifecycle, 'lifecycle')
  if (lifecycle !== 'persistent' && lifecycle !== 'ephemeral') {
    throw new Error(`deployment target has unknown lifecycle: ${lifecycle}`)
  }

  return {
    stage: requireText(target.stage, 'stage'),
    profile: requireText(target.profile, 'profile'),
    lifecycle,
    owner: requireText(target.owner, 'owner'),
    service: requireText(target.service, 'service'),
    accountId: requireText(target.accountId, 'accountId'),
    region: requireText(target.region, 'region'),
    dashboardOrigin: requireText(target.dashboardOrigin, 'dashboardOrigin'),
    ...(typeof target.approved === 'boolean' ? { approved: target.approved } : {}),
    ...(typeof target.disposable === 'boolean' ? { disposable: target.disposable } : {}),
    ...(typeof target.expiresAt === 'string' ? { expiresAt: target.expiresAt } : {}),
    ...(typeof target.budgetAmountUsd === 'number' &&
    Number.isFinite(target.budgetAmountUsd) &&
    target.budgetAmountUsd > 0
      ? { budgetAmountUsd: target.budgetAmountUsd }
      : {}),
    ...(Array.isArray(target.alertEmails) &&
    target.alertEmails.length > 0 &&
    target.alertEmails.every((entry) => typeof entry === 'string' && entry.trim() !== '')
      ? { alertEmails: target.alertEmails.map((entry) => (entry as string).trim()) }
      : {}),
  }
}

export function hasBudgetConfig(
  target: DeploymentTarget
): target is DeploymentTarget & { budgetAmountUsd: number; alertEmails: string[] } {
  return (
    typeof target.budgetAmountUsd === 'number' &&
    target.budgetAmountUsd > 0 &&
    Array.isArray(target.alertEmails) &&
    target.alertEmails.length > 0
  )
}

export function validateDeploymentTarget(target: DeploymentTarget) {
  requireText(target.stage, 'stage')
  requireText(target.owner, 'owner')
  requireText(target.service, 'service')
  requireText(target.profile, 'profile')
  const dashboardOrigin = requireText(target.dashboardOrigin, 'dashboardOrigin')
  let dashboardURL
  try {
    dashboardURL = new URL(dashboardOrigin)
  } catch {
    throw new Error('deployment target dashboardOrigin must be an absolute URL')
  }
  if (
    dashboardURL.protocol !== 'https:' ||
    dashboardURL.pathname !== '/' ||
    dashboardURL.search ||
    dashboardURL.hash
  ) {
    throw new Error('deployment target dashboardOrigin must be an HTTPS origin without a path')
  }
  if (!/^\d{12}$/.test(target.accountId)) {
    throw new Error('deployment target accountId must be a 12-digit AWS account ID')
  }
  if (!/^[a-z]{2}-[a-z]+-\d$/.test(target.region)) {
    throw new Error('deployment target region must be an AWS region')
  }

  const normalizedStage = normalizeStageName(target.stage)
  const PROTECTED_NORMALIZED = new Set(['prod', 'production'])

  if (target.lifecycle === 'persistent') {
    if (normalizedStage.startsWith('smoke')) {
      throw new Error('persistent target cannot use a smoke stage name')
    }
    if (target.approved !== true) {
      throw new Error(`persistent target requires approved=true: ${target.stage}`)
    }
    if (target.disposable !== undefined || target.expiresAt !== undefined) {
      throw new Error('persistent target cannot declare disposable or expiresAt')
    }
    return
  }

  if (target.disposable !== true) {
    throw new Error('ephemeral target requires disposable=true')
  }
  if (target.approved !== undefined) {
    throw new Error('ephemeral target cannot declare approved')
  }
  if (
    target.expiresAt === undefined ||
    Number.isNaN(Date.parse(target.expiresAt)) ||
    Date.parse(target.expiresAt) <= Date.now()
  ) {
    throw new Error('ephemeral target requires a valid expiresAt')
  }
  if (PROTECTED_NORMALIZED.has(normalizedStage)) {
    throw new Error(`ephemeral stage uses protected name: ${target.stage}`)
  }
}

export function loadDeploymentTargetFromPath(configPath: string): DeploymentTarget {
  let raw: unknown
  try {
    raw = JSON.parse(readFileSync(configPath, 'utf8'))
  } catch {
    throw new Error(`unable to read deployment target file: ${configPath}`)
  }
  const target = parseTarget(raw)
  validateDeploymentTarget(target)
  return target
}

export function lifecyclePolicy(target: DeploymentTarget): LifecyclePolicy {
  const persistent = target.lifecycle === 'persistent'
  return {
    appRemoval: 'remove',
    appProtect: persistent,
    retainDurableResources: persistent,
    tags: {
      service: target.service,
      stage: target.stage,
      owner: target.owner,
      lifecycle: target.lifecycle,
      ...(target.expiresAt === undefined ? {} : { expiresAt: target.expiresAt }),
    },
  }
}
