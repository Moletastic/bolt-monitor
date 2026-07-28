## Context

Operations preflight a target path then invoke SST. SST independently derives a path from target name, creating two authorities.

## Goals / Non-Goals

**Goals:** Guarantee preflight, SST provider config, and stack policy load one file. Fail closed on ambiguous inputs.

**Non-Goals:** Support arbitrary untracked targets, weaken account/region preflight, or change public Make target names.

## Decisions

- Use one explicit resolved target path propagated from orchestrator to SST config, or remove `TARGET_FILE` entirely. Chosen implementation must preserve normal `TARGET=<name>` workflow.
- Validate canonical resolved path and target stage before invocation. Reject path/name disagreement.
- Use argument-array subprocesses and inherited environment only; no shell interpolation or added services/cost.

## Risks / Trade-offs

- External test fixtures need explicit path support -> Tests pass exact temporary path through same contract.
- Users relying on undocumented environment variable -> Fail with migration guidance instead of silently choosing another target.

## Migration Plan

Deploy resolver change and test mismatch cases. Existing named target workflow remains unchanged. No AWS migration.
