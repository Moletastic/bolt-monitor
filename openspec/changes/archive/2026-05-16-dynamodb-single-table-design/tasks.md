## 1. Define Single-Table Item Families

- [x] 1.1 Document canonical PK/SK patterns for tenant, monitor, status, run, incident, and audit item families.
- [x] 1.2 Define required duplicated attributes and typed item metadata needed for access-pattern-driven reads.

## 2. Define Query And Index Strategy

- [x] 2.1 Document primary access patterns for monitor listings, monitor detail, status reads, incident views, and audit history.
- [x] 2.2 Define initial GSIs required for open incidents and dashboard status reads.

## 3. Define Retention And Integration Boundaries

- [x] 3.1 Specify TTL or retention expectations for high-volume `CheckRun` items.
- [x] 3.2 Align shared model-to-record mapping work with the single-table storage contract for future implementation.
