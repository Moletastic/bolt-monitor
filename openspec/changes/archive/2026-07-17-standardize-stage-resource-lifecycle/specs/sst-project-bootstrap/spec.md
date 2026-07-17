## MODIFIED Requirements

### Requirement: Developer workflow is documented for the SST bootstrap
The repository SHALL document the commands and prerequisites required to install dependencies, run SST locally, preview, deploy, remove, and understand where future infrastructure changes belong. Every local or credentialed infrastructure workflow SHALL require an explicitly configured persistent or ephemeral stage, SHALL show the effective non-secret AWS account, region, profile or credential source, owner, and lifecycle class before mutation, and SHALL fail rather than infer a class or silently use an omitted stage.

#### Scenario: New contributor follows bootstrap instructions
- **WHEN** a contributor reads the bootstrap documentation
- **THEN** they can identify the required tooling and the commands needed to start working with the SST project
- **AND** they can distinguish approved persistent staging from developer-owned ephemeral operation

#### Scenario: Contributor previews a deployment target
- **WHEN** a contributor supplies explicit stage lifecycle and AWS target configuration
- **THEN** the documented preflight shows the effective account, region, credential-source label, stage, class, service, and owner without exposing credentials
- **AND** the preview stops on a target mismatch or incomplete classification

#### Scenario: Contributor omits stage lifecycle configuration
- **WHEN** a contributor runs a local, preview, deploy, or remove workflow without a complete explicit stage target
- **THEN** the workflow fails with corrective guidance rather than selecting `staging`, a local stage, or a lifecycle class implicitly
