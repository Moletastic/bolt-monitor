## Why

The repository already contains a functional Next.js operator dashboard, but it only has local development guidance and no deployable hosting path. That leaves the product story split: the API can be deployed through SST, while the web console still requires ad hoc local startup and manual environment configuration.

## What Changes

- Add SST-managed hosting for `apps/dashboard` using a standalone `sst.aws.Nextjs` site
- Wire the deployed dashboard runtime to the deployed monitor API by injecting `MONITOR_API_BASE_URL` from the SST stack
- Publish the dashboard URL as a stack output so teams can use the generated CloudFront hostname without custom DNS
- Document the dashboard deployment path and runtime expectations

## Capabilities

### New Capabilities
- `dashboard-site-hosting`: Deploy the Next.js dashboard through SST with generated hosting URL and runtime API wiring

### Modified Capabilities

(None)

## Impact

- `infra/stacks/bootstrap.ts` and related SST outputs
- `apps/dashboard` deployment/runtime configuration
- Root and dashboard documentation for deployment and environment behavior
- AWS resources for Next.js hosting, including CloudFront and server runtime infrastructure managed by SST
