import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, realpathSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import process from 'node:process'
import { loadDeploymentTargetFromPath, validateTargetExpiry } from '../deployment-target.ts'
import { cleanupEphemeral, outputsPath } from './cleanup.mjs'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

function projectRoot() {
  return resolve(__dirname, '..', '..')
}

function namedTargetPath(targetName) {
  const suffix = targetName.endsWith('.target.json') ? targetName : `${targetName}.target.json`
  return resolve(projectRoot(), 'infra', 'targets', suffix)
}

function resolvedTargetPath() {
  const configuredName = process.env.TARGET
  const name = configuredName ?? 'staging'
  const namedPath = namedTargetPath(name)
  const configuredPath = process.env.TARGET_FILE
  const path = resolve(configuredPath ?? namedPath)

  if (!existsSync(path)) {
    throw new Error(
      `target file not found: ${path}\nCopy infra/targets/example.target.json to ${path} and fill it in.`
    )
  }
  if (
    configuredPath !== undefined &&
    configuredName !== undefined &&
    (!existsSync(namedPath) || realpathSync(path) !== realpathSync(namedPath))
  ) {
    throw new Error(
      `conflicting target selection: TARGET=${configuredName} resolves to ${namedPath}, TARGET_FILE=${configuredPath}`
    )
  }
  return realpathSync(path)
}

export function resolveTarget(operation = 'status') {
  const target = loadDeploymentTargetFromPath(resolvedTargetPath())
  validateTargetExpiry(target, operation === 'remove')
  return target
}

function targetEnvironment(target, operation) {
  return {
    ...process.env,
    AWS_PROFILE: target.profile,
    AWS_REGION: target.region,
    TARGET: target.stage,
    SST_TARGET_FILE: resolvedTargetPath(),
    SST_OPERATION: operation,
  }
}

export function targetSummary(target, accountId) {
  return [
    `  target: ${target.stage}`,
    `  stage: ${target.stage}`,
    `  class: ${target.lifecycle}`,
    `  owner: ${target.owner}`,
    `  service: ${target.service}`,
    `  account: ${accountId}`,
    `  region: ${target.region}`,
    `  profile: ${target.profile}`,
    `  dashboard origin: ${target.dashboardOrigin}`,
  ].join('\n')
}

export function runCommand(command, args, environment, { inherit = false, cwd } = {}) {
  return execFileSync(command, args, {
    encoding: 'utf8',
    env: environment,
    stdio: inherit ? 'inherit' : 'pipe',
    ...(cwd === undefined ? {} : { cwd }),
  })
}

export function preflight(target, environment = process.env, { run = runCommand } = {}) {
  const env = environment
  const accountId = run(
    'aws',
    ['sts', 'get-caller-identity', '--query', 'Account', '--output', 'text'],
    env
  ).trim()
  const region =
    env.AWS_REGION ??
    env.AWS_DEFAULT_REGION ??
    run('aws', ['configure', 'get', 'region'], env).trim()
  if (accountId !== target.accountId) {
    throw new Error(`AWS account mismatch: expected ${target.accountId}, got ${accountId}`)
  }
  if (region !== target.region) {
    throw new Error(`AWS region mismatch: expected ${target.region}, got ${region || 'unset'}`)
  }
  return { accountId, region }
}

export function sstArgs(action, target) {
  if (action === 'dev')
    return ['--dir', 'infra', 'exec', 'sst', 'dev', '--mode=mono', '--stage', target.stage]
  if (action === 'deploy')
    return ['--dir', 'infra', 'exec', 'sst', 'deploy', '--stage', target.stage]
  if (action === 'remove')
    return ['--dir', 'infra', 'exec', 'sst', 'remove', '--stage', target.stage]
  throw new Error(`unknown action: ${action}`)
}

const requiredDeployOutputFields = ['apiUrl', 'dashboardUrl', 'appTableName', 'authTableName']

function isRecord(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function hasDeployOutputField(value) {
  return requiredDeployOutputFields.some((field) => Object.hasOwn(value, field))
}

export function resolveDeployOutputs(all, stage) {
  if (!isRecord(all)) {
    throw new Error('deploy outputs are malformed: expected an object')
  }

  let data
  if (Object.hasOwn(all, stage)) {
    if (hasDeployOutputField(all)) {
      throw new Error('deploy outputs are malformed: mixed flat and stage-scoped outputs')
    }
    data = all[stage]
  } else if (hasDeployOutputField(all)) {
    data = all
  } else if (Object.values(all).some(isRecord)) {
    throw new Error(`deploy outputs do not contain selected stage: ${stage}`)
  } else {
    data = all
  }

  if (!isRecord(data)) {
    throw new Error(`deploy outputs are malformed for stage ${stage}: expected an object`)
  }

  const missing = requiredDeployOutputFields.filter(
    (field) => typeof data[field] !== 'string' || data[field].trim() === ''
  )
  if (missing.length > 0) {
    throw new Error(
      `deploy outputs missing required fields for stage ${stage}: ${missing.join(', ')}`
    )
  }
  return data
}

export function sstOutputs(stage) {
  const path = resolve(projectRoot(), 'infra', '.sst', 'outputs.json')
  if (!existsSync(path)) {
    throw new Error(`SST outputs not found at ${path}; deploy first`)
  }
  const all = JSON.parse(readFileSync(path, 'utf8'))
  return resolveDeployOutputs(all, stage)
}

export function verifyHealthResponse(body) {
  let response
  try {
    response = JSON.parse(body)
  } catch {
    throw new Error('public health endpoint returned malformed JSON')
  }
  if (response?.status !== 'success' || response?.data?.status !== 'ok') {
    throw new Error('public health endpoint returned an unsuccessful health envelope')
  }
}

export function verifyPersistentTable(tableName, environment, run = runCommand) {
  const tableJson = JSON.parse(
    run(
      'aws',
      ['dynamodb', 'describe-table', '--table-name', tableName, '--output', 'json'],
      environment
    )
  )
  if (tableJson.Table?.DeletionProtectionEnabled !== true) {
    throw new Error(`persistent table lacks deletion protection: ${tableName}`)
  }
  const backupJson = JSON.parse(
    run(
      'aws',
      ['dynamodb', 'describe-continuous-backups', '--table-name', tableName, '--output', 'json'],
      environment
    )
  )
  const status =
    backupJson?.ContinuousBackupsDescription?.PointInTimeRecoveryDescription
      ?.PointInTimeRecoveryStatus
  if (status !== 'ENABLED') {
    throw new Error(
      `persistent table lacks point-in-time recovery: ${tableName} (status=${status ?? 'unknown'})`
    )
  }
}

export async function status({ run = runCommand } = {}) {
  const target = resolveTarget('status')
  const { accountId } = preflight(target, targetEnvironment(target, 'status'), { run })
  console.log(`SST target validated:\n${targetSummary(target, accountId)}`)
  return { target, accountId }
}

export async function deploy({
  run = runCommand,
  preflight: pre = preflight,
  outputs = sstOutputs,
} = {}) {
  const target = resolveTarget('deploy')
  const env = targetEnvironment(target, 'deploy')
  const { accountId } = pre(target, env, { run })
  console.log(`SST deploy target:\n${targetSummary(target, accountId)}`)
  run('pnpm', sstArgs('deploy', target), env, { inherit: true })
  const data = resolveDeployOutputs(outputs(target.stage), target.stage)
  try {
    const body = run('curl', ['-fsS', `${data.apiUrl.replace(/\/$/, '')}/api/health`], env, {
      inherit: false,
    })
    verifyHealthResponse(body)
    console.log('Public health endpoint validated')
  } catch (error) {
    throw new Error(`public health endpoint contract failed after deploy: ${error.message}`)
  }
  if (target.lifecycle === 'persistent') {
    verifyPersistentTable(data.appTableName, env, run)
    verifyPersistentTable(data.authTableName, env, run)
  }
}

export async function dev({ run = runCommand, preflight: pre = preflight } = {}) {
  const target = resolveTarget('dev')
  const env = targetEnvironment(target, 'dev')
  const { accountId } = pre(target, env, { run })
  console.log(`SST dev target:\n${targetSummary(target, accountId)}`)
  run('pnpm', sstArgs('dev', target), env, { inherit: true })
}

export async function remove(
  options = {},
  { run = runCommand, preflight: pre = preflight, cleanup = cleanupEphemeral } = {}
) {
  const target = resolveTarget('remove')
  if (target.lifecycle === 'persistent' && options.destroy !== true) {
    throw new Error('persistent removal requires DESTROY=yes')
  }
  const env = targetEnvironment(target, 'remove')
  const { accountId } = pre(target, env, { run })
  console.log(`SST remove target:\n${targetSummary(target, accountId)}`)
  if (target.lifecycle === 'ephemeral') {
    const result = cleanup(target, env)
    console.log(
      `SST cleanup verified zero residual resources across: ${result.coveredResourceKinds.join(', ')}`
    )
    console.log(`SST cleanup pre-removal inventory: ${result.stateResources.join(', ') || 'none'}`)
    return result
  }
  run(
    'pnpm',
    sstArgs('remove', target),
    {
      ...env,
      SST_ALLOW_PERSISTENT_REMOVAL: '1',
    },
    { inherit: true }
  )
  return null
}

export async function inviteAdmin(
  email,
  { run = runCommand, preflight: pre = preflight, outputs = outputsPath } = {}
) {
  if (typeof email !== 'string' || email.trim() === '') {
    throw new Error('invite-admin requires EMAIL=<email>')
  }
  const target = resolveTarget('invite-admin')
  const env = targetEnvironment(target, 'invite-admin')
  const { accountId } = pre(target, env, { run })
  const data = outputs(target.stage)
  const userPoolId = data?.operatorUserPoolId
  const authTable = data?.authTableName
  if (typeof userPoolId !== 'string' || typeof authTable !== 'string') {
    throw new Error('invite-admin requires deploy outputs; run make deploy-infra first')
  }
  console.log(`Invite target:\n${targetSummary(target, accountId)}\n  user: ${email}`)
  run(
    'go',
    [
      'run',
      './tools/admin-bootstrap',
      '--email',
      email,
      '--user-pool-id',
      userPoolId,
      '--auth-table',
      authTable,
      '--stage',
      target.stage,
    ],
    env,
    { inherit: true }
  )
}

export async function rotateAuthKey({ run = runCommand, preflight: pre = preflight } = {}) {
  const target = resolveTarget('rotate-auth-key')
  const env = targetEnvironment(target, 'rotate-auth-key')
  const { accountId } = pre(target, env, { run })
  console.log(`Auth key rotation target: ${targetSummary(target, accountId)}`)
  const { randomBytes } = await import('node:crypto')
  const key = randomBytes(32).toString('base64url')
  const parameterName = `/${target.service}/${target.stage}/auth/aes-256-gcm`
  run(
    'aws',
    [
      'ssm',
      'put-parameter',
      '--name',
      parameterName,
      '--type',
      'SecureString',
      '--value',
      key,
      '--overwrite',
      '--output',
      'json',
    ],
    env
  )
  run(
    'aws',
    [
      'ssm',
      'add-tags-to-resource',
      '--resource-type',
      'Parameter',
      '--resource-id',
      parameterName,
      '--tags',
      `Key=service,Value=${target.service}`,
      `Key=stage,Value=${target.stage}`,
      `Key=owner,Value=${target.owner}`,
      `Key=lifecycle,Value=${target.lifecycle}`,
    ],
    env
  )
  console.log(`Wrote SecureString parameter ${parameterName}`)
}

export async function setupReadiness(
  email,
  { rotate = false, run = runCommand, preflight: pre = preflight, outputs = outputsPath } = {}
) {
  const target = resolveTarget('setup-readiness')
  const env = targetEnvironment(target, 'setup-readiness')
  const { accountId } = pre(target, env, { run })
  const data = outputs(target.stage)
  if (!data?.operatorUserPoolId || !data?.authTableName)
    throw new Error('setup-readiness requires deploy outputs; run make deploy-infra first')
  const parameterName = `/${target.service}/${target.stage}/readiness/password`
  const username =
    typeof email === 'string' && email.trim() !== ''
      ? email.trim()
      : `synthetic-readiness+${target.stage}@example.invalid`
  console.log(
    `Readiness setup target:\n${targetSummary(target, accountId)}\n  user: ${username}\n  parameter: ${parameterName}`
  )
  try {
    run(
      'aws',
      [
        'ssm',
        'get-parameter',
        '--name',
        parameterName,
        '--query',
        'Parameter.Name',
        '--output',
        'text',
      ],
      env
    )
    if (!rotate) {
      console.log('Readiness credentials already configured; use ROTATE=yes to rotate')
      return
    }
  } catch {
    // Missing parameter is expected on initial setup.
  }
  const { randomBytes } = await import('node:crypto')
  const password = `${randomBytes(18).toString('base64url')}!Aa1`
  run(
    'go',
    [
      'run',
      './tools/admin-bootstrap',
      '--email',
      username,
      '--user-pool-id',
      data.operatorUserPoolId,
      '--auth-table',
      data.authTableName,
      '--stage',
      target.stage,
    ],
    { ...env, SYNTHETIC_PASSWORD: password },
    { inherit: true }
  )
  run(
    'aws',
    [
      'ssm',
      'put-parameter',
      '--name',
      parameterName,
      '--type',
      'SecureString',
      '--value',
      password,
      '--overwrite',
    ],
    env
  )
  console.log('Readiness credentials configured')
}

export async function readinessAPI(
  email,
  { run = runCommand, preflight: pre = preflight, outputs = outputsPath } = {}
) {
  const target = resolveTarget('readiness-api')
  const env = targetEnvironment(target, 'readiness-api')
  pre(target, env, { run })
  const data = outputs(target.stage)
  if (!data?.apiUrl || !data?.readinessUserPoolClientId)
    throw new Error('readiness-api requires deploy outputs; run make deploy-infra first')
  const parameterName = `/${target.service}/${target.stage}/readiness/password`
  const username =
    typeof email === 'string' && email.trim() !== ''
      ? email.trim()
      : `synthetic-readiness+${target.stage}@example.invalid`
  run(
    'go',
    [
      'run',
      './tools/readiness-probe',
      '--api-url',
      data.apiUrl,
      '--client-id',
      data.readinessUserPoolClientId,
      '--username',
      username,
      '--password-parameter',
      parameterName,
    ],
    env,
    { inherit: true }
  )
}

export function parseOperationInputs(argv) {
  const flags = {}
  for (const arg of argv) {
    const eq = arg.indexOf('=')
    if (eq === -1) continue
    const key = arg.slice(0, eq).replace(/^--/, '')
    const value = arg.slice(eq + 1)
    if (key === 'DESTROY' || key === 'EMAIL' || key === 'ROTATE') flags[key] = value
  }
  return flags
}

function positional(argv) {
  return argv.filter((arg) => !arg.startsWith('--') && !arg.includes('='))
}

export async function dispatch(
  command,
  inputs,
  rest = [],
  operations = {
    status,
    deploy,
    dev,
    remove,
    inviteAdmin,
    rotateAuthKey,
    setupReadiness,
    readinessAPI,
  }
) {
  if (command === 'status') return operations.status()
  if (command === 'deploy') return operations.deploy()
  if (command === 'dev') return operations.dev()
  if (command === 'remove') return operations.remove({ destroy: inputs.DESTROY === 'yes' })
  if (command === 'invite-admin') return operations.inviteAdmin(inputs.EMAIL ?? rest[0])
  if (command === 'rotate-auth-key') return operations.rotateAuthKey()
  if (command === 'setup-readiness')
    return operations.setupReadiness(inputs.EMAIL ?? rest[0], { rotate: inputs.ROTATE === 'yes' })
  if (command === 'readiness-api') return operations.readinessAPI(inputs.EMAIL ?? rest[0])
  throw new Error(`unknown command: ${command}`)
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const args = process.argv.slice(2)
  const [command, ...rest] = positional(args)
  const inputs = parseOperationInputs(args)
  try {
    await dispatch(command, inputs, rest)
  } catch (error) {
    console.error(error.message)
    process.exit(1)
  }
}
