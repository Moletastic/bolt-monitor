import { createHash } from 'node:crypto'
import { readFileSync } from 'node:fs'

export type LifecycleClass = 'persistent' | 'ephemeral'

export interface DeploymentTarget {
  stage: string
  lifecycle: LifecycleClass
  owner: string
  service: string
  accountId: string
  region: string
  credentialSource: string
  dashboardOrigin: string
  approved?: boolean
  disposable?: boolean
  expiresAt?: string
}

interface DeploymentTargetFile {
  targets: DeploymentTarget[]
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

function parseTarget(value: unknown): DeploymentTarget {
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
    lifecycle,
    owner: requireText(target.owner, 'owner'),
    service: requireText(target.service, 'service'),
    accountId: requireText(target.accountId, 'accountId'),
    region: requireText(target.region, 'region'),
    credentialSource: requireText(target.credentialSource, 'credentialSource'),
    dashboardOrigin: requireText(target.dashboardOrigin, 'dashboardOrigin'),
    ...(typeof target.approved === 'boolean' ? { approved: target.approved } : {}),
    ...(typeof target.disposable === 'boolean' ? { disposable: target.disposable } : {}),
    ...(typeof target.expiresAt === 'string' ? { expiresAt: target.expiresAt } : {}),
  }
}

export function validateDeploymentTarget(
  target: DeploymentTarget,
  approvedPersistentNames: readonly string[]
) {
  requireText(target.stage, 'stage')
  requireText(target.owner, 'owner')
  requireText(target.service, 'service')
  requireText(target.credentialSource, 'credentialSource')
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
  const protectedNames = new Set([
    'prod',
    'production',
    ...approvedPersistentNames.map(normalizeStageName),
  ])

  if (target.lifecycle === 'persistent') {
    if (normalizedStage.startsWith('smoke')) {
      throw new Error('persistent target cannot use a smoke stage name')
    }
    if (
      target.approved !== true ||
      !approvedPersistentNames.map(normalizeStageName).includes(normalizedStage)
    ) {
      throw new Error(`persistent stage is not approved: ${target.stage}`)
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
  if (protectedNames.has(normalizedStage)) {
    throw new Error(`ephemeral stage uses protected name: ${target.stage}`)
  }
}

export function loadDeploymentTarget(
  stage: string | undefined,
  environment: NodeJS.ProcessEnv = process.env
): DeploymentTarget {
  if (stage === undefined || stage.trim() === '') {
    throw new Error('SST stage is required; set --stage and SST_TARGET_CONFIG')
  }
  const configPath = environment.SST_TARGET_CONFIG
  if (configPath === undefined || configPath.trim() === '') {
    throw new Error(
      'SST_TARGET_CONFIG is required; copy deployment-target.example.json outside source control'
    )
  }

  let file: DeploymentTargetFile
  try {
    file = JSON.parse(readFileSync(configPath, 'utf8')) as DeploymentTargetFile
  } catch {
    throw new Error(`unable to read deployment target config: ${configPath}`)
  }
  if (!Array.isArray(file.targets))
    throw new Error('deployment target config requires targets array')

  const targets = file.targets.map(parseTarget)
  const approvedPersistentNames = targets
    .filter((target) => target.lifecycle === 'persistent' && target.approved === true)
    .map((target) => target.stage)
  const matching = targets.filter((target) => target.stage === stage)
  if (matching.length !== 1)
    throw new Error(`stage must have exactly one configured target: ${stage}`)

  const target = matching[0]
  validateDeploymentTarget(target, approvedPersistentNames)
  return target
}

export function confirmationFor(target: DeploymentTarget, action: string) {
  return createHash('sha256')
    .update(
      [
        action,
        target.service,
        target.stage,
        target.lifecycle,
        target.owner,
        target.accountId,
        target.region,
        target.credentialSource,
        target.dashboardOrigin,
      ].join(':')
    )
    .digest('hex')
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
