## 1. Hook Tooling

- [x] 1.1 Add Lefthook to the root JavaScript development dependencies using the repository package manager policy.
- [x] 1.2 Add a root package script that installs or initializes Lefthook hooks non-interactively.
- [x] 1.3 Add committed Lefthook configuration for pre-commit and commit-msg hooks.

## 2. Makefile Delegation

- [x] 2.1 Add file-scoped dashboard formatting target that accepts repo-root staged file paths.
- [x] 2.2 Add file-scoped infrastructure formatting target that accepts repo-root staged file paths.
- [x] 2.3 Add or adjust a Makefile commitlint target that can validate the commit message file passed by Git hooks.

## 3. Hook Behavior

- [x] 3.1 Configure pre-commit dashboard formatting to call the dashboard Makefile target for staged dashboard files and restage fixes.
- [x] 3.2 Configure pre-commit infrastructure formatting to call the infrastructure Makefile target for staged infrastructure files and restage fixes.
- [x] 3.3 Configure commit-msg validation to call the Makefile commitlint target.

## 4. Documentation and Verification

- [x] 4.1 Document local hook setup and exceptional bypass behavior in repository developer guidance.
- [x] 4.2 Verify Lefthook installs from the documented setup command.
- [x] 4.3 Verify a staged dashboard formatting issue is fixed and restaged by pre-commit.
- [x] 4.4 Verify a staged infrastructure formatting issue is fixed and restaged by pre-commit.
- [x] 4.5 Verify an invalid commit message is rejected by the commit-msg hook.
- [x] 4.6 Run relevant repository checks for changed tooling files.
