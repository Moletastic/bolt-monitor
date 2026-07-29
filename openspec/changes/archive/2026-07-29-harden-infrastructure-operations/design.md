## Context

`make` exposes the supported SST operations through `infra/scripts/ops.mjs`. The orchestrator already resolves one validated target, confirms the effective AWS account and region, and uses argument-array subprocesses. Its CLI flag parser, however, accepts only `--key=value` while the public Make targets pass `KEY=value`. Persistent deployment verifies only `AppTable`, health verification accepts any successful HTTP response, and ephemeral cleanup records SST state before removal without using that inventory as evidence afterward.

The change must retain explicit target selection, persistent-removal intent, bounded operator-invoked checks, and the existing pay-for-use cost posture.

## Goals / Non-Goals

**Goals:**

- Make documented `DESTROY=yes` and `EMAIL=<address>` inputs reach their intended operations without accepting ambiguous values.
- Verify the persistent data authorities and the health envelope after an explicit deploy.
- Make ephemeral cleanup evidence match its success claim using bounded, exact-stage checks.
- Cover each corrected path through unit tests without requiring AWS credentials.

**Non-Goals:**

- Change target selection, lifecycle classes, resource retention policy, or budget policy.
- Add scheduled cleanup, broad account scans, custom metrics, or always-on resources.
- Add a general command-line parsing dependency or shell-based subprocess execution.
- Perform credentialed deployment testing in CI.

## Decisions

### Accept only documented key-value forms at the CLI boundary

The parser will recognize the public `KEY=value` form passed by Make and the existing `--KEY=value` form, then dispatch only explicit supported parameters. `DESTROY` remains effective only when its exact value is `yes`; `EMAIL` remains non-empty validated input to the invitation operation.

Alternative considered: change the Make targets to emit only `--KEY=value`. Rejected because external callers and documented commands already use Make-style assignment and the CLI can safely support both forms without ambiguity.

### Verify both persistent DynamoDB authorities after deploy

Postflight will resolve `appTableName` and `authTableName` from validated SST outputs, then confirm deletion protection and point-in-time recovery for each. These are bounded control-plane reads performed only after an explicit persistent deployment.

Alternative considered: verify only stack configuration. Rejected because deployed resource state is the operational contract and can drift from source configuration.

### Validate the health envelope from the existing postflight response

The deploy operation will retain one health request and parse its body as JSON, requiring the standard successful envelope and healthy data instead of only accepting a 2xx response.

Alternative considered: add a second contract-check request. Rejected because parsing the existing response provides stronger evidence with no additional request or cost.

### Use SST inventory as bounded cleanup evidence

Ephemeral removal will preserve the pre-removal SST inventory and verify that SST no longer reports the exact stage and that taggable resources have no exact-stage residuals. The inventory will be surfaced as evidence and used to identify an incomplete state transition; the implementation will not introduce broad provider scans or a recurring janitor.

Alternative considered: query every possible AWS resource API for each recorded URN. Rejected because it increases API surface, cost, and maintenance without improving the stage-level ownership contract beyond SST state plus exact ownership tags.

## Risks / Trade-offs

- [CLI compatibility accepts unintended keys] -> Dispatch only known operation parameters and retain strict value validation.
- [Postflight adds control-plane calls] -> Run them only for explicit persistent deploys and only for two named tables.
- [Health response changes shape] -> Fail deployment postflight with a clear contract error and retain handler/contract tests as the source of API behavior.
- [Provider cleanup leaves an untaggable orphan] -> Retain pre-removal SST inventory and exact-stage state verification; report bounded non-secret identifiers rather than claiming provider-wide proof.

## Migration Plan

1. Add parsing, deploy, health, and cleanup tests before changing behavior.
2. Update operations and cleanup code while retaining the public Make targets.
3. Run infrastructure tests and type/format checks.
4. Deploy only through the existing target preflight; rollback restores the script behavior without data migration or resource replacement.

## Open Questions

None.
