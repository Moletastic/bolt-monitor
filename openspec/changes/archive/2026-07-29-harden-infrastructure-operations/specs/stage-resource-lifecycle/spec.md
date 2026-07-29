## ADDED Requirements

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
