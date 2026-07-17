# Bolt Monitor

Bolt Monitor is a self-deployed uptime and incident platform built on AWS serverless. It schedules HTTP health checks, stores execution and status history, manages incident lifecycle, and exposes an operator dashboard without requiring a permanently running monitoring server.

It is targeted at small teams that want to own their monitoring stack on AWS and prefer predictable, pay-for-execution infrastructure over a long-lived control plane.

## Why Bolt Monitor

- AWS-native deployment: API Gateway, Lambda, DynamoDB, EventBridge, SQS, Cognito, and CloudFront wired through a single SST stack.
- Service-scoped data model: services own monitors, runs, incidents, audit, and escalation policies.
- Scheduled HTTP checks with execution and status history.
- Incident acknowledgement, resolution, and escalation against persisted notification channels.
- Operator dashboard backed by server-side API fetches; no client-side data shadowing.
- Honest scope: single-tenant today, single-region execution, Cognito-authenticated dashboard and API.

## Capabilities

- Service-scoped monitor creation, editing, enable and disable, maintenance mode, manual run, archive and reactivate.
- Recurring HTTP checks with configurable intervals, expected status, and body assertions.
- Run history and latest status, monitor audit trail, and service audit trail.
- Incident list, detail, acknowledge, resolve, and escalation-state inspection.
- Escalation policies and notification channels (webhook, email, SMS, PagerDuty, Telegram).
- Scheduler configuration endpoints for cadence and dispatch policy.
- Operator dashboard with sidebar navigation, monitor and incident workflows, audit timeline, and admin scheduler view.
- Public `GET /api/health` and invite-only Cognito authentication for `/api/v1/**`.
- Checked-in OpenAPI contract with deterministic SST, Bruno, OpenAPI, and handler-route drift gates.
- Bruno API collection with explicit per-route auth classification and locally stored credentials.

Detailed route inventories live in `openapi/openapi.yaml` and `.bruno/collections/`.

## How it works

1. The operator creates a service and a monitor in the dashboard or directly through the API.
2. EventBridge invokes the scheduler Lambda once per minute per active monitor.
3. The scheduler computes due work and enqueues it on the execution SQS queue.
4. The worker Lambda consumes the queue and performs the HTTP check against the monitor target.
5. The worker persists the result on the DynamoDB single table and updates monitor status.
6. Repeated failures create or update an incident; ack and resolve transition its state.
7. The dashboard renders status, run history, incident detail, and audit entries by calling the API through server-side fetches.

The escalation runtime consumes the notification queue and dispatches through the configured channel registry.

## Architecture

```
Browser
  |
  v
Next.js dashboard (apps/dashboard, deployed via SST)
  |
  | server-side fetches and server actions
  v
API Gateway V2 (infra/stacks/bootstrap.ts)
  |
  +-- GET /api/health             -> services/api-health
  +-- /api/v1/**                  -> services/monitor-api  (Cognito JWT authorizer)
  +-- /api/v1/auth/**             -> services/monitor-api
                                       |
                                       v
                                    DynamoDB (AppTable, AuthTable)

EventBridge cron (rate 1 minute)
  |
  +-- scheduler mode -> services/check-runtime
  +-- worker mode    -> services/check-runtime
                                  |
                                  v
                              DynamoDB (AppTable)

Execution SQS queue           Notification SQS queue
        |                              |
        v                              v
services/check-runtime       services/escalation-runtime
                                       |
                                       v
                                notification senders
                                (webhook, email, SMS, PagerDuty, Telegram)
```

The dashboard ships as an OpenNext-built Next.js app served through CloudFront.

### Design choices

| Choice | Rationale |
| --- | --- |
| Go for Lambda services | Compact cold-start footprint, static binaries, single-language shared domain modules. |
| TypeScript for infrastructure and UI | SST authoring and a single frontend toolchain. |
| DynamoDB single-table design | Keeps persistence serverless and removes a long-lived control plane. |
| EventBridge Scheduler for dispatch | Managed, per-minute invocation without self-hosted schedulers. |
| SQS between scheduler and workers | Bounded retries with DLQs, no in-process queue. |
| Server-side dashboard fetches | Authoritative state, no client-side data shadowing. |
| OpenAPI and Bruno for contracts | Drift gates compare SST, Bruno, OpenAPI, OpenSpec, and the monitor handler inventory. |
| Cognito user pool for authentication | Invite-only operators, scoped JWT access tokens, AuthTable-backed authorization. |
| SST for deployment | Single source of truth for AWS wiring; lifecycle guard rejects production or ambiguous targets. |

## Project status

Functional today:

- Public `GET /api/health` returns the standard success envelope.
- Service-scoped monitor lifecycle, runs, status, audit, and incident flows.
- Invitation-only Cognito authentication for the dashboard and the versioned API.
- Recurring and manual HTTP execution through the worker Lambda.
- Notification escalation through registered channels.
- Local deterministic release gates: Go, dashboard, infrastructure, Bruno, and OpenAPI contract drift.

Known limitations:

- Single built-in tenant ID (`DEFAULT`) and single-region execution.
- No user-management or RBAC beyond `ADMIN` membership activation.
- No multi-region probe execution, public status page, or PromQL/log ingestion.
- AWS credentials remain required to deploy or run a staging environment; the dashboard itself does not call AWS directly.

Production-readiness caveats:

- No promised SLA, RPO, or RTO. Recovery drills and capacity guardrails are pending.
- Backups and retention policies apply to persistent tables and Cognito; ephemeral smoke stages are cleaned up after teardown.
- Security boundary is now internet-facing with bearer-token-protected API and authenticated dashboard; full SSRF and outbound-boundary hardening is in progress.

## Quick start

Bolt Monitor targets Node.js 22, pnpm 10, Go 1.26+, and SST 4.14.

1. Install dependencies with pnpm from each workspace root.

   ```bash
   cd infra && pnpm install --frozen-lockfile
   cd ../apps/dashboard && pnpm install --frozen-lockfile
   ```

2. Run the local release gates.

   ```bash
   make test-go-all
   make check-infra
   make lint-dashboard
   make check-bruno check-api-contract
   ```

3. Configure an explicit deployment target outside source control. Use `infra/deployment-target.example.json` as a template. Choose a developer-owned ephemeral stage for local work, and reserve `staging` for deliberate shared validation.

   ```json
   {
     "targets": [
       {
         "stage": "dev-jane-20260715",
         "lifecycle": "ephemeral",
         "owner": "Your Team",
         "service": "bolt-monitor",
         "accountId": "123456789012",
         "region": "us-east-1",
         "credentialSource": "AWS profile your-team",
         "dashboardOrigin": "http://localhost:3000",
         "disposable": true,
         "expiresAt": "2026-08-01T00:00:00Z"
       }
     ]
   }
   ```

4. Start SST local development.

   ```bash
   export SST_TARGET_CONFIG="$HOME/.config/bolt-monitor/deployment-target.json"
   export SST_STAGE=dev-jane-20260715
   export AWS_PROFILE=bolt-monitor
   export SST_LIFECYCLE_ACTION=dev
   node scripts/sst-lifecycle.mjs
   ```

   The lifecycle wrapper refuses production stage names, mismatched account or region, incomplete target configuration, and missing confirmation. It is the only supported entry point for `sst deploy`, `sst remove`, `sst preview`, `sst dev`, and `sst install`.

5. Run the dashboard against the deployed API.

   ```bash
   export NEXT_PUBLIC_MONITOR_API_BASE_URL=<api-url>
   cd apps/dashboard
   pnpm run dev
   ```

6. Open local API documentation.

   ```bash
   cd openapi
   npm install
   npm run docs
   ```

   Swagger UI runs at `http://127.0.0.1:4173/swagger.html` and Redoc at `http://127.0.0.1:4173/redoc.html`.

For staging validation and administrative operations, see [`docs/auth-operations.md`](./docs/auth-operations.md), [`docs/stage-resource-lifecycle.md`](./docs/stage-resource-lifecycle.md), and [`docs/persistent-resource-operations.md`](./docs/persistent-resource-operations.md).

## Commands

| Intent | Command |
| --- | --- |
| Install infra dependencies | `cd infra && pnpm install --frozen-lockfile` |
| Install dashboard dependencies | `cd apps/dashboard && pnpm install --frozen-lockfile` |
| Test Go services and shared modules | `make test-go-all` |
| Lint Go code | `make lint-go` |
| Build Go handlers | `make build-go` |
| Lint and typecheck dashboard | `make lint-dashboard check-dashboard test-dashboard build-dashboard` |
| Typecheck infrastructure | `make check-infra test-infra` |
| Validate API contract drift | `make check-bruno check-api-contract` |
| Run the pre-cutover gate | `make check-pre-cutover-gate` |
| Start local SST development | `make dev-infra` with explicit target environment |
| Deploy infrastructure | `make deploy-infra` with explicit target environment |
| Remove infrastructure | `make remove-infra` with explicit target environment |
| Bootstrap or invite an administrator | `make bootstrap-admin EMAIL=<email> USER_POOL_ID=<id> AUTH_TABLE_NAME=<name>` |
| Rotate the dashboard auth key | `make rotate-auth-key` |
| Run the local staging smoke | `make smoke-staging` after a deliberate deploy |
| Run local OpenAPI documentation | `cd openapi && npm run docs` |

The lifecycle wrapper handles target validation, confirmation, destructive intent, and key rotation; do not invoke `sst deploy` or `sst remove` directly.

## Repository layout

| Path | Purpose |
| --- | --- |
| `infra/` | SST app and lifecycle policy |
| `infra/stacks/bootstrap.ts` | API Gateway, dashboard, DynamoDB, queues, schedule, runtime wiring |
| `services/api-health` | Go Lambda behind `GET /api/health` |
| `services/monitor-api` | Go Lambda for monitor, incident, audit, admin, and authentication flows |
| `services/check-runtime` | Go runtime for scheduler and worker modes |
| `services/escalation-runtime` | Go runtime that dispatches notification channels |
| `shared/` | Canonical Go domain modules wired by `go.work` |
| `apps/dashboard` | Next.js 15 App Router operator console |
| `openapi/` | Checked-in OpenAPI contract and local Swagger/Redoc tooling |
| `openspec/` | Spec-driven change workflow and merged capability specs |
| `.bruno/collections/` | Bruno API collection with domain-grouped requests and direct Cognito helpers |
| `docs/` | Lifecycle, authentication, and persistent-resource operations runbooks |
| `scripts/` | Repository-owned validators, lifecycle wrapper, and operator helpers |

## Documentation boundaries

- [`README.md`](./README.md) — product overview, architecture summary, status, setup, and command reference.
- [`CONSTITUTION.md`](./CONSTITUTION.md) — engineering principles and policy statements.
- [`AGENTS.md`](./AGENTS.md) — workflow, commands, conventions, response envelope, error handling, and implementation guidance.
- [`openspec/specs/`](./openspec/specs) — merged behavioral and technical specifications.
- [`DESIGN.md`](./DESIGN.md) — product and interface direction.
- [`docs/`](./docs) — operational runbooks for lifecycle, authentication, and recovery.

## License

Licensed under Apache License 2.0. See [`LICENSE`](./LICENSE).