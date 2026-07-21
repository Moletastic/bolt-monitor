## MODIFIED Requirements

### Requirement: Developer workflow is documented for the SST bootstrap
The repository SHALL document one root `make setup` command that installs locked infrastructure and dashboard dependencies and synchronizes the Go workspace. It SHALL document `make deploy-infra` as the ordinary deployment command, `make infra-status` as the non-mutating target inspection command, `make invite-admin EMAIL=<email>` as the administrator bootstrap command, and explicit target selection only for advanced alternate targets. Every credentialed infrastructure workflow SHALL resolve one configured target file, show the effective non-secret AWS account, region, profile, owner, stage, and lifecycle class before mutation, and fail rather than infer a class or silently use an unconfigured stage.

#### Scenario: New contributor follows bootstrap instructions
- **WHEN** a contributor reads the bootstrap documentation
- **THEN** they can run `make setup`, create one `infra/targets/staging.target.json` from the committed example, deploy with `make deploy-infra`, and invite an initial administrator with only an email address
- **AND** they can distinguish normal staging deployment from an explicit developer-owned ephemeral target

#### Scenario: Contributor inspects a deployment target
- **WHEN** a contributor invokes `make infra-status` for a configured target
- **THEN** the command shows the effective account, region, AWS profile, stage, class, service, and owner without exposing credentials
- **AND** it stops on a target mismatch or incomplete classification

#### Scenario: Contributor omits target configuration
- **WHEN** a contributor runs a local, deploy, or remove workflow without the selected target file
- **THEN** the workflow fails with corrective guidance rather than selecting an implicit stage or lifecycle class
