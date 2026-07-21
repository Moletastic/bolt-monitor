## Context

Infrastructure lifecycle hardening introduced a valid safety model but exposed it as daily operator work. A normal deployment requires an ignored multi-target registry, four exported variables, a copied target-bound hash, and a direct Node command. Profile selection is only a display label, direct SST package scripts bypass preflight, and administrator bootstrap requires operators to copy deployment output identifiers manually.

The repository is a small self-hosted project. Credentialed staging release-smoke automation, including locally supplied tokens and MFA challenges, has more operational cost than value. Lifecycle protections for persistent resources and verified removal for deliberate ephemeral stages remain required.

## Goals / Non-Goals

**Goals:**

- Make `make setup` prepare repository dependencies from the root.
- Make `make deploy-infra` the normal staging deployment command without exported deployment variables or direct script invocation.
- Store each deployment target in one ignored, self-contained `infra/targets/<name>.target.json` file.
- Bind AWS profile, account, and region from the selected target before preflight and SST invocation.
- Preserve fail-closed account/region checks, lifecycle-derived resource protection, and explicit destructive removal intent.
- Make administrator invitation require only `EMAIL`, using target-selected SST output internally.
- Remove local credentialed staging smoke automation and related documentation.

**Non-Goals:**

- Add CI deployment, test credentials, custom domains, a release pipeline, or a new AWS service.
- Change monitoring-domain resources, APIs, authentication policy, or TOTP behavior.
- Automatically create, mutate, or delete target files beyond the ordinary deployment configuration workflow.
- Replace standard AWS profile or credential-provider configuration.

## Decisions

### Use one file per target

Each ignored file at `infra/targets/<name>.target.json` declares one stage, AWS profile, expected account, region, lifecycle class, owner, and dashboard origin. A committed `infra/targets/example.target.json` documents the schema. The default normal target is the explicitly named `staging.target.json`; `TARGET=<name>` selects another file.

This replaces the current registry because a stage is an operational unit, not an entry that operators need to search within a local file. The filename makes target selection visible and prevents profile/account/region drift between entries.

Alternative: retain one target registry with an active-target field. Rejected because it adds selection state and makes the normal command harder to explain. Alternative: export target variables. Rejected because configuration remains distributed and shell state is error-prone.

### Use one internal infrastructure orchestrator

`infra/scripts/ops.ts` is the only imperative entrypoint for deploy, development, removal, status, and output-driven administrator invitation. A small target-loading module under the same directory parses and validates target files. Make targets call `ops.ts`; operators do not.

The orchestrator sets `AWS_PROFILE` and `AWS_REGION` from the selected target, resolves the effective STS account, verifies account and region, invokes SST with the validated stage, and performs action-specific postflight checks. `sst.config.ts` consumes the same target model. `bootstrap.ts` receives only validated deployment values and contains no profile logic.

Alternative: put JSON parsing and AWS checks directly in Make recipes. Rejected because shell parsing and error handling obscure the workflow. Alternative: retain separate root scripts per lifecycle step. Rejected because the public behavior remains scattered and bypassable.

### Treat ordinary deployment as explicit intent

`make deploy-infra` is an explicit ordinary deploy request. It requires no hash confirmation and no `CONFIRM` variable. Before mutation it prints the selected non-secret target summary and verifies effective AWS account and region.

`make remove-infra TARGET=<name> DESTROY=yes` retains separate destructive intent and persistent-stage protection requirements. Ephemeral removal continues to perform verified residual cleanup.

Alternative: retain a copied target-bound confirmation hash for deployment. Rejected because it duplicates configuration data without adding a meaningful safety boundary after the command, target file, and STS identity are explicit. Alternative: remove all confirmations. Rejected because removal can destroy resources.

### Make deploy postflight small and automatic

Deploy postflight verifies SST outputs exist, persistent resource protection and tags remain valid, and public `GET /api/health` succeeds. It does not obtain an access token, invoke TOTP, create an administrator, or test protected application behavior.

Alternative: retain `make smoke-staging`. Rejected because authenticated smoke needs local credentials and sometimes MFA, adds a release process beyond project scope, and does not belong in ordinary deployment.

### Resolve administrator resources from SST output

`make invite-admin EMAIL=<email>` uses the selected target and `infra/.sst/outputs.json` to locate `operatorUserPoolId` and `authTableName`, then invokes the existing idempotent bootstrap tool. Operators do not copy identifiers from output into command arguments.

Alternative: retain `make bootstrap-admin` with output IDs. Rejected because the IDs are already authoritative deployment output and manual copying is avoidable operational friction.

### Keep dashboard origin explicit

The dashboard CSRF boundary requires a canonical HTTPS origin. A target file continues to declare `dashboardOrigin`; generated CloudFront URLs must be recorded after initial infrastructure provisioning before dashboard auth mutations are relied upon. This change corrects the invalid localhost documentation but does not add custom-domain or automatic-origin bootstrap behavior.

Alternative: infer origin from arbitrary request headers. Rejected because it weakens the explicit CSRF origin boundary. Alternative: add custom domains or two-pass automatic origin mutation. Rejected as unnecessary infrastructure and orchestration scope for this change.

## Risks / Trade-offs

- [Target file contains incorrect AWS identity] -> STS account and region checks fail before SST mutation.
- [A developer deploys an alternate target accidentally] -> default is explicit `staging.target.json`; alternate target requires visible `TARGET=<name>`.
- [Direct SST invocation bypasses lifecycle checks] -> remove mutation package scripts and document Make as the supported entrypoint.
- [Generated dashboard URL is unknown on first deployment] -> document the canonical-origin update requirement; auth mutation fails closed until it is configured.
- [Removing smoke loses authenticated acceptance coverage] -> retain deterministic tests and manual dashboard verification; do not represent them as a release gate.
- [Ephemeral stages remain after experimentation] -> retain explicit remove command and residual-resource verification.

## Migration Plan

1. Add target-file example and ignore pattern, then migrate the existing staging configuration to `infra/targets/staging.target.json`.
2. Introduce the internal orchestrator and Make entrypoints while preserving account/region/lifecycle checks and removal protection.
3. Route target loading through SST configuration; remove direct SST mutation scripts and obsolete root wrappers.
4. Add output-driven administrator invitation and deploy postflight checks.
5. Remove staging-smoke scripts, commands, tests, specs, and runbook material.
6. Update README, AGENTS, and lifecycle/auth documentation to the reduced setup and operation path.

Rollback restores the previous command wrappers only before their removal. Persistent resource protections, tags, and retention behavior remain in force throughout rollback.

## Open Questions

None.
