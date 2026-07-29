## 1. Pre-commit Credential Scan

- [x] 1.1 Add a Lefthook pre-commit job that verifies Gitleaks is available and runs a redacted staged-content scan.
- [x] 1.2 Make the missing-tool path fail with concise credential-analysis installation guidance.

## 2. Verification

- [x] 2.1 Add focused coverage for the configured Gitleaks command and missing-tool guidance.
- [x] 2.2 Run the hook configuration test and `openspec validate add-gitleaks-precommit-check --strict`.
