# OpenSpec Roadmap

This roadmap orders active capability work by security, data integrity, operator value, and dependency. It does not promise dates. Each change remains independently reviewed and applied through the OpenSpec workflow.

## Priority Definitions

| Priority | Meaning |
| --- | --- |
| P0 | Required before high-risk infrastructure and internet-boundary cutovers |
| P1 | Required before monitoring evidence and notification behavior can be treated as operationally trustworthy |
| P2 | Completes multi-user operation, interaction quality, and persisted health visibility |
| P3 | Establishes recoverability, bounded scale, and cost guardrails |
| P4 | Adds optional reliability objectives after observation data is trustworthy |

## Phase 0: Establish Deployment And Release Safety

| Priority | Change | Depends On | Exit Evidence |
| --- | --- | --- | --- |
| P0 | `standardize-stage-resource-lifecycle` | None | Every deployable stage is explicitly persistent or ephemeral; persistent state is protected and inventoried; ephemeral teardown leaves no stage-owned resources; AWS target identity is confirmed before mutation |
| P0 | `strengthen-repository-release-gates` local-foundation milestone | None | CI builds deployable artifacts and detects SST, handler, OpenAPI, Bruno, health-envelope, and authentication-metadata drift before protected-route refactoring |

Phase 0 lands basic persistent AppTable PITR/deletion protection, portable stage/profile validation, and the authentication-independent `strengthen-repository-release-gates` local foundation. Credentialed staging authentication smoke remains a separate later milestone, gated by both `add-single-tenant-operator-authentication` and `standardize-stage-resource-lifecycle`.

## Phase 1: Secure The Installation

| Priority | Change | Depends On | Exit Evidence |
| --- | --- | --- | --- |
| P0 | `harden-outbound-http-monitoring-boundaries` | Phase 0 | Monitor and notification HTTP clients reject all non-global/special-use destinations, unsafe redirects, stale connection reuse, DNS/dial changes, and oversized responses while safe public targets continue working |
| P0 | `add-single-tenant-operator-authentication` | `standardize-stage-resource-lifecycle`, release-gates local foundation | Every `/api/v1/**` route requires a scoped Cognito access token and authoritative AuthTable membership, custom dashboard authentication works, one administrator can be bootstrapped/recovered, and `/api/health` remains public |
| P0 | `strengthen-repository-release-gates` local-smoke milestone | `add-single-tenant-operator-authentication`, `standardize-stage-resource-lifecycle` | Opt-in operator-run local smoke proves public health and protected-route denial/acceptance against declared persistent staging without exposing secrets |

Both security changes must land before claiming an internet-facing installation has a production-oriented security boundary.

## Phase 2: Make Monitoring Evidence Trustworthy

| Priority | Change                                               | Depends On                                         | Exit Evidence                                                                                                                                                                                                                  |
| -------- | ---------------------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| P1       | `make-check-execution-retry-safe`                    | Phase 1                                            | One accepted or scheduled execution has one durable identity and schedule version; duplicate, partial, stale, out-of-order, and client-retried processing cannot create conflicting observations, incidents, or outbox records |
| P1       | `assure-notification-and-escalation-delivery`        | `make-check-execution-retry-safe` canonical outbox | One dispatcher delivers canonical outbox records; every channel has durable outcomes; retry/replay is idempotent; poison work reaches a DLQ; delayed schedules clean themselves up                                             |
| P1       | `expose-monitoring-pipeline-health` signal milestone | Corrected execution and notification contracts     | Structured correlation logs, scheduler heartbeat, queue-age and DLQ alarms, bounded default alarm inventory, and runbooks expose core pipeline failure without per-monitor resources                                           |

Phase 2 is complete only when a missing or duplicated check and a failed notification are visible and recoverable rather than silently improving or corrupting reported health.

## Phase 3: Complete Operator And Health Experience

| Priority | Change                                                          | Depends On                                                             | Exit Evidence                                                                                                                                                                       |
| -------- | --------------------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P2       | `add-operator-user-management-and-rbac`                         | `add-single-tenant-operator-authentication` canonical AuthTable schema | Administrators can safely invite, assign fixed roles, disable, and revoke users; `auth_time` revocation and last-active-admin protection hold without a second membership authority |
| P2       | `improve-dashboard-interaction-responsiveness`                  | Authenticated dashboard adapter available for protected actions        | Same-page mutations provide pending or narrowly optimistic feedback without unnecessary navigation, while server-rendered truth and router conventions remain intact                |
| P2       | `expose-monitoring-pipeline-health` persisted-summary milestone | Retry-safe/delivery states and bounded due-time projections            | Operators can distinguish target failure, delayed monitoring, and failed notification delivery; incomplete evidence returns `UNKNOWN`, never healthy                                |

These changes may proceed in domain-sized slices. Phase 2 reliability work remains higher priority than interaction polish or expanded administration.

## Phase 4: Establish Recovery And Capacity Boundaries

| Priority | Change                                            | Depends On                                             | Exit Evidence                                                                                                                                                                    |
| -------- | ------------------------------------------------- | ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P3       | `establish-data-recovery-and-capacity-guardrails` | Stage lifecycle and Phases 1-3 data authorities stable | AppTable/AuthTable recovery domains have explicit retention, restore-to-new-table drills pass, growing reads are bounded, and low-use/expected/stress cost profiles are measured |

Observed recovery timings are operational evidence, not an SLA, RPO, or RTO commitment.

## Phase 5: Add Honest Reliability Objectives

| Priority | Change                                    | Depends On                                                                                                                                                                                                       | Exit Evidence                                                                                                                                                                                                                                       |
| -------- | ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P4       | `define-availability-observation-windows` | `make-check-execution-retry-safe`, `assure-notification-and-escalation-delivery`, `expose-monitoring-pipeline-health`, `establish-data-recovery-and-capacity-guardrails`, `standardize-stage-resource-lifecycle` | Scheduler-independent finalization classifies every mature expected slot as good, bad, missing, or explicitly excluded; reports expose completeness and integer-precise raw counts; optional objectives never convert missing evidence into success |

This phase adds user-selected objectives, not contractual SLAs. Recent sample metrics remain separate from objective-window reporting.

## Explicit Deferrals

- Custom password or identity-provider implementation
- Machine-to-machine authentication
- Multi-tenant SaaS behavior
- Static dashboard migration
- Multi-region or user-selectable probe locations
- Public status pages
- PromQL, arbitrary metrics ingestion, log aggregation, tracing storage, or dashboard-builder behavior
- General on-call scheduling or incident-management platform behavior

Deferrals require a new OpenSpec proposal before implementation.
