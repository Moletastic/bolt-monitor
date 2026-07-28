## MODIFIED Requirements

### Requirement: Credentialed mutations confirm the effective AWS target
Before deploy, removal, import, adoption, or protection changes, the infrastructure orchestrator SHALL set AWS profile and region from the selected target file, resolve the effective AWS caller account and region, and compare them with explicit expected configuration. It SHALL present application, target name, stage, lifecycle class, owner, account, region, and profile without printing credentials. Ordinary deployment is confirmed by explicit invocation of `make deploy-infra`; persistent removal or protection changes SHALL require separately caller-supplied `DESTROY=yes` destructive intent and SHALL NOT receive that intent from a Makefile default.

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
- **AND** tooling requires caller-supplied `DESTROY=yes` destructive intent for the identified persistent target and resources
