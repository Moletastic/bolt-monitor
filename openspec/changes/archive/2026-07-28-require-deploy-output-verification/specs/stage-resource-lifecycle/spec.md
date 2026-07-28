## ADDED Requirements

### Requirement: Deploy postflight verifies target-scoped outputs and health
After SST deploy, the orchestrator SHALL resolve outputs unambiguously for the selected target and require expected non-secret deployment identifiers before reporting success. It SHALL call the target API public health endpoint and fail the deploy command when outputs are missing, malformed, stage-mismatched, or health is unreachable.

#### Scenario: Deploy outputs match selected stage
- **WHEN** a selected target deploy completes and its required outputs are present
- **THEN** the orchestrator verifies the selected API health endpoint before reporting success

#### Scenario: Outputs are missing or belong to another stage
- **WHEN** SST output data cannot be unambiguously associated with selected target
- **THEN** deploy postflight fails
- **AND** it does not report health or protection verification for another stage
