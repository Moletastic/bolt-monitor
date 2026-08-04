## MODIFIED Requirements

### Requirement: CI failures are actionable and cost-bounded
Repository CI SHALL report the failing validation surface clearly and SHALL avoid unnecessary deployments or duplicate local validation work in ordinary pull-request and `main` validation. The infrastructure validation surface SHALL execute every repository-owned root script test exactly once while retaining focused production validation commands for contract and policy checks.

#### Scenario: A release gate fails
- **WHEN** a Go, dashboard, infrastructure, Bruno, or API-contract check detects a violation
- **THEN** the workflow identifies the failed command or job
- **AND** the validator reports the affected file or normalized route and the expected correction where available

#### Scenario: Ordinary repository CI runs
- **WHEN** CI runs for a pull request or a push to `main`
- **THEN** it performs only local deterministic build and validation work
- **AND** it does not deploy SST resources
- **AND** repeated setup or validation is grouped or cached where doing so does not weaken isolation or reproducibility

#### Scenario: Root script tests run
- **WHEN** the infrastructure validation target runs
- **THEN** it executes every root `scripts/*.test.mjs` test once
- **AND** it does not execute the same root script test through a second Makefile prerequisite in that target set
- **AND** focused Makefile commands remain available to execute production contract and policy validators
