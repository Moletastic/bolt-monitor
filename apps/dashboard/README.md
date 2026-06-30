# Dashboard App

Next.js operator dashboard for monitor CRUD, status, and recent run inspection.

## Environment

- `MONITOR_API_BASE_URL`: base URL for monitor API, such as local SST API Gateway URL

For local development, export `MONITOR_API_BASE_URL` before running `pnpm run dev`.
For deployed hosting, SST injects this variable automatically from the stack's deployed API URL.

## Commands

This package root uses `pnpm` with a committed `pnpm-lock.yaml`. Do not use
`npm install` against this directory; it will drift from the lockfile and
bypass the install-script trust allowlist.

```bash
pnpm install --frozen-lockfile
pnpm run dev
pnpm run lint
```

## Deployment

This app is deployed through the SST stack in `infra/` using a standalone Next.js site.

```bash
cd infra
pnpm exec sst deploy --stage staging
```

After deployment, SST outputs:

- `apiUrl`: deployed monitor API base URL
- `dashboardUrl`: generated CloudFront URL for the dashboard

The first deployment path uses the generated URL directly; no custom DNS is required.

For local SST-backed development, use the same explicit stage name:

```bash
cd infra
pnpm exec sst dev --stage staging --mode=mono
```

## Scope Notes

- Root `/` is dashboard landing page; current monitor overview now lives at `/services`.
- Sidebar modules are `Dashboard`, `Services`, `Integrations`, `Audit Trail`, and `Config`.
- Uses server-side fetches and server actions for same-origin-friendly bootstrap integration.
- Probe location picker stays pinned to current built-in `iad` assumption for dashboard v1.
- Runtime operator surfaces now exist in API for manual runs, incidents, monitor audit history, and scheduler admin control, but dashboard has not wired those views yet.
- If runtime probe-location discovery becomes required, handle it as explicit follow-on change rather than expanding this app silently.
