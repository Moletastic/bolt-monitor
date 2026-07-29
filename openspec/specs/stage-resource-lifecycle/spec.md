## Purpose

Define the stage classification, lifecycle protection, removal verification, and credentialed mutation contract for every deployable SST stage.

## Requirements

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

### Requirement: Persistent target budget configuration fails closed
When budget configuration is supplied, the target validator SHALL require a finite positive USD amount and one or more non-empty alert email addresses as one paired configuration. A persistent target SHALL require valid paired budget configuration unless it declares a documented explicit FinOps opt-out; malformed or partial fields SHALL fail validation before AWS mutation.

#### Scenario: Persistent target has valid budget configuration
- **WHEN** a persistent target provides positive budget amount and alert recipients
- **THEN** target validation succeeds and conditional budget infrastructure remains enabled

#### Scenario: Persistent target has malformed budget configuration
- **WHEN** a persistent target supplies only one budget field or an invalid amount or recipient list
- **THEN** validation fails before deploy
- **AND** it does not silently disable budget alerts

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

### Requirement: Orchestrator and SST use one resolved target file
Every credentialed infrastructure operation SHALL provide SST configuration the same validated target file used for account/region preflight. If target path and target name inputs identify different files or target identities, the operation SHALL fail before AWS mutation.

#### Scenario: Named target deployment
- **WHEN** an operator selects `TARGET=<name>`
- **THEN** preflight and SST stack evaluation load the same `infra/targets/<name>.target.json` file

#### Scenario: Conflicting target inputs are supplied
- **WHEN** a target path and target name resolve to different target identities
- **THEN** the operation fails before SST deploy, development, or removal begins
- **AND** the error identifies the conflicting non-secret target inputs

### Requirement: Expired ephemeral targets remain eligible for verified removal
The infrastructure orchestrator SHALL permit an expired target only when it is structurally valid, `ephemeral`, and `disposable=true`, and only for exact-stage removal plus residual verification. It SHALL reject that expired target before development, deployment, status, invitation, or credential/key mutation.

#### Scenario: Operator removes expired ephemeral stage
- **WHEN** an operator selects an otherwise valid expired disposable target for `make remove-infra`
- **THEN** the orchestrator runs the existing exact-stage SST removal and residual verification
- **AND** does not require expiry to be in the future

#### Scenario: Operator deploys expired ephemeral stage
- **WHEN** an operator selects an expired ephemeral target for deployment or development
- **THEN** validation fails before AWS mutation
- **AND** the error identifies expiry without exposing credentials

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

### Requirement: Infrastructure operations accept documented Makefile inputs
The infrastructure operation CLI SHALL accept the documented `KEY=value` arguments emitted by public Make targets and the equivalent `--KEY=value` form. Persistent removal SHALL require `DESTROY=yes` exactly, and administrator invitation SHALL require a non-empty `EMAIL` value.

#### Scenario: Persistent removal receives explicit intent
- **WHEN** an operator runs `DESTROY=yes make remove-infra` for a persistent target
- **THEN** the orchestrator receives destructive intent and may continue only after its existing target preflight

#### Scenario: Administrator invitation receives email
- **WHEN** an operator runs `make invite-admin EMAIL=operator@example.com`
- **THEN** the orchestrator passes that non-empty email to the invitation operation

#### Scenario: Destructive intent is absent or invalid
- **WHEN** a persistent removal omits `DESTROY=yes` or supplies another value
- **THEN** the orchestrator fails before SST removal begins

### Requirement: Persistent deployment verifies both DynamoDB authorities
After an explicit persistent deployment, the infrastructure orchestrator SHALL verify deletion protection and point-in-time recovery for the output-named `AppTable` and `AuthTable`. It SHALL perform only bounded control-plane reads for those exact tables.

#### Scenario: Persistent postflight succeeds
- **WHEN** a persistent target deploys and both output-named tables have deletion protection and point-in-time recovery enabled
- **THEN** postflight verification succeeds without scanning account resources

#### Scenario: AuthTable protection is missing
- **WHEN** persistent postflight finds the output-named `AuthTable` missing deletion protection or point-in-time recovery
- **THEN** the deployment command fails with the table name and missing protection state

### Requirement: Ephemeral cleanup reports bounded evidence
Ephemeral cleanup SHALL retain the exact-stage pre-removal SST inventory as non-secret evidence, verify SST no longer reports the stage deployed, and verify no ownership-tagged residual resources remain. It SHALL not claim provider-wide absence or add scheduled cleanup work.

#### Scenario: Ephemeral cleanup completes
- **WHEN** SST removal succeeds, exact-stage SST state is absent, and ownership-tagged residual inventory is empty
- **THEN** cleanup reports successful bounded verification and the captured SST inventory

#### Scenario: Cleanup evidence is incomplete
- **WHEN** SST state remains deployed, residual resources remain, or required state evidence cannot be obtained
- **THEN** cleanup fails and reports bounded non-secret diagnostic identifiers
