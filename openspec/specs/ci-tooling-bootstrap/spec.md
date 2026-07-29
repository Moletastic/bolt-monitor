## Purpose

Define the minimal tooling bootstrap required for each CI release-gate surface.

## Requirements

### Requirement: CI jobs bootstrap only required tooling
Each CI job SHALL install only tools and dependencies required by its invoked
release gates. Go-only Make targets SHALL NOT require pnpm or JavaScript
workspace installation.

#### Scenario: Backend CI runs Go gates
- **WHEN** the backend CI job runs Go format, vet, test, lint, and build gates
- **THEN** those gates complete without a pnpm executable

### Requirement: Infrastructure CI generates SST platform types
Infrastructure CI SHALL invoke the pinned SST CLI to generate platform types
before infrastructure type checking. It SHALL pass the committed example target
through the canonical target-path environment contract and invoke SST with that
target's declared stage. It SHALL NOT invoke an undeclared package script for
this work.

#### Scenario: Infrastructure CI runs release gates
- **WHEN** infrastructure CI installs dependencies before type checking
- **THEN** it passes the committed example target path to SST, runs `sst install`
  through pnpm with the matching `staging` stage, and generated platform types
  are available
