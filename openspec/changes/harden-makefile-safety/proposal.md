## Why

The root Makefile can silently accept failed Go checks, runs dependency installation for a bare `make`, and hardcodes the confirmation required to remove persistent infrastructure. These behaviors make routine automation surprising and weaken the repository's lifecycle safety boundary.

## What Changes

- Make a bare `make` show available commands rather than mutate local dependencies.
- Configure Make recipes to fail on command and pipeline errors, and remove partially created file targets after a failed recipe.
- Preserve `DESTROY=yes` as caller-supplied intent for persistent infrastructure removal.
- Document every public Make target through `make help`.
- Define the supported input contract for targeted dashboard and infrastructure formatting.
- Replace repeated Go Lambda build and packaging commands with one loop over the canonical service list.
- Render selected infrastructure target details as labeled multi-line output for operator commands.

## Capabilities

### New Capabilities

- `makefile-task-safety`: Safe, discoverable root task-runner behavior for local development and operational commands.

### Modified Capabilities

- `stage-resource-lifecycle`: Persistent removal continues to require separately supplied destructive intent.

## Impact

Affected areas include the root `Makefile`, its Make-target tests or validation, and lifecycle documentation. No application API, AWS resource topology, or dependency version changes are included.
