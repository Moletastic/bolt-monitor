## Why

Formatting failures are reaching CI because local checks are easy to forget. The repository needs a low-friction local Git hook setup that runs the existing formatting entrypoints before commits, while keeping CI as the final source of truth.

## What Changes

- Add Lefthook as the repository Git hook runner.
- Add local pre-commit hooks that format relevant staged dashboard and infrastructure files through Makefile-owned commands.
- Add a local commit message hook that reuses the existing commitlint setup.
- Extend the Makefile with file-scoped formatting targets so hook commands do not duplicate package-level Prettier scripts.
- Document the local hook installation and expected bypass behavior for exceptional cases.

## Capabilities

### New Capabilities

- `repository-local-hooks`: Defines local Git hook behavior for formatting and commit message validation.

### Modified Capabilities

- None.

## Impact

- Root hook configuration: new Lefthook config file.
- Root JavaScript tooling: add Lefthook dependency and install/setup script as needed.
- Root `Makefile`: add hook-friendly formatting and commitlint targets that remain the source of truth for repeated commands.
- Developer workflow: commits may auto-format staged dashboard and infrastructure files before completing.
