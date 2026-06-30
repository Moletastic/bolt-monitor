## Context

The repository already has a local OpenAPI docs workspace under `openapi/`, with Swagger UI and Redoc backed by one checked-in `openapi.yaml`. The next step is to publish that same documentation as part of the SST-managed infrastructure, while ensuring the deployed docs point interactive requests at the real deployed API URL instead of local or placeholder server entries.

## Goals / Non-Goals

**Goals:**
- Deploy the existing OpenAPI docs workspace as a static site through SST-managed infrastructure.
- Publish a docs URL as part of stack outputs.
- Ensure the deployed OpenAPI contract targets the live API URL for interactive docs usage.
- Preserve the existing local docs workflow without forcing a different source format.

**Non-Goals:**
- No migration of docs into the dashboard app.
- No OpenAPI-driven API Gateway deployment.
- No auth, access control, or custom domain work in this change.
- No rewrite of local docs tooling beyond what is needed to support a deployable site build.

## Decisions

### Deploy the existing `openapi/` workspace as a static site
- Decision: reuse the checked-in `openapi/` workspace as the source for an SST-managed static docs site.
- Rationale: keeps one documentation codepath for local and deployed use; avoids duplicating docs assets under `infra/` or `apps/dashboard`.
- Alternative considered: copy docs into the dashboard app or maintain a second deployment-only site directory.
- Why not: both approaches create duplicate ownership and increase drift risk.

### Generate a deployed OpenAPI variant with the real API URL
- Decision: produce a deployed copy of `openapi.yaml` during the site build or deploy flow, replacing the local/example server target with the current stack `apiUrl`.
- Rationale: keeps the checked-in source contract local-friendly while making the deployed docs actually usable for Swagger "Try it out".
- Alternative considered: hardcode the deployed URL in source or patch it client-side in the browser.
- Why not: hardcoding fights local development and multi-stage deploys; client-side patching hides contract behavior in page scripts instead of the published spec.

### Expose docs hosting through SST stack outputs
- Decision: return the docs site URL from the infrastructure stack alongside the existing API outputs.
- Rationale: makes the site discoverable in the same place as the API URL and keeps deploy verification simple.
- Alternative considered: require developers to inspect provider consoles or infer the URL from internal resource names.
- Why not: that weakens the workflow and makes the feature feel bolted on.

### Keep external renderer assets for now
- Decision: allow the deployed pages to continue loading Swagger UI and Redoc assets from public CDNs.
- Rationale: lowest-friction path to a deployable docs site with minimal new build complexity.
- Alternative considered: bundle or vendor renderer assets into the site.
- Why not: reduces external dependencies, but adds asset management work that is not necessary for the first hosted version.

## Risks / Trade-offs

- [Risk] The deployed spec generation step can drift from the checked-in local spec if it performs too much transformation. -> Mitigation: limit generation to server URL substitution and keep all other content identical.
- [Risk] The docs site becomes stage-specific while the checked-in source remains generic. -> Mitigation: treat the generated deployed spec as a build artifact, not a new source file.
- [Risk] CDN-hosted renderer assets introduce a runtime dependency outside AWS. -> Mitigation: keep the HTML minimal so assets can be vendored later without changing the contract structure.

## Migration Plan

1. Add docs-site deployment support to the SST stack.
2. Add a small build or staging step that emits a deployable docs directory with a stack-aware `openapi.yaml`.
3. Publish the docs site and return its URL from stack outputs.
4. Document how to deploy and verify the hosted docs URL.

Rollback: remove the docs site resource and generated docs-site build step. The local `openapi/` workflow remains intact.

## Open Questions

- Should the deployed site keep both local and deployed server entries in the published spec, or only the deployed one?
- Should the docs site be included in all stages or only in shared environments such as production?
