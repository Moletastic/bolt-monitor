/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
  async app(input?: { stage?: string }) {
    const { lifecyclePolicy, loadDeploymentTarget } = await import('./deployment-target')
    const target = loadDeploymentTarget(input?.stage)
    const policy = lifecyclePolicy(target)
    return {
      name: 'bolt-monitor',
      home: 'aws',
      removal: policy.appRemoval,
      protect: policy.appProtect && process.env.SST_ALLOW_PERSISTENT_REMOVAL !== '1',
      providers: {
        aws: {
          region: target.region,
          ...(process.env.AWS_PROFILE === undefined ? {} : { profile: process.env.AWS_PROFILE }),
        },
      },
    }
  },
  async run() {
    const { loadDeploymentTarget } = await import('./deployment-target')
    const target = loadDeploymentTarget($app.stage)
    const { createBootstrapStack } = await import('./stacks/bootstrap')
    return createBootstrapStack(target)
  },
})
