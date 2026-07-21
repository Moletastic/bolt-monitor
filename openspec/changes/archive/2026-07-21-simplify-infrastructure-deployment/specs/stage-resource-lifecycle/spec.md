## MODIFIED Requirements

### Requirement: Every deployable stage has an explicit lifecycle class
The infrastructure system SHALL classify every deployable SST stage as exactly one of `persistent` or `ephemeral` from one validated target file at `infra/targets/<name>.target.json` before evaluating or mutating AWS resources. The target file SHALL declare the stage, AWS profile, expected AWS account, expected AWS region, owner, service, lifecycle class, dashboard origin, and required class-specific configuration. The default ordinary target SHALL be `infra/targets/staging.target.json`; `TARGET=<name>` SHALL select another target file. The system SHALL NOT infer lifecycle class, silently fall back to an unconfigured stage, or accept missing or contradictory target configuration.

#### Scenario: Persistent stage is approved
- **WHEN** a caller runs an infrastructure command with a complete persistent target file
- **THEN** infrastructure evaluation uses the persistent resource policy
- **AND** reports the target name, stage, class, owner, service, expected account, expected region, and AWS profile without exposing credentials

#### Scenario: Ephemeral stage is explicit
- **WHEN** a caller selects a complete ephemeral target file with `disposable=true`
- **THEN** infrastructure evaluation uses the ephemeral resource policy
- **AND** reports its cleanup or expiration deadline

#### Scenario: Classification is absent or inconsistent
- **WHEN** the selected target file is absent, malformed, has a missing AWS identity, has an unknown class, or conflicts with its disposal or approval configuration
- **THEN** validation fails before any AWS resource mutation
- **AND** the error identifies the invalid configuration without exposing credentials

### Requirement: Credentialed mutations confirm the effective AWS target
Before deploy, removal, import, adoption, or protection changes, the infrastructure orchestrator SHALL set AWS profile and region from the selected target file, resolve the effective AWS caller account and region, and compare them with explicit expected configuration. It SHALL present application, target name, stage, lifecycle class, owner, account, region, and profile without printing credentials. Ordinary deployment is confirmed by explicit invocation of `make deploy-infra`; persistent removal or protection changes SHALL require separate destructive intent from ordinary deployment.

#### Scenario: Effective target matches for deploy
- **WHEN** an operator invokes `make deploy-infra` and the resolved caller account and region match the selected target file
- **THEN** the requested deployment may proceed without a separately copied confirmation value
- **AND** no credential or secret value is printed

#### Scenario: Account or region differs
- **WHEN** the resolved AWS account or region differs from the selected target file
- **THEN** tooling fails before resource mutation with the mismatched non-secret identifiers

#### Scenario: Persistent resource destruction is requested
- **WHEN** an operator requests persistent removal or disables a protection
- **THEN** ordinary deployment intent is insufficient
- **AND** tooling requires explicit destructive intent for the identified persistent target and resources

### Requirement: Local, staging, and credentialed smoke workflows declare lifecycle intent
The supported workflows SHALL document `staging.target.json` as the named persistent target only when its file explicitly approves long-lived shared validation. Local SST development SHALL explicitly select either `staging.target.json` or a developer-owned `TARGET=<name>` ephemeral target and SHALL NOT gain a lifecycle class from an omitted or unconfigured target. The repository SHALL NOT provide a credentialed staging smoke workflow or require staging token, invitation, or MFA automation as a deployment gate.

#### Scenario: Developer starts local SST
- **WHEN** a developer starts SST local mode
- **THEN** the command resolves a configured persistent or ephemeral target file
- **AND** an ephemeral local target is subject to the same verified cleanup contract

#### Scenario: Ordinary staging deployment completes
- **WHEN** the configured persistent staging target deploys successfully
- **THEN** the deployment verifies its outputs, persistent protections, and public health
- **AND** it does not obtain authentication credentials or run credentialed smoke checks
