## REMOVED Requirements

### Requirement: Repository provides an opt-in local staging release smoke helper
**Reason**: Credentialed staging smoke requires local access tokens and sometimes MFA, creating release-process overhead beyond this project's scope.

**Migration**: Run deterministic repository checks before deployment. `make deploy-infra` performs non-secret deployment postflight checks for target identity, outputs, persistent protections, and public health. Verify authenticated dashboard behavior manually when authentication changes require it.

### Requirement: Staging smoke uses a retention-safe lifecycle
**Reason**: The repository no longer provides a staging smoke helper.

**Migration**: Use explicit persistent or ephemeral target lifecycle commands for deployment and removal. Do not create a smoke target as a replacement release process.

### Requirement: Staging smoke proves public health and protected authentication behavior
**Reason**: Authenticated smoke automation is not required for this small self-hosted project.

**Migration**: `make deploy-infra` verifies public health. Perform protected authentication checks manually after relevant changes.

### Requirement: Staging smoke protects credentials and bounds cloud cost
**Reason**: Removing the smoke helper removes its credentialed execution path and supporting cloud-operation concern.

**Migration**: Keep credentials in standard AWS profile configuration and retain target lifecycle checks and verified ephemeral removal.
