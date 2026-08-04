## Context

`README.md` is the contributor-facing summary of repository commands and layout. Its infrastructure test description and directory entries no longer fully match the current Makefile and source tree.

## Goals / Non-Goals

**Goals:**

- Align README commands and layout descriptions with checked-in repository structure.
- Keep existing terminology and avoid introducing a new top-level taxonomy.

**Non-Goals:**

- Change Makefile targets, module boundaries, or runtime behavior.
- Redesign the README or duplicate detailed guidance from `AGENTS.md`.

## Decisions

- Treat the Makefile and current directory ownership as source of truth for command and layout wording.
- Describe `shared/` by responsibility, not as domain-only code, because it also contains backend-platform modules.

## Risks / Trade-offs

- README wording can drift again as commands change → Keep command descriptions aligned when Makefile targets change.
