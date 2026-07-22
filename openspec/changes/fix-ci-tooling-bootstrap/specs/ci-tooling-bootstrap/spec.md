## ADDED Requirements

### Requirement: CI jobs bootstrap only required tooling
Each CI job SHALL install only tools and dependencies required by its invoked
release gates. Go-only Make targets SHALL NOT require pnpm or JavaScript
workspace installation.

#### Scenario: Backend CI runs Go gates
- **WHEN** the backend CI job runs Go format, vet, test, lint, and build gates
- **THEN** those gates complete without a pnpm executable

### Requirement: Infrastructure CI invokes declared scripts only
Infrastructure CI SHALL invoke only scripts declared in `infra/package.json`.

#### Scenario: Infrastructure CI runs release gates
- **WHEN** infrastructure CI installs dependencies and runs its release gates
- **THEN** it does not invoke an undeclared package script
