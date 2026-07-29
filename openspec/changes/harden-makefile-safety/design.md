## Context

The root Makefile is repository's public task-runner surface. It currently defaults to `setup`, expands multi-module Go checks into semicolon-separated shell commands, and passes `DESTROY=yes` directly to infrastructure removal. Formatting helpers accept whitespace-delimited `FILES` values without stating that constraint.

## Goals / Non-Goals

**Goals:**
- Make bare invocation non-mutating and discoverable.
- Fail the invoking target if any recipe command or pipeline fails.
- Retain caller-supplied destructive confirmation for persistent removal.
- State targeted formatting input constraints in command help.

**Non-Goals:**
- Replace Make with another task runner.
- Add file-level incremental build tracking or parallel execution.
- Support source paths containing whitespace or single quotes in `FILES`.
- Change infrastructure orchestration or lifecycle classification.

## Decisions

### Use Make-level strict Bash execution

The Makefile will set Bash as its shell and use fail-fast, unset-variable, and pipeline-failure flags. This makes existing generated Go command sequences stop at first failed module without duplicating loop logic. It will also enable `.DELETE_ON_ERROR` so future file targets are not left as valid-looking outputs after a failed recipe.

Alternative: rewrite each Go command list as an explicit shell loop. Rejected because strict recipe behavior is needed across all targets, not only Go validation.

### Make help the default public action

The default goal will be `help`, which lists documented public targets and their descriptions. Commands retaining mutation stay explicit targets.

Alternative: retain `setup` as default. Rejected because a bare command must not install dependencies or synchronize workspace state.

### Forward removal confirmation unchanged

`remove-infra` will pass its caller's `DESTROY` value to the orchestrator, which already validates `DESTROY=yes` for persistent targets. The Makefile will not inject approval.

Alternative: add separate persistent and ephemeral Make targets. Rejected because target lifecycle is resolved from the selected target file and the existing orchestrator owns validation.

### Declare `FILES` as whitespace-delimited

Formatting target help will state that `FILES` is a whitespace-delimited list and paths containing whitespace or single quotes are unsupported. This matches Make word-list semantics and avoids fragile quoting transformations.

Alternative: create a shell wrapper with another transport format. Rejected as disproportionate for repository paths and current developer workflow.

### Build Lambda artifacts from the canonical service list

`build-go` will iterate `GO_SERVICES`, which already lists each Lambda service built today. Each iteration will cross-compile the handler and package its `function.zip`; strict Bash execution stops the target on either failure. `clean` will use the same list so artifact cleanup cannot drift from the build set.

Alternative: add Make pattern rules for handler and zip files. Rejected because no source-file prerequisites exist today, so they would imply incorrect incremental freshness behavior or extra maintenance.

### Render infrastructure target details one field per line

The infrastructure orchestrator will format its selected target summary with one labeled field per line and an action heading. This preserves every non-secret detail already printed while making `make infra-status` readable in terminals and build logs. Deploy, development, removal, and invitation output will use the same formatter.

## Risks / Trade-offs

- [Existing recipe relies on an unset variable] -> strict mode surfaces failure; validate all documented targets before merge.
- [A developer omits `DESTROY=yes` for persistent removal] -> removal fails before mutation, which is intended.
- [Target description comments drift] -> `make help` output is validated in automated checks.
- [A service needs a different artifact layout] -> give it an explicit target before removing it from `GO_SERVICES`.

## Migration Plan

1. Add Make defaults, documented target comments, and caller-supplied removal confirmation.
2. Add a focused validation script or test for default help, strict shell configuration, and removal forwarding.
3. Run documented Make dry-runs and existing relevant checks.

Rollback reverts only Makefile and its validation; no deployed infrastructure or persisted data changes.

## Open Questions

None.
