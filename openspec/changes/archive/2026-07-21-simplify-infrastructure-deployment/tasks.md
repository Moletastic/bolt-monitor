## 1. Target Configuration

- [x] 1.1 Replace the ignored multi-target registry with ignored `infra/targets/<name>.target.json` files and a committed `example.target.json` template.
- [x] 1.2 Update target parsing and validation for one target per file, including required AWS profile, account, region, lifecycle class, owner, service, dashboard origin, and class-specific fields.
- [x] 1.3 Preserve persistent approval/protection rules, ephemeral expiry/disposal rules, and protected stage-name checks under target-file selection.
- [x] 1.4 Add target-file unit tests for default staging selection, `TARGET=<name>` selection, invalid files, AWS profile binding, and lifecycle validation.

## 2. Infrastructure Orchestration

- [x] 2.1 Create `infra/scripts/ops.ts` and its small target-loading helper as the sole internal entrypoint for status, deploy, development, removal, and administrator invitation actions.
- [x] 2.2 Bind `AWS_PROFILE` and `AWS_REGION` from the selected target, verify STS account and region before mutation, and print a secret-safe target summary.
- [x] 2.3 Update `sst.config.ts` to consume the selected target model and keep AWS profile handling outside `stacks/bootstrap.ts`.
- [x] 2.4 Make ordinary `make deploy-infra` deployment intent sufficient; retain `DESTROY=yes` and persistent protection requirements for removal.
- [x] 2.5 Add deploy postflight checks for SST outputs, persistent resource protections/tags, and public health.
- [x] 2.6 Preserve verified residual cleanup for explicit ephemeral removal and cover it through orchestrator tests.

## 3. Root Commands And Administrator Bootstrap

- [x] 3.1 Add idempotent root `make setup` for locked infra/dashboard dependency installation and Go workspace synchronization.
- [x] 3.2 Route `make infra-status`, `make deploy-infra`, `make dev-infra`, `make remove-infra`, and `make invite-admin EMAIL=<email>` through the internal orchestrator.
- [x] 3.3 Make administrator invitation resolve user-pool and auth-table identifiers from selected-target SST outputs before invoking the existing idempotent bootstrap tool.
- [x] 3.4 Remove direct SST mutation package scripts, obsolete lifecycle/target wrappers, copied-confirmation behavior, and manual bootstrap resource-ID arguments.
- [x] 3.5 Add Make/orchestrator and output-driven invitation tests, including missing output and target mismatch failures.

## 4. Remove Credentialed Staging Smoke

- [x] 4.1 Remove `make smoke-staging`, credentialed staging smoke scripts, token helper scripts, their tests, and smoke-specific target examples.
- [x] 4.2 Remove staging-smoke references from deterministic release gates, infrastructure test commands, and operational documentation.

## 5. Documentation And Verification

- [x] 5.1 Rewrite README setup and command sections around target files, `make setup`, normal deployment, administrator invitation, activation, and optional TOTP enrollment.
- [x] 5.2 Update AGENTS and lifecycle/auth runbooks to remove stale profile, exported-variable, direct-script, smoke, and manual-output-ID instructions.
- [x] 5.3 Document the canonical HTTPS dashboard-origin requirement for a generated dashboard URL without suggesting unsafe origin inference.
- [x] 5.4 Run `make check-infra`, `make test-infra`, `make check-auth-cutover-prerequisites`, relevant Go tests, and documentation command checks.
