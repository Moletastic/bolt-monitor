import assert from 'node:assert/strict';

export function resolveSSTTarget(environment = process.env) {
  const profile = environment.AWS_PROFILE || 'bolt-monitor';
  const stage = environment.SST_STAGE || 'staging';
  if (stage.toLowerCase() === 'production' || stage.toLowerCase() === 'prod') throw new Error(`refusing production SST stage: ${stage}`);
  return { profile, stage };
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const target = resolveSSTTarget();
  assert.ok(target.profile, 'SST profile must be non-empty');
  console.log(`SST target validated: profile ${target.profile}, stage ${target.stage}`);
}
