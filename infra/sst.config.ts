/// <reference path="./.sst/platform/config.d.ts" />

import { lifecyclePolicy, loadDeploymentTarget } from './deployment-target'

export default $config({
  app(input?: { stage?: string }) {
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
          defaultTags: { tags: policy.tags },
          ...(process.env.AWS_PROFILE === undefined ? {} : { profile: process.env.AWS_PROFILE }),
        },
      },
    }
  },
  async run(input?: { stage?: string }) {
    const target = loadDeploymentTarget(input?.stage)
    const { createBootstrapStack } = await import('./stacks/bootstrap')
    return createBootstrapStack(target)
  },
})
