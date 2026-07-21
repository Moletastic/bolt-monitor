## Why

The lifecycle hardening made ordinary deployment require exported variables, a target registry, a copied confirmation hash, and direct invocation of repository scripts. This contradicts the documented `make deploy-infra` workflow, leaves direct SST bypasses, and imposes release-smoke operations that are not justified for this small project.

## What Changes

- Replace the multi-target deployment registry with one ignored `infra/targets/<name>.target.json` file per deployment target.
- Centralize AWS profile, account, and region in each target file; bind the selected profile to every infrastructure action and validate its effective STS identity before mutation.
- Make `make setup` install both JavaScript workspaces and synchronize the Go workspace.
- Make `make deploy-infra` the ordinary deployment entrypoint without exported variables or direct Node/SST commands. Keep explicit destructive intent only for removal.
- Consolidate infrastructure orchestration behind a small internal `infra/scripts/` surface and remove direct SST mutation entrypoints and obsolete target-validation wrappers.
- Add deploy postflight validation for target identity, SST outputs, persistent-resource protections, and public health.
- Make administrator invitation read the selected target and deployed SST outputs so operators supply only an email address.
- Remove the opt-in credentialed staging release-smoke helper, its scripts, tests, commands, and runbook requirements.
- Resolve first-deployment dashboard-origin configuration without requiring operators to know a generated CloudFront URL before deployment.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `stage-resource-lifecycle`: Simplify target configuration and routine mutation confirmation while retaining explicit lifecycle classification, AWS identity validation, and destructive removal protection.
- `sst-project-bootstrap`: Define the one-command dependency setup and ordinary infrastructure deployment workflow.
- `operator-identity-lifecycle`: Let the credentialed administrator bootstrap command resolve deployed identity resources from target-selected SST outputs.
- `staging-release-smoke`: Remove the local credentialed staging release-smoke capability and its operational requirements.

## Impact

Affected areas include `Makefile`, `infra/`, root deployment scripts, target configuration examples and ignore rules, infrastructure tests, the bootstrap-admin command path, README and lifecycle/auth runbooks. No monitoring-domain API behavior or AWS resource topology is added.
