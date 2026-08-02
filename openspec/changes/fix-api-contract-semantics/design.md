## Context

The common envelope requires failures to declare `status:error`, but channel deletion returns a `409` success envelope. OpenAPI route parity exists without complete response semantics.

## Goals / Non-Goals

**Goals:** runtime, OpenAPI, Bruno, and release gates agree on operation semantics.

**Non-Goals:** RFC 7807 migration, route redesign, or generated SDKs.

## Decisions

- Add a typed channel-in-use error mapped to `409`; safe route references live in error details.
- Keep the existing envelope as canonical API shape. OpenAPI gains separate success/error schemas.
- Contract checks validate static semantic conventions. Handler tests remain source of runtime proof.

## Risks / Trade-offs

- Better schemas require maintenance -> gate changes with tests and existing route parity command.
- Clients relying on false success conflict body break -> HTTP status already signaled failure; stable error semantics correct contract.

## Migration Plan

Publish OpenAPI and Bruno with runtime change. Dashboard branches on existing error envelope. Rollback restores previous handler only if required.

## Open Questions

None.
