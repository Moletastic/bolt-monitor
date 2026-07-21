import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import process from 'node:process'
import { loadDeploymentTargetFromPath } from '../deployment-target.ts'
import { cleanupEphemeral, outputsPath } from './cleanup.mjs'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

function projectRoot() {
  return resolve(__dirname, '..', '..')
}

function targetPath(targetName) {
  if (process.env.TARGET_FILE !== undefined) {
    return process.env.TARGET_FILE
  }
  const suffix = targetName.endsWith('.target.json') ? targetName : `${targetName}.target.json`
  return resolve(projectRoot(), 'infra', 'targets', suffix)
}

function resolveTarget() {
  const name = process.env.TARGET ?? 'staging'
  const path = targetPath(name)
  if (!existsSync(path)) {
    throw new Error(
      `target file not found: ${path}\nCopy infra/targets/example.target.json to ${path} and fill it in.`
    )
  }
  return loadDeploymentTargetFromPath(path)
}

function targetEnvironment(target) {
  return {
    ...process.env,
    AWS_PROFILE: target.profile,
    AWS_REGION: target.region,
    TARGET: target.stage,
  }
}

function targetSummary(target, accountId) {
  return [
    `target=${target.stage}`,
    `stage=${target.stage}`,
    `class=${target.lifecycle}`,
    `owner=${target.owner}`,
    `service=${target.service}`,
    `account=${accountId}`,
    `region=${target.region}`,
    `profile=${target.profile}`,
    `dashboard-origin=${target.dashboardOrigin}`,
  ].join(' ')
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

export function sstOutputs(stage) {
  const path = resolve(projectRoot(), 'infra', '.sst', 'outputs.json')
  if (!existsSync(path)) return null
  const all = JSON.parse(readFileSync(path, 'utf8'))
  if (all && typeof all === 'object' && all[stage] !== undefined) {
    return all[stage]
  }
  return all
}

export async function status({ run = runCommand } = {}) {
  const target = resolveTarget()
  const { accountId } = preflight(target, targetEnvironment(target), { run })
  console.log(`SST target validated: ${targetSummary(target, accountId)}`)
  return { target, accountId }
}

export async function deploy({
  run = runCommand,
  preflight: pre = preflight,
  outputs = sstOutputs,
} = {}) {
  const target = resolveTarget()
  const env = targetEnvironment(target)
  const { accountId } = pre(target, env, { run })
  console.log(`SST deploy target: ${targetSummary(target, accountId)}`)
  run('pnpm', sstArgs('deploy', target), env, { inherit: true })
  const data = outputs(target.stage)
  if (typeof data?.apiUrl === 'string') {
    try {
      run('curl', ['-fsS', `${data.apiUrl.replace(/\/$/, '')}/api/health`], env, { inherit: false })
      console.log('Public health endpoint reachable')
    } catch (error) {
      throw new Error(`public health endpoint failed after deploy: ${error.message}`)
    }
  }
  if (target.lifecycle === 'persistent') {
    const tableName = data?.appTableName
    if (typeof tableName !== 'string') {
      throw new Error('persistent deploy verification could not resolve AppTable output')
    }
    const tableJson = JSON.parse(
      run('aws', ['dynamodb', 'describe-table', '--table-name', tableName, '--output', 'json'], env)
    )
    if (tableJson.Table?.DeletionProtectionEnabled !== true) {
      throw new Error(`persistent AppTable lacks deletion protection: ${tableName}`)
    }
    const backupJson = JSON.parse(
      run('aws', ['dynamodb', 'describe-continuous-backups', '--table-name', tableName, '--output', 'json'], env)
    )
    const status = backupJson?.ContinuousBackupsDescription?.PointInTimeRecoveryDescription?.PointInTimeRecoveryStatus
    if (status !== 'ENABLED') {
      throw new Error(`persistent AppTable lacks point-in-time recovery: ${tableName} (status=${status ?? 'unknown'})`)
    }
  }
}

export async function dev({ run = runCommand, preflight: pre = preflight } = {}) {
  const target = resolveTarget()
  const { accountId } = pre(target, targetEnvironment(target), { run })
  console.log(`SST dev target: ${targetSummary(target, accountId)}`)
  run('pnpm', sstArgs('dev', target), targetEnvironment(target), { inherit: true })
}

export async function remove(options = {}, { run = runCommand, preflight: pre = preflight } = {}) {
  const target = resolveTarget()
  if (target.lifecycle === 'persistent' && options.destroy !== true) {
    throw new Error('persistent removal requires DESTROY=yes')
  }
  const { accountId } = pre(target, targetEnvironment(target), { run })
  console.log(`SST remove target: ${targetSummary(target, accountId)}`)
  if (target.lifecycle === 'ephemeral') {
    const result = cleanupEphemeral(target, targetEnvironment(target))
    console.log(
      `SST cleanup verified zero residual resources across: ${result.coveredResourceKinds.join(', ')}`
    )
    return result
  }
  run(
    'pnpm',
    sstArgs('remove', target),
    {
      ...targetEnvironment(target),
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
  const target = resolveTarget()
  const { accountId } = pre(target, targetEnvironment(target), { run })
  const data = outputs(target.stage)
  const userPoolId = data?.operatorUserPoolId
  const authTable = data?.authTableName
  if (typeof userPoolId !== 'string' || typeof authTable !== 'string') {
    throw new Error('invite-admin requires deploy outputs; run make deploy-infra first')
  }
  console.log(`Invite target: ${targetSummary(target, accountId)} user=${email}`)
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
    targetEnvironment(target),
    { inherit: true }
  )
}

export async function rotateAuthKey({ run = runCommand, preflight: pre = preflight } = {}) {
  const target = resolveTarget()
  const env = targetEnvironment(target)
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

function parseFlags(argv) {
  const flags = {}
  for (const arg of argv) {
    const eq = arg.indexOf('=')
    if (eq === -1) continue
    const key = arg.slice(0, eq)
    const value = arg.slice(eq + 1)
    if (key.startsWith('--')) flags[key.slice(2)] = value
  }
  return flags
}

function positional(argv) {
  return argv.filter((arg) => !arg.startsWith('--') && !arg.includes('='))
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const args = process.argv.slice(2)
  const [command, ...rest] = positional(args)
  const flags = parseFlags(args)
  try {
    if (command === 'status') await status()
    else if (command === 'deploy') await deploy()
    else if (command === 'dev') await dev()
    else if (command === 'remove') await remove({ destroy: flags.DESTROY === 'yes' })
    else if (command === 'invite-admin') await inviteAdmin(flags.EMAIL ?? rest[0])
    else if (command === 'rotate-auth-key') await rotateAuthKey()
    else throw new Error(`unknown command: ${command}`)
  } catch (error) {
    console.error(error.message)
    process.exit(1)
  }
}
