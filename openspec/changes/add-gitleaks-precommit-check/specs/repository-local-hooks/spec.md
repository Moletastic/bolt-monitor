## ADDED Requirements

### Requirement: Pre-commit hook detects staged credential leaks
The local pre-commit hook SHALL run Gitleaks against staged content before the commit completes. It SHALL fail the commit when Gitleaks detects a credential leak and SHALL redact detected secret values in output.

#### Scenario: Staged content has no detected credential
- **WHEN** a contributor commits staged content and Gitleaks is installed
- **THEN** the pre-commit credential scan completes successfully before the commit continues

#### Scenario: Staged content contains a detected credential
- **WHEN** Gitleaks detects a credential in staged content
- **THEN** the pre-commit hook fails the commit
- **AND** the hook output does not reveal the full detected credential

### Requirement: Missing Gitleaks fails with installation guidance
The local pre-commit hook SHALL fail closed when Gitleaks is unavailable and SHALL print concise text that credential-leak analysis did not run because Gitleaks must be installed.

#### Scenario: Contributor lacks Gitleaks
- **WHEN** a contributor attempts a commit without Gitleaks on `PATH`
- **THEN** the pre-commit hook fails before commit completion
- **AND** output explains that no credential-leak analysis ran and Gitleaks must be installed
