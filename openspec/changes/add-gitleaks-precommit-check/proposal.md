## Why

Local hooks format files and validate commit messages but do not inspect staged content for credentials. A fail-closed Gitleaks check gives contributors immediate feedback before a secret reaches repository history.

## What Changes

- Run Gitleaks against staged content in Lefthook's pre-commit hook.
- Fail the commit with concise installation guidance when Gitleaks is unavailable.
- Keep CI authoritative; this change adds only local commit-time validation.

## Capabilities

### New Capabilities

### Modified Capabilities
- `repository-local-hooks`: Require staged credential-leak detection and a clear missing-tool failure.

## Impact

- Affects `lefthook.yml` and local contributor workflow.
- Requires contributors to install Gitleaks; it adds no application runtime dependency.
