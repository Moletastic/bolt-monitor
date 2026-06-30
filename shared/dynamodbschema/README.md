# DynamoDB Single-Table Schema Contract

Shared storage contract for Bolt Monitor application data.

## Primary table

- Table name reference: `bolt-monitor-app`
- One primary table for core monitoring application data
- Typed items distinguished by `entityType`

## Core item families

- `Workspace`: `PK=TENANT#<tenantId>`, `SK=META`
- `Monitor`: `PK=MONITOR#<monitorId>`, `SK=META`
- `MonitorRef`: `PK=TENANT#<tenantId>`, `SK=MONITOR#<monitorId>`
- `MonitorStatus`: `PK=MONITOR#<monitorId>`, `SK=STATUS`
- `CheckRun`: `PK=MONITOR#<monitorId>`, `SK=RUN#<startedAt>#<runId>`
- `AlertState`: `PK=MONITOR#<monitorId>`, `SK=ALERT_STATE`
- `Incident`: `PK=MONITOR#<monitorId>`, `SK=INCIDENT#<openedAt>#<incidentId>`
- `AuditEvent`: `PK=TENANT#<tenantId>`, `SK=AUDIT#<timestamp>#<auditId>`
- `AuditChange`: `PK=AUDIT#<auditId>`, `SK=CHANGE#<fieldPath>`
- `IncidentActivity`: `PK=INCIDENT#<incidentId>`, `SK=ACTIVITY#<timestamp>#<activityId>`

## Initial GSIs

- `gsi1`: open incidents by tenant
- `gsi2`: dashboard status reads by tenant

## Retention

- `CheckRun` items are high-volume and should carry TTL metadata
- default retention target: 30 days until rollup strategy lands

## Core access patterns

- list monitors for tenant
- get monitor config
- get monitor status
- get recent runs for monitor
- get open incidents for tenant
- get dashboard status view for tenant
- get audit history for tenant
