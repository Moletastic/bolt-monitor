## MODIFIED Requirements

### Requirement: Infrastructure CI generates SST platform types
Infrastructure CI SHALL invoke the pinned SST CLI to generate platform types before infrastructure type checking. It SHALL pass the committed example target through the canonical target-path environment contract and invoke SST with that target's declared stage. It SHALL NOT invoke an undeclared package script for this work.

#### Scenario: Infrastructure CI runs release gates
- **WHEN** infrastructure CI installs dependencies before type checking
- **THEN** it passes the committed example target path to SST, runs `sst install` through pnpm with the matching `staging` stage, and generated platform types are available
