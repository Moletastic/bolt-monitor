import { execFileSync } from 'node:child_process';

export const coveredResourceKinds = [
  'Cognito user pools', 'DynamoDB tables', 'SSM parameters and SST Secrets',
  'EventBridge schedules', 'SQS queues', 'S3 buckets', 'functions', 'APIs',
  'dashboard resources', 'log groups', 'subscriptions', 'SST support resources',
];

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

export function cleanupEphemeral(target, environment = process.env, execute = execFileSync, query = aws) {
  if (target.lifecycle !== 'ephemeral' || target.disposable !== true) {
    throw new Error('ephemeral cleanup refuses persistent or non-disposable target');
  }
  let removeError;
  try {
    execute('pnpm', ['--dir', 'infra', 'exec', 'sst', 'remove', '--stage', target.stage], { stdio: 'inherit', env: environment });
  } catch (error) {
    removeError = error;
  }
  const residuals = residualInventory(target, environment, query);
  if (removeError !== undefined || residuals.length > 0) {
    const original = removeError instanceof Error ? removeError.message : 'none';
    throw new Error(`ephemeral cleanup failed; original=${original}; residuals=${residuals.join(',') || 'none'}`);
  }
  return { residuals, coveredResourceKinds };
}
