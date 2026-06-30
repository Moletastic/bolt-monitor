## Context

The repository already has a small API surface implemented through SST route declarations in `infra/` and Go Lambda handlers in `services/`, but it does not have a source-controlled API contract or a local docs workflow. The design should add useful OpenAPI documentation without changing deployed API behavior or introducing a heavyweight generation pipeline.

## Goals / Non-Goals

**Goals:**
- Add one source-controlled OpenAPI document that describes the current HTTP API.
- Render that same document locally in both Swagger UI and Redoc.
- Keep docs tooling isolated from application runtime packages.
- Provide one clear local command for viewing API docs.

**Non-Goals:**
- No API Gateway import or OpenAPI-driven deployment.
- No server stub generation or Go schema generation.
- No contract-test or CI enforcement in this change.
- No auth redesign or public developer platform concerns beyond documenting the current API honestly.

## Decisions

### Keep the OpenAPI source in a dedicated `openapi/` workspace
- Decision: place `openapi.yaml` and docs-specific tooling under a top-level `openapi/` directory.
- Rationale: keeps contract/docs concerns separate from `infra/`, `apps/dashboard`, and Go service modules; minimizes cross-package coupling and keeps docs easy to run or remove.
- Alternative considered: attach docs tooling to `infra/` or `apps/dashboard`.
- Why not: those packages have different responsibilities and would make API docs feel like an implementation detail of one runtime surface.

### Use one `openapi.yaml` as the source for both Swagger UI and Redoc
- Decision: maintain a single hand-authored OpenAPI document and feed both Swagger UI and Redoc from it.
- Rationale: one source file is enough for the current API size and avoids duplicate documentation formats.
- Alternative considered: separate Swagger-specific and Redoc-specific inputs, or generate OpenAPI from Go code.
- Why not: separate inputs would drift immediately; code generation is awkward with the current Lambda event-switch handler shape and adds more tooling than value.

### Serve docs locally through simple static tooling
- Decision: run Swagger UI and Redoc as local static docs pages that read the checked-in OpenAPI file.
- Rationale: this satisfies the local documentation goal with low maintenance and no impact on deployed infrastructure.
- Alternative considered: host docs inside the Next.js dashboard or expose them from the SST API.
- Why not: embedding docs into app runtime couples docs to unrelated delivery paths and adds deployment questions that are out of scope.

### Provide one obvious local command surface
- Decision: expose a small npm-based command surface inside `openapi/`, with one primary command for opening docs locally and companion commands for specific views if needed.
- Rationale: the repo root currently has no shared package manifest, so a dedicated docs package keeps setup explicit while still giving developers a consistent workflow.
- Alternative considered: add a root Makefile or root npm workspace command.
- Why not: those are viable, but a self-contained `openapi/` package is the smallest change that does not assume new repo-wide tooling conventions.

## Risks / Trade-offs

- [Risk] The OpenAPI document can drift from the actual SST route table or Lambda handler behavior. -> Mitigation: scope the spec to the current endpoints only and document it as the single checked-in contract for the API.
- [Risk] Local docs may imply features such as auth or multiple server environments that the project does not fully support yet. -> Mitigation: document only the current behavior and keep server/auth sections minimal and accurate.
- [Risk] Adding two docs renderers can create redundant maintenance. -> Mitigation: both renderers consume the same `openapi.yaml`, so only the presentation layer duplicates, not the content.

## Migration Plan

1. Add `openapi/` workspace files, including `openapi.yaml` and docs tooling configuration.
2. Describe the existing API routes and schemas in the OpenAPI document.
3. Add local Swagger UI and Redoc entry pages that point to that document.
4. Add a documented command for launching docs locally.
5. Update repo documentation with the docs workflow.

Rollback: remove the `openapi/` workspace and the related documentation changes. Runtime API behavior remains unchanged.

## Open Questions

- Should the OpenAPI document advertise only a localhost-style server example, or also include a placeholder deployed server entry?
- Should the primary local docs command open a single landing page linking to both renderers, or directly launch one preferred renderer by default?
