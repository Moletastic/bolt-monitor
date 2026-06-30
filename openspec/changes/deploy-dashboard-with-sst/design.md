## Context

`apps/dashboard` is already a server-rendered Next 15 App Router application, but the repository only documents how to run it locally with a manually exported `MONITOR_API_BASE_URL`. The deployable SST app in `infra/` currently manages the API and backing infrastructure, so the missing piece is a production hosting path for the dashboard that stays inside the same SST-controlled deployment surface.

The dashboard fetch layer reads `process.env.MONITOR_API_BASE_URL` on the server, which makes it a strong fit for SST-managed runtime environment injection. The user does not want custom DNS yet; the first deploy should use the generated CloudFront hostname that SST provides automatically.

## Goals / Non-Goals

**Goals:**
- Deploy `apps/dashboard` through SST using the native `sst.aws.Nextjs` component
- Inject the deployed API URL into the dashboard runtime without manual post-deploy configuration
- Expose the generated dashboard URL as a stack output for immediate use
- Keep the first deployment simple: standalone distribution, no custom domain, no shared router
- Update docs so the dashboard deployment path is discoverable and repeatable

**Non-Goals:**
- Adding authentication, RBAC, or network restrictions to the dashboard
- Moving the API under the same hostname or path-prefix routing scheme
- Introducing custom DNS, ACM certificate management, or Route53 records
- Reworking dashboard application behavior beyond deployment-related requirements

## Decisions

### 1. Use `sst.aws.Nextjs` as a standalone site

**Decision:** Model the dashboard as a standalone `sst.aws.Nextjs` resource in the existing SST stack.

**Rationale:** The app is already a standard Next.js SSR application with server actions and runtime env needs. `sst.aws.Nextjs` is the SST-native way to package and host this shape of app, and it automatically provisions the CloudFront-backed site URL needed for the first deployment.

**Alternatives considered:**
- `sst.aws.Router` plus path/subdomain routing: rejected for now because there is no DNS requirement yet and it adds routing/basePath complexity early.
- Separate hosting platform like Vercel: rejected because this change is specifically about keeping deployment inside SST and AWS.
- Static export: rejected because the dashboard uses server-side fetches and runtime environment variables.

### 2. Inject `MONITOR_API_BASE_URL` from `api.url`

**Decision:** Wire the dashboard runtime to the existing API Gateway URL by setting `MONITOR_API_BASE_URL` from the SST `api.url` output.

**Rationale:** This removes manual environment coordination and guarantees the deployed dashboard always points at the deployed API in the same stack.

**Alternatives considered:**
- Manual environment setting outside SST: rejected because it is error-prone and breaks the one-command deployment story.
- Hard-coded API URL in the dashboard app: rejected because it would drift across stages and environments.

### 3. Publish the dashboard CloudFront URL as a stack output

**Decision:** Add a `dashboardUrl` output alongside the existing stack outputs.

**Rationale:** With no custom domain yet, the generated SST/CloudFront URL is the operator entrypoint. Surfacing it as a standard output keeps deployment ergonomics consistent with the existing `apiUrl` output.

### 4. Keep documentation explicit about public exposure

**Decision:** Document that the dashboard can now be deployed and that the first version uses a generic generated URL.

**Rationale:** The dashboard currently has no auth layer. Deployment docs should make the hosting path obvious without implying that public exposure concerns are already solved.

## Risks / Trade-offs

[Risk] Dashboard becomes internet-reachable through a generated CloudFront URL without application auth.
→ Mitigation: Call this out explicitly in documentation and treat auth or access control as a follow-on change.

[Risk] Adding a Next.js site increases deploy time and infrastructure cost compared with API-only SST deployments.
→ Mitigation: Keep the first version minimal and standalone; avoid extra routing or DNS complexity.

[Risk] Runtime/API wiring works in one stage but drifts in future if someone manually overrides env behavior.
→ Mitigation: Make `MONITOR_API_BASE_URL` stack-managed in SST and document it as the canonical deployment path.

[Risk] Future desire for shared hostname/path routing could force refactoring.
→ Mitigation: Deliberately choose standalone deployment first; revisit `sst.aws.Router` only when DNS/path requirements appear.

## Migration Plan

1. Add the `sst.aws.Nextjs` dashboard site to the bootstrap stack.
2. Inject `MONITOR_API_BASE_URL` from the deployed API URL.
3. Add `dashboardUrl` to stack outputs.
4. Deploy via existing SST deploy workflow.
5. Verify the generated dashboard URL renders and can reach the monitor API.

Rollback:
- Remove the `sst.aws.Nextjs` resource from the stack and redeploy.
- Existing API and Dynamo resources remain independent of the dashboard site.

## Open Questions

- None for the first standalone deployment. Future questions about auth, custom domains, or shared hostname routing are explicitly deferred.
