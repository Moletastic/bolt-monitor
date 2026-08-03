## Context

Delivery listing loads all records then truncates to 200. Existing history endpoints already encode and validate opaque DynamoDB cursors.

## Goals / Non-Goals

**Goals:** bounded delivery reads, stable traversal, no hidden records.

**Non-Goals:** paginate every current list, provide total counts, or change delivery ordering.

## Decisions

- Reuse opaque signed/resource-scoped cursor helpers and DynamoDB `LastEvaluatedKey`.
- Default and maximum limits bound read cost. Ordering is creation time then delivery ID.
- Return envelope cursor pagination. No total-count query.

## Risks / Trade-offs

- Cursor format becomes API contract -> treat as opaque and validate resource identity.
- Concurrent inserts can affect later pages -> stable DynamoDB order and tie-breaker prevent duplicates for supplied continuation key.

## Migration Plan

Deploy response continuation metadata before dashboard pagination UI. Existing clients retain first bounded page but now see continuation metadata.

## Open Questions

None.
