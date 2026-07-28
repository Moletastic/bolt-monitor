/// <reference path="./.sst/platform/config.d.ts" />

function resolveTargetPath(): string {
  const targetPath = process.env.SST_TARGET_FILE
  if (targetPath === undefined || targetPath.trim() === '') {
    throw new Error(
      'SST_TARGET_FILE is required; invoke infrastructure commands through infra/scripts/ops.mjs'
    )
  }
  return targetPath
}

async function loadTarget(stage?: string) {
  const { loadDeploymentTargetFromPath, validateTargetExpiry } = await import('./deployment-target')
  const target = loadDeploymentTargetFromPath(resolveTargetPath())
  validateTargetExpiry(target, process.env.SST_OPERATION === 'remove')
  if (stage !== undefined && stage !== target.stage) {
    throw new Error(`SST stage ${stage} conflicts with target stage ${target.stage}`)
  }
  return target
}

export default $config({
  async app(input?: { stage?: string }) {
    const { lifecyclePolicy } = await import('./deployment-target')
    const target = await loadTarget(input?.stage)
    const policy = lifecyclePolicy(target)
    return {
      name: 'bolt-monitor',
      home: 'aws',
      removal: policy.appRemoval,
      protect: policy.appProtect && process.env.SST_ALLOW_PERSISTENT_REMOVAL !== '1',
      providers: {
        aws: {
          region: target.region,
          profile: target.profile,
        },
      },
    }
  },
  async run() {
    const target = await loadTarget($app.stage)
    const { createBootstrapStack } = await import('./stacks/bootstrap')
    return createBootstrapStack(target)
  },
})
