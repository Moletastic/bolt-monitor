import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const projectRoot = () => resolve(__dirname, '..', '..')

export const coveredResourceKinds = [
  'Cognito user pools',
  'DynamoDB tables',
  'SSM parameters and SST Secrets',
  'EventBridge schedules',
  'SQS queues',
  'S3 buckets',
  'functions',
  'APIs',
  'dashboard resources',
  'log groups',
  'subscriptions',
  'SST support resources',
]

function resourceUrns(value) {
  if (typeof value !== 'object' || value === null) return []
  if (Array.isArray(value)) return value.flatMap(resourceUrns)
  const object = value
  const own =
    typeof object.urn === 'string' &&
    (object.type?.startsWith('aws:') || object.type?.startsWith('sst:'))
      ? [object.urn]
      : []
  return [...own, ...Object.values(object).flatMap(resourceUrns)]
}

export function stateInventory(target, environment = process.env, execute = execFileSync) {
  try {
    const output = execute(
      'pnpm',
      ['--dir', 'infra', 'exec', 'sst', 'state', 'export', '--stage', target.stage],
      {
        encoding: 'utf8',
        env: environment,
      }
    )
    return [...new Set(resourceUrns(JSON.parse(output)))]
  } catch {
    return []
  }
}

export function stageStateIsDeployed(target, environment = process.env, execute = execFileSync) {
  const output = execute('pnpm', ['--dir', 'infra', 'exec', 'sst', 'state', 'list'], {
    encoding: 'utf8',
    env: environment,
  })
  if (typeof output !== 'string') return false
  const stageLine = output.split('\n').find((line) => line.trim().startsWith(target.stage))
  return stageLine !== undefined && !stageLine.includes('(not deployed)')
}

function defaultAws(environment, args) {
  return execFileSync('aws', args, { encoding: 'utf8', env: environment }).trim()
}

function errorText(error) {
  if (!(error instanceof Error)) return String(error)
  return `${error.message}\n${String(error.stderr ?? '')}`
}

function isLiveCognitoUserPool(arn, environment, query) {
  const userPoolID = arn.match(/userpool\/([^/]+)$/)?.[1]
  if (userPoolID === undefined) return true
  try {
    query(environment, ['cognito-idp', 'describe-user-pool', '--user-pool-id', userPoolID])
    return true
  } catch (error) {
    if (/ResourceNotFoundException/i.test(errorText(error))) return false
    return true
  }
}

export function residualInventory(target, environment = process.env, query = defaultAws) {
  const output = query(environment, [
    'resourcegroupstaggingapi',
    'get-resources',
    '--output',
    'json',
    '--tag-filters',
    `Key=service,Values=${target.service}`,
    `Key=stage,Values=${target.stage}`,
  ])
  const resources = JSON.parse(output).ResourceTagMappingList ?? []
  return resources
    .map((resource) => resource.ResourceARN)
    .filter((arn) => typeof arn === 'string')
    .filter(
      (arn) => !arn.includes(':cognito-idp:') || isLiveCognitoUserPool(arn, environment, query)
    )
}

function removeEphemeralAuthKey(target, environment, execute) {
  const parameterName = `/${target.service}/${target.stage}/auth/aes-256-gcm`
  try {
    execute('aws', ['ssm', 'delete-parameter', '--name', parameterName], { env: environment })
  } catch (error) {
    if (/ParameterNotFound/i.test(errorText(error))) return
    throw error
  }
}

export function cleanupEphemeral(
  target,
  environment = process.env,
  execute = execFileSync,
  query = defaultAws
) {
  if (target.lifecycle !== 'ephemeral' || target.disposable !== true) {
    throw new Error('ephemeral cleanup refuses persistent or non-disposable target')
  }
  const stateResources = stateInventory(target, environment, execute)
  let removeError
  try {
    execute('pnpm', ['--dir', 'infra', 'exec', 'sst', 'remove', '--stage', target.stage], {
      stdio: 'inherit',
      env: environment,
    })
  } catch (error) {
    removeError = error
  }
  try {
    removeEphemeralAuthKey(target, environment, execute)
  } catch (error) {
    removeError ??= error
  }
  const residuals = residualInventory(target, environment, query)
  let stateError
  let stateDeployed = true
  try {
    stateDeployed = stageStateIsDeployed(target, environment, execute)
  } catch (error) {
    stateError = error
  }
  if (
    removeError !== undefined ||
    residuals.length > 0 ||
    stateDeployed ||
    stateError !== undefined
  ) {
    const original = removeError instanceof Error ? removeError.message : 'none'
    const state =
      stateError instanceof Error
        ? stateError.message
        : stateDeployed
          ? 'stage remains deployed'
          : 'none'
    throw new Error(
      `ephemeral cleanup failed; original=${original}; residuals=${residuals.join(',') || 'none'}; state=${state}`
    )
  }
  return { residuals, stateResources, coveredResourceKinds }
}

export function outputsPath(stage) {
  const path = resolve(projectRoot(), 'infra', '.sst', 'outputs.json')
  if (!existsSync(path)) {
    throw new Error(`SST outputs not found at ${path}; deploy first`)
  }
  const all = JSON.parse(readFileSync(path, 'utf8'))
  if (all && typeof all === 'object' && all[stage] !== undefined) {
    return all[stage]
  }
  return all
}
