/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
  app(input?: { stage?: string }) {
    return {
      name: 'bolt-monitor',
      home: 'aws',
      providers: {
        aws: {
          profile: 'bolt-monitor',
        },
      },
    }
  },
  async run() {
    const { createBootstrapStack } = await import('./stacks/bootstrap')
    return createBootstrapStack()
  },
})
