import { readFileSync } from 'node:fs';

const root = new URL('..', import.meta.url);

function source(path) {
  return readFileSync(new URL(path, root), 'utf8');
}

function requireMatch(value, pattern, message) {
  if (!pattern.test(value)) throw new Error(`auth cutover lifecycle prerequisite failed: ${message}`);
}

export function checkAuthCutoverPrerequisites() {
  const target = source('infra/deployment-target.ts');
  const stack = source('infra/stacks/bootstrap.ts');
  const orchestrator = source('infra/scripts/ops.mjs');
  const cleanup = source('infra/scripts/cleanup.mjs');

  requireMatch(target, /validateDeploymentTarget\(target\)/, 'stage classification is not validated');
  requireMatch(
    stack,
    /new sst\.aws\.Dynamo\(\s*'AppTable',[\s\S]*?deletionProtection: policy\.retainDurableResources,[\s\S]*?durableOptions/,
    'persistent AppTable protection is absent'
  );
  requireMatch(stack, /retainedResourceInventory: policy\.retainDurableResources/, 'persistent retained inventory is absent');
  requireMatch(stack, /logicalName: 'AuthEncryptionKey', name: authEncryptionKey\.name/, 'auth key is absent from retained inventory');
  requireMatch(cleanup, /target\.lifecycle !== 'ephemeral' \|\| target\.disposable !== true/, 'ephemeral cleanup is not lifecycle guarded');
  requireMatch(cleanup, /ssm', 'delete-parameter'/, 'ephemeral auth key cleanup is absent');
  requireMatch(orchestrator, /DESTROY=yes/, 'persistent removal destructive intent gate is absent');
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  checkAuthCutoverPrerequisites();
  console.log('auth cutover lifecycle prerequisites verified');
}