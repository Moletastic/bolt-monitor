## Context

SST configuration now loads only `SST_TARGET_FILE` and rejects a requested SST stage that differs from the target's declared stage. CI still supplies only `TARGET=example` to `sst install`, leaving configuration unable to load its target.

## Goals / Non-Goals

**Goals:**
- Generate SST platform types in CI using the committed example target without AWS credentials.
- Make CI pass the same canonical target-path and stage values required by SST configuration.

**Non-Goals:**
- Run deployment, preflight, or AWS mutation in CI.
- Change target selection for operator workflows.

## Decisions

- Set `SST_TARGET_FILE` to `${{ github.workspace }}/infra/targets/example.target.json` so the CI path is absolute and independent of SST working-directory behavior.
- Invoke `sst install --stage staging`, matching the committed example target's declared stage.
- Set `SST_OPERATION=install` to identify this non-removal configuration evaluation; expired-target exceptions remain removal-only.

## Risks / Trade-offs

- Example target stage changes → CI command must change with its declared stage. The explicit value makes mismatch fail early.
- Direct SST bootstrap bypasses the orchestrator → CI passes only non-secret configuration needed for type generation and no AWS credentials.
