## Context

The repository already exposes verified local commands through the root `Makefile`, and JavaScript formatting is currently owned by the package-level Prettier scripts in `apps/dashboard` and `infra`. There is no local Git hook runner configured, so contributors can commit formatting drift and only discover it when CI runs format checks.

The root `package.json` already includes commitlint dependencies and a `commitlint` script, but there is no committed hook that invokes it.

## Goals / Non-Goals

**Goals:**

- Use Lefthook to run local Git hooks from a committed repository config.
- Keep formatting command ownership in the Makefile and package scripts instead of duplicating Prettier logic in hook config.
- Auto-format staged dashboard and infrastructure files before commits complete.
- Validate commit messages locally through the existing commitlint setup.
- Keep CI checks authoritative even when hooks are bypassed.

**Non-Goals:**

- Replace CI validation.
- Add full test suites to pre-commit hooks.
- Format unrelated files when only a small staged set changed.
- Introduce operator-facing product behavior.

## Decisions

### Use Lefthook as the hook runner

Lefthook provides a single committed configuration for Git hooks and works well across polyglot repositories. It can call `make` directly, so the hook file stays thin.

Alternative considered: Husky with lint-staged. This is familiar for frontend projects, but this repository also has Go modules and Makefile-owned validation, making Lefthook a better fit.

### Keep Makefile as command source of truth

The hook configuration should call Makefile targets rather than embedding full Prettier commands. The Makefile can expose file-scoped targets for dashboard and infrastructure formatting, while package-level `format` scripts remain unchanged for whole-package formatting.

Alternative considered: Put `pnpm --dir ... prettier --write` commands directly in `lefthook.yml`. This is shorter but duplicates logic outside the Makefile and makes future command changes easier to miss.

### Use file-scoped formatting for pre-commit

Pre-commit should only format staged dashboard and infrastructure files matched by Lefthook globs, then restage fixed files. This avoids surprising rewrites elsewhere in the workspace.

Alternative considered: Call existing whole-package `make format-dashboard` and `make format-infra` targets. That avoids new Makefile targets but may format unrelated files and create noisy commits.

### Start local enforcement with pre-commit and commit-msg

The first hook setup should focus on the known failure mode: Prettier drift before CI. Commit message validation is low-cost because commitlint is already present. Heavier pre-push checks can be added later if local friction stays acceptable.

Alternative considered: Add a pre-push hook that runs typechecks and format checks. This catches more issues earlier but can slow pushes and encourage bypassing all hooks.

## Risks / Trade-offs

- Hooks can be bypassed with Git flags -> CI remains the final gate and documentation should describe bypass as exceptional.
- Lefthook must be installed locally -> document installation/setup and add a package script so contributors can install hooks predictably.
- File-scoped Makefile targets need careful path handling -> run Prettier from the repository root via `pnpm --dir <package>` so repo-root staged paths remain valid.
- Prettier may touch generated or dependency files if globs are broad -> restrict Lefthook globs to source/config/doc extensions under `apps/dashboard` and `infra`.

## Migration Plan

1. Add Lefthook dependency/configuration at the repository root.
2. Add Makefile targets for file-scoped dashboard and infrastructure formatting plus commit message validation if needed.
3. Install Lefthook hooks locally and verify a staged formatting change is fixed before commit.
4. Document developer setup.
5. Rollback by removing Lefthook config/dependency and Makefile hook targets; CI remains unchanged.

## Open Questions

- Should a follow-up add pre-push format/typecheck gates after measuring local workflow friction?
