## ADDED Requirements

### Requirement: Orchestrator and SST use one resolved target file
Every credentialed infrastructure operation SHALL provide SST configuration the same validated target file used for account/region preflight. If target path and target name inputs identify different files or target identities, the operation SHALL fail before AWS mutation.

#### Scenario: Named target deployment
- **WHEN** an operator selects `TARGET=<name>`
- **THEN** preflight and SST stack evaluation load the same `infra/targets/<name>.target.json` file

#### Scenario: Conflicting target inputs are supplied
- **WHEN** a target path and target name resolve to different target identities
- **THEN** the operation fails before SST deploy, development, or removal begins
- **AND** the error identifies the conflicting non-secret target inputs
