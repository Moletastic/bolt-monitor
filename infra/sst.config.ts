/// <reference path="./.sst/platform/config.d.ts" />

async function resolveTargetPath(stage?: string): Promise<string> {
  const path = await import('node:path')
  const url = await import('node:url')
  const fs = await import('node:fs')
  const targetName = process.env.TARGET ?? 'staging'
  const suffix = targetName.endsWith('.target.json') ? targetName : `${targetName}.target.json`
  const here = url.fileURLToPath(import.meta.url)
  let dir = path.dirname(here)
  for (let i = 0; i < 6; i += 1) {
    const candidate = path.resolve(dir, 'infra', 'targets', suffix)
    if (fs.existsSync(candidate)) return candidate
    const parent = path.dirname(dir)
    if (parent === dir) break
    dir = parent
  }
  return path.resolve(dir, 'infra', 'targets', suffix)
}

export default $config({
  async app(input?: { stage?: string }) {
    const { loadDeploymentTargetFromPath, lifecyclePolicy } = await import('./deployment-target')
    const target = loadDeploymentTargetFromPath(await resolveTargetPath(input?.stage))
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
    const { loadDeploymentTargetFromPath } = await import('./deployment-target')
    const target = loadDeploymentTargetFromPath(await resolveTargetPath($app.stage))
    const { createBootstrapStack } = await import('./stacks/bootstrap')
    return createBootstrapStack(target)
  },
})
