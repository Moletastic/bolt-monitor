## Context

The dashboard package declares `devicons-react`, but dashboard service icons now use the repository-owned `TechIcon` component. No dashboard source, test, or build script imports the package. The dashboard uses pnpm with a committed lockfile and hardened install policy.

## Goals / Non-Goals

**Goals:**

- Remove the proven-unused `devicons-react` dependency.
- Preserve reproducible dashboard installs and behavior.

**Non-Goals:**

- Review or modify Go modules, infrastructure dependencies, OpenAPI dependencies, or install-script allowlists.
- Remove or replace used dependencies including Recharts, validator, or ulid.
- Add dependency inventory, vulnerability, bundle-analysis, or unused-dependency CI tooling.

## Decisions

### Remove the dependency with its lockfile entry

Verify the package has no source, test, generated-code, or build-script use before removal. Remove it through pnpm so the manifest and lockfile stay aligned. The existing install-script allowlist does not contain `devicons-react` and remains unchanged.

## Risks / Trade-offs

- [Dynamic or generated use is missed] → Verify source, tests, generated code, and build scripts before removal; run dashboard checks afterward.
- [Lockfile or install policy drifts] → Update only the dashboard lockfile through pnpm and validate a frozen install.

## Migration Plan

1. Verify `devicons-react` has no dashboard use.
2. Remove it from the dashboard manifest and lockfile.
3. Run the frozen install and dashboard validation commands.

Rollback restores the dashboard manifest and lockfile. No runtime data or public protocol migration occurs.
