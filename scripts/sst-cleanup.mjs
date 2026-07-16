import { execFileSync } from 'node:child_process';

export const coveredResourceKinds = [
  'Cognito user pools', 'DynamoDB tables', 'SSM parameters and SST Secrets',
  'EventBridge schedules', 'SQS queues', 'S3 buckets', 'functions', 'APIs',
  'dashboard resources', 'log groups', 'subscriptions', 'SST support resources',
];

function resourceUrns(value) {
  if (typeof value !== 'object' || value === null) return [];
  if (Array.isArray(value)) return value.flatMap(resourceUrns);
  const object = value;
  const own = typeof object.urn === 'string' && (object.type?.startsWith('aws:') || object.type?.startsWith('sst:'))
    ? [object.urn]
    : [];
  return [...own, ...Object.values(object).flatMap(resourceUrns)];
}

export function stateInventory(target, environment = process.env, execute = execFileSync) {
  try {
    const output = execute('pnpm', ['--dir', 'infra', 'exec', 'sst', 'state', 'export', '--stage', target.stage], {
      encoding: 'utf8',
      env: environment,
    });
    return [...new Set(resourceUrns(JSON.parse(output)))];
  } catch {
    return [];
  }
}

export function stageStateIsDeployed(target, environment = process.env, execute = execFileSync) {
  const output = execute('pnpm', ['--dir', 'infra', 'exec', 'sst', 'state', 'list'], {
    encoding: 'utf8',
    env: environment,
  });
  if (typeof output !== 'string') return false;
  const stageLine = output.split('\n').find((line) => line.trim().startsWith(target.stage));
  return stageLine !== undefined && !stageLine.includes('(not deployed)');
}

function aws(environment, args) {
  return execFileSync('aws', args, { encoding: 'utf8', env: environment }).trim();
}

export function residualInventory(target, environment = process.env, query = aws) {
  const output = query(environment, [
    'resourcegroupstaggingapi', 'get-resources', '--output', 'json',
    '--tag-filters', `Key=service,Values=${target.service}`, `Key=stage,Values=${target.stage}`,
  ]);
  const resources = JSON.parse(output).ResourceTagMappingList ?? [];
  return resources.map((resource) => resource.ResourceARN).filter((arn) => typeof arn === 'string');
}

function removeEphemeralAuthKey(target, environment, execute) {
  const parameterName = `/${target.service}/${target.stage}/auth/aes-256-gcm`;
  try {
    execute('aws', ['ssm', 'delete-parameter', '--name', parameterName], {
      stdio: 'inherit',
      env: environment,
    });
  } catch (error) {
    if (error instanceof Error && /ParameterNotFound/i.test(error.message)) return;
    throw error;
  }
}

export function cleanupEphemeral(target, environment = process.env, execute = execFileSync, query = aws) {
  if (target.lifecycle !== 'ephemeral' || target.disposable !== true) {
    throw new Error('ephemeral cleanup refuses persistent or non-disposable target');
  }
  const stateResources = stateInventory(target, environment, execute);
  let removeError;
  try {
    execute('pnpm', ['--dir', 'infra', 'exec', 'sst', 'remove', '--stage', target.stage], { stdio: 'inherit', env: environment });
  } catch (error) {
    removeError = error;
  }
  try {
    removeEphemeralAuthKey(target, environment, execute);
  } catch (error) {
    removeError ??= error;
  }
  const residuals = residualInventory(target, environment, query);
  let stateError;
  let stateDeployed = true;
  try {
    stateDeployed = stageStateIsDeployed(target, environment, execute);
  } catch (error) {
    stateError = error;
  }
  if (removeError !== undefined || residuals.length > 0 || stateDeployed || stateError !== undefined) {
    const original = removeError instanceof Error ? removeError.message : 'none';
    const state = stateError instanceof Error ? stateError.message : stateDeployed ? 'stage remains deployed' : 'none';
    throw new Error(`ephemeral cleanup failed; original=${original}; residuals=${residuals.join(',') || 'none'}; state=${state}`);
  }
  return { residuals, stateResources, coveredResourceKinds };
}
