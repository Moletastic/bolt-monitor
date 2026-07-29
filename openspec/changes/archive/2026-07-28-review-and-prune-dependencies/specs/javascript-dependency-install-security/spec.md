## ADDED Requirements

### Requirement: Dependency changes preserve install trust controls
Any dependency addition or removal in an in-scope JavaScript package root SHALL preserve committed lockfile reproducibility and the explicit install-script trust allowlist. The review SHALL update the allowlist and its documentation only when the dependency change changes an approved install-script exception.

#### Scenario: Dashboard dependency is removed
- **WHEN** the dashboard removes a direct dependency
- **THEN** its pnpm lockfile is updated without unrelated resolution drift
- **AND** existing install-script allowlist entries remain unchanged unless the removed package owns that exception

#### Scenario: Dependency review changes an install-script exception
- **WHEN** review adds or removes an allowlisted install-script dependency
- **THEN** the package root's allowlist and matching trust documentation are updated together
- **AND** immutable installation succeeds with only approved scripts allowed
