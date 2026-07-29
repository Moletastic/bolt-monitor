## Why

The SST orchestrator safely validates deployment targets, but Makefile arguments are not reaching its CLI and its postflight verification does not fully prove the documented protection and health contracts. Ephemeral cleanup also overstates its residual-resource evidence.

## What Changes

- Accept the documented Makefile `DESTROY=yes` and `EMAIL=<address>` arguments in infrastructure operations, while retaining explicit persistent-removal intent.
- Verify both persistent DynamoDB tables have deletion protection and point-in-time recovery after deployment.
- Validate the public health endpoint's standard success envelope after deployment.
- Strengthen ephemeral cleanup evidence using the pre-removal SST resource inventory without adding recurring AWS work.

## Capabilities

### New Capabilities

### Modified Capabilities
- `stage-resource-lifecycle`: Strengthen supported operation inputs, persistent postflight checks, and ephemeral cleanup verification.
- `api-health-endpoint`: Require deploy verification to validate the public health response envelope.

## Impact

- Affects `infra/scripts/ops.mjs`, `infra/scripts/cleanup.mjs`, their tests, and lifecycle documentation.
- Adds bounded control-plane reads only during explicit deploy or removal; it adds no scheduled resources, custom metrics, or always-on compute.
