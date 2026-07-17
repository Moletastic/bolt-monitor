## Why

Bolt Monitor currently accepts arbitrary SST stage names without classifying whether their resources are durable or disposable, so routine deploy, smoke, and removal workflows can either delete important state or strand retained billable resources. A single explicit lifecycle contract is needed before persistent storage protection and retained authentication resources land.

## What Changes

- Define two SST stage classes, `persistent` and `ephemeral`, with explicit configuration and fail-closed classification; unknown, incomplete, or contradictory stage configuration does not deploy.
- Require persistent stages to use approved names, identify owner/stage/service, protect applicable DynamoDB tables with point-in-time recovery and deletion safeguards, retain durable Cognito and SSM/SST secret resources, expose a retained-resource inventory, and support deliberate cleanup and infrastructure re-adoption.
- Require ephemeral stages to opt in as disposable, reject protected production names, disable retention and deletion protection, bound resource lifetime where supported, and remove all stage-owned tables, user pools, parameters, schedules, queues, buckets, and other resources deterministically.
- Make account, region, profile, stage, and lifecycle class visible and explicitly confirmed for credentialed deployment and cleanup workflows while preserving a low-friction open-source setup path.
- Preserve the existing named `staging` workflow as a deliberately configured stage and distinguish SST local development expectations from deployable stage lifecycle behavior.
- Require credentialed release smoke to use either an explicitly ephemeral stage with verified cleanup or an approved long-lived persistent staging stage, never a unique retained stage.
- Land the shared lifecycle policy and basic table protection before authentication cutover; leave detailed restore validation and recovery drills to `establish-data-recovery-and-capacity-guardrails`.
- Prevent orphaned fixed- and usage-based AWS resources without introducing a new service.

## Capabilities

### New Capabilities

- `stage-resource-lifecycle`: Explicit persistent and ephemeral stage classification, resource protection and disposal policy, deployment identity confirmation, retained-resource inventory, cleanup, and re-adoption requirements.

### Modified Capabilities

- `sst-project-bootstrap`: Make stage lifecycle configuration and explicit AWS target confirmation part of the supported SST developer and deployment workflow.

## Impact

- Affects SST app configuration, bootstrap resource options and outputs, deployment/removal/smoke workflows, infrastructure validation, and operator documentation.
- Establishes lifecycle prerequisites consumed by the active authentication, recovery/capacity, and release-smoke changes without implementing their application behavior or detailed recovery drill.
- Changes arbitrary-stage deployment from permissive behavior to fail-closed validation. Existing `staging` and local workflows require an explicit documented classification rather than an implicit fallback.
- Adds no application API behavior and no new AWS service; cost impact is limited to protection already planned for persistent resources and reduced orphan risk for disposable stages.
