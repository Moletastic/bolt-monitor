import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import process from 'node:process';
import { confirmationFor, loadDeploymentTarget } from '../infra/deployment-target.ts';
import { cleanupEphemeral } from './sst-cleanup.mjs';

const actions = new Set(['dev', 'preview', 'deploy', 'remove', 'adopt', 'unprotect']);

function run(command, args, environment) {
  return execFileSync(command, args, { encoding: 'utf8', env: environment }).trim();
}

export function preflight(environment = process.env, execute = run) {
  const action = environment.SST_LIFECYCLE_ACTION;
  if (!actions.has(action)) throw new Error('SST_LIFECYCLE_ACTION must be dev, preview, deploy, remove, adopt, or unprotect');
  const target = loadDeploymentTarget(environment.SST_STAGE, environment);
  const accountId = execute('aws', ['sts', 'get-caller-identity', '--query', 'Account', '--output', 'text'], environment);
  const region = environment.AWS_REGION ?? environment.AWS_DEFAULT_REGION ?? execute('aws', ['configure', 'get', 'region'], environment);
  if (accountId !== target.accountId) throw new Error(`AWS account mismatch: expected ${target.accountId}, got ${accountId}`);
  if (region !== target.region) throw new Error(`AWS region mismatch: expected ${target.region}, got ${region || 'unset'}`);

  const confirmation = confirmationFor(target, action);
  if (environment.SST_TARGET_CONFIRMATION !== confirmation) {
    throw new Error(`target confirmation required: set SST_TARGET_CONFIRMATION=${confirmation}`);
  }
  const destructive = action === 'remove' || action === 'unprotect';
  if (target.lifecycle === 'persistent' && destructive && environment.SST_DESTRUCTIVE_CONFIRMATION !== confirmationFor(target, 'destroy')) {
    throw new Error('persistent destruction requires SST_DESTRUCTIVE_CONFIRMATION bound to this target');
  }

  return {
    target,
    summary: `application=${target.service} stage=${target.stage} class=${target.lifecycle} disposable=${target.disposable === true} owner=${target.owner} account=${accountId} region=${region} credential-source=${target.credentialSource}`,
  };
}

export function sstArgs(action, target) {
  const base = ['--dir', 'infra', 'exec', 'sst'];
  if (action === 'dev') return [...base, 'dev', '--mode=mono', '--stage', target.stage];
  if (action === 'preview') {
    throw new Error('SST 4.14.1 has no safe preview command; use the retained-resource runbook and do not substitute sst diff')
  }
  if (action === 'deploy') return [...base, 'deploy', '--stage', target.stage];
  if (action === 'remove') return [...base, 'remove', '--stage', target.stage];
  throw new Error(`${action} requires the documented SST/Pulumi adoption runbook; automatic mutation is disabled`);
}

export function verifyPersistentDeployment(target, tableName, environment = process.env, execute = run) {
  const table = JSON.parse(execute('aws', ['dynamodb', 'describe-table', '--table-name', tableName, '--output', 'json'], environment)).Table;
  if (table.DeletionProtectionEnabled !== true) throw new Error(`persistent AppTable lacks deletion protection: ${tableName}`);
  const tags = JSON.parse(execute('aws', ['dynamodb', 'list-tags-of-resource', '--resource-arn', table.TableArn, '--output', 'json'], environment)).Tags;
  const actual = new Map(tags.map((tag) => [tag.Key, tag.Value]));
  for (const [key, value] of Object.entries({ service: target.service, stage: target.stage, owner: target.owner })) {
    if (actual.get(key) !== value) throw new Error(`persistent AppTable tag mismatch: ${key}`);
  }
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const { target, summary } = preflight();
  console.log(`SST lifecycle target: ${summary}`);
  const action = process.env.SST_LIFECYCLE_ACTION;
  const sstEnvironment = {
    ...process.env,
    SST_TARGET_CONFIG: resolve(process.env.SST_TARGET_CONFIG),
  };
  if (action === 'remove' && target.lifecycle === 'ephemeral') {
    const result = cleanupEphemeral(target, sstEnvironment);
    console.log(`SST cleanup verified zero residual resources across: ${result.coveredResourceKinds.join(', ')}`);
  } else {
    execFileSync('pnpm', sstArgs(action, target), {
      stdio: 'inherit',
      env: target.lifecycle === 'persistent' && action === 'remove'
        ? { ...sstEnvironment, SST_ALLOW_PERSISTENT_REMOVAL: '1' }
        : sstEnvironment,
    });
    if (action === 'deploy' && target.lifecycle === 'persistent') {
      const outputs = JSON.parse(readFileSync('infra/.sst/outputs.json', 'utf8'));
      const tableName = outputs.appTableName;
      if (typeof tableName !== 'string') throw new Error('persistent deploy verification could not resolve AppTable output');
      verifyPersistentDeployment(target, tableName);
    }
  }
}
