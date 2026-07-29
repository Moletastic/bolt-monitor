## Context

The committed Lefthook configuration currently runs Prettier for staged dashboard and infrastructure files, then validates commit messages. Gitleaks is not a repository dependency and may be absent from a contributor workstation, so a silent skip would falsely imply credential analysis occurred.

## Goals / Non-Goals

**Goals:**

- Scan staged content for credentials before a commit completes.
- Fail closed with brief actionable installation guidance when Gitleaks is unavailable.
- Preserve existing formatter and commit-message hooks.

**Non-Goals:**

- Add Gitleaks as a JavaScript dependency, run historical repository scans, or add CI scanning.
- Add a custom Gitleaks allowlist or suppressions.
- Scan unstaged files or expose matched secret values in hook output.

## Decisions

### Use a fail-closed staged Gitleaks command

The Lefthook pre-commit job will first check `command -v gitleaks`, then run Gitleaks in staged-content mode with redaction enabled. If unavailable, it exits nonzero and prints a concise message stating that credential-leak analysis did not run and Gitleaks must be installed.

Alternative considered: warn and continue when missing. Rejected because a warning makes the hook appear protective while allowing unreviewed credentials into history.

### Keep Gitleaks externally installed

Gitleaks remains an independently installed native tool rather than a package dependency. This avoids platform-specific binary lifecycle and install-script trust expansion in the pnpm workspaces.

Alternative considered: package Gitleaks through Node tooling. Rejected because the repository already treats external developer tooling separately and needs no runtime integration.

## Risks / Trade-offs

- [Contributors without Gitleaks cannot commit normally] -> The hook prints the required install action; bypass remains an explicit Git decision and CI remains authoritative for its configured checks.
- [False positives block a commit] -> Start with default detection and require an evidence-backed follow-up before adding a narrowly scoped allowlist.
- [Hook scan adds latency] -> Scan staged content only, not full history.

## Migration Plan

1. Add the Lefthook job and a focused configuration test.
2. Verify the installed, missing-tool, and detected-secret paths.
3. Contributors install Gitleaks before their next normal commit; no application migration or rollback is required.

## Open Questions

None.
