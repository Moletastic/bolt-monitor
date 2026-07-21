# Bolt Monitor

[![CI](https://github.com/Moletastic/bolt-monitor/actions/workflows/ci.yml/badge.svg)](https://github.com/Moletastic/bolt-monitor/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev)
[![Node](https://img.shields.io/badge/Node-22-339933?logo=node.js)](https://nodejs.org)
[![pnpm](https://img.shields.io/badge/pnpm-10-F69220?logo=pnpm)](https://pnpm.io)
[![Next.js](https://img.shields.io/badge/Next.js-15-000?logo=next.js)](https://nextjs.org)
[![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-FF9900?logo=amazon-aws)](https://aws.amazon.com/lambda/)
[![SST](https://img.shields.io/badge/SST-4.14-E73358?logo=react)](https://sst.dev)

Bolt Monitor is a self-deployed uptime and incident platform built on AWS serverless. It schedules HTTP health checks, stores execution and status history, manages incident lifecycle, and exposes an operator dashboard without requiring a permanently running monitoring server.

It is targeted at small teams that want to own their monitoring stack on AWS and prefer pay-for-execution infrastructure over a long-lived control plane.

## 🛰 Why Bolt Monitor

- AWS-native deployment: API Gateway, Lambda, DynamoDB, EventBridge, SQS, Cognito, and CloudFront wired through a single SST stack.
- Service-scoped data model: services own monitors, runs, incidents, audit, and escalation policies.
- Scheduled HTTP checks with execution and status history, plus manual run and audit trails.
- Incident acknowledgement, resolution, and escalation through persisted notification channels.
- Operator dashboard backed by server-side API fetches; no client-side data shadowing.
- Single-tenant today with Cognito-authenticated dashboard and versioned API.

## 🧰 Capabilities

- Service-scoped monitor creation, editing, enable and disable, maintenance mode, manual run, archive and reactivate.
- Recurring HTTP checks with configurable intervals, expected status, and optional body assertions.
- Run history and latest status, monitor audit trail, and service audit trail.
- Incident list, detail, acknowledge, resolve, and escalation-state inspection.
- Escalation policies and notification channels (webhook, email, SMS, PagerDuty, Telegram).
- Scheduler configuration endpoints for cadence and dispatch policy.
- Operator dashboard with sidebar navigation, monitor and incident workflows, audit timeline, and admin scheduler view.
- Public `GET /api/health` and invite-only Cognito authentication for `/api/v1/**`.
- Checked-in OpenAPI contract with deterministic SST, Bruno, OpenAPI, and handler-route drift gates.
- Bruno API collection with explicit per-route auth classification and locally stored credentials.

Detailed route inventories live in `openapi/openapi.yaml` and `.bruno/collections/`.

## ⏱ How it works

1. The operator creates a service and a monitor in the dashboard or directly through the API.
2. A managed EventBridge schedule invokes the scheduler Lambda once per minute.
3. The scheduler reads configuration, filters enabled and non-maintenance monitors, and enqueues due executions on the execution SQS queue.
4. The execution SQS queue invokes the worker Lambda.
5. The worker performs the HTTP check against the monitor target and persists the result on `AppTable`.
6. Status and incident state are updated from the result.
7. State transitions enqueue notification events on the notification SQS queue.
8. The notification queue invokes the escalation runtime, which resolves the service escalation policy and dispatches the configured notification senders.
9. The dashboard renders status, run history, incident detail, and audit entries through server-side fetches against the API.

Manual runs reuse the same worker code path and persistence layer as recurring checks.

## 🏗 Architecture

```
Browser
  |
  v
Next.js dashboard (apps/dashboard, deployed via SST/OpenNext)
  |
  | server-side fetches and server actions
  v
API Gateway V2 (infra/stacks/bootstrap.ts)
  |
  +-- GET /api/health              -> services/api-health
  +-- /api/v1/**                   -> services/monitor-api  (Cognito JWT authorizer)
                                        |
                                        v
                                  DynamoDB (AppTable, AuthTable)

Managed EventBridge schedule (rate 1 minute)
        |
        v
scheduler Lambda (services/check-runtime, RUNTIME_MODE=scheduler)
        |
        v
Execution SQS queue
        |
        v
worker Lambda (services/check-runtime, RUNTIME_MODE=worker)
        |                          \
        |                           +--> AppTable (DynamoDB)
        |
        +--> Notification SQS queue (on status transitions)
                       |
                       v
           escalation runtime (services/escalation-runtime)
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
| DynamoDB single-table operational model | Serverless persistence for monitors, runs, incidents, and audit without a long-lived database. |
| Separate `AuthTable` for authentication state | Authoritative memberships and session material, independent of operational data. |
| Managed EventBridge schedule | One per-minute invocation drives the scheduler Lambda; no self-hosted scheduler. |
| SQS between scheduler, worker, and escalation | Bounded retries with DLQs, no in-process queue. |
| Server-side dashboard fetches | Authoritative state, no client-side data shadowing. |
| OpenAPI and Bruno for contracts | Drift gates compare SST, Bruno, OpenAPI, OpenSpec, and the monitor handler inventory. |
| Cognito user pool with AuthTable authorization | Invite-only operators, scoped JWT access tokens, application-owned authorization decisions. |
| SST for deployment | Single source of truth for AWS wiring; lifecycle guard rejects production or ambiguous targets. |

## 📊 Project status

Functional today:

- Public `GET /api/health` returns the standard success envelope.
- Service-scoped monitor lifecycle, runs, status, audit, and incident flows.
- Invitation-only Cognito authentication with optional software-token MFA for the dashboard and the versioned API.
- Recurring and manual HTTP execution through the worker Lambda, with run and status history persisted.
- Escalation dispatch through webhook, email, SMS, PagerDuty, and Telegram channels.
- Local deterministic release gates for Go, dashboard, infrastructure, Bruno, and OpenAPI contract drift.

Known limitations:

- Single built-in tenant ID (`DEFAULT`) and single-region execution.
- No user-management or RBAC beyond activating an `ADMIN` membership through the bootstrap tool.
- No multi-region probe execution, public status page, or PromQL/log ingestion.
- AWS credentials remain required to deploy or run a staging environment; the dashboard itself does not call AWS directly.

Production-readiness caveats:

- No promised SLA, RPO, or RTO. Recovery drills and capacity guardrails are pending follow-on work.
- Backups, point-in-time recovery, and retention policies apply to persistent `AppTable`, `AuthTable`, and the operator Cognito user pool.
- The API and dashboard accept Cognito access tokens and reject protected-route requests without one. SSRF and outbound-boundary hardening is in progress.
- Operational runbooks cover lifecycle, authentication, and recovery; they are written for the implemented surface, not for the proposed scope.

## 🚀 Quick start

Bolt Monitor targets Node.js 22, pnpm 10, Go 1.26+, and SST 4.14.

1. Install AWS CLI tooling and configure one named profile for the target AWS account.

   ```bash
   aws configure --profile bolt-monitor
   ```

2. Install repository dependencies and synchronize the Go workspace.

   ```bash
   make setup
   ```

3. Run the local release gates.

   ```bash
   make test-go-all
   make check-infra
   make lint-dashboard
   make check-bruno check-api-contract
   ```

4. Create the staging target file from the committed template. The target file is ignored by Git and is the only place AWS identity, region, owner, lifecycle class, and dashboard origin live.

   ```bash
   cp infra/targets/example.target.json infra/targets/staging.target.json
   ```

   Edit `infra/targets/staging.target.json` and replace the example AWS account, region, profile, dashboard origin, and owner with the values for the staging installation. `dashboardOrigin` must be a canonical HTTPS origin without a path. For a generated CloudFront URL, deploy once with a placeholder dashboard origin or an existing HTTPS origin, copy the deployed `dashboardUrl` from `infra/.sst/outputs.json` into the target file, and redeploy so dashboard authentication uses the correct canonical origin.

5. Deploy the staging infrastructure.

   ```bash
   make deploy-infra
   ```

   The orchestrator binds `AWS_PROFILE` and `AWS_REGION` from the target file, verifies the effective STS account and region, runs SST, and verifies outputs, persistent protections, and public health.

6. Invite the first administrator. The orchestrator resolves deployed Cognito and `AuthTable` identifiers from SST output automatically.

   ```bash
   make invite-admin EMAIL=operator@example.com
   ```

   The operator opens the invitation email, sets a permanent password through the dashboard activation flow, and optionally enrolls a software-token MFA secret.

7. Run the dashboard against the deployed API.

   ```bash
   export NEXT_PUBLIC_MONITOR_API_BASE_URL=<api-url>
   cd apps/dashboard
   pnpm run dev
   ```

8. Open local API documentation.

   ```bash
   cd openapi
   npm install
   npm run docs
   ```

   Swagger UI runs at `http://127.0.0.1:4173/swagger.html` and Redoc at `http://127.0.0.1:4173/redoc.html`.

For staging validation and operational mechanics, see [`docs/auth-operations.md`](./docs/auth-operations.md), [`docs/stage-resource-lifecycle.md`](./docs/stage-resource-lifecycle.md), and [`docs/persistent-resource-operations.md`](./docs/persistent-resource-operations.md).

## ⚡ Commands

| Intent | Command |
| --- | --- |
| Install repository dependencies and Go workspace | `make setup` |
| Test Go services and shared modules | `make test-go-all` |
| Lint, typecheck, test, and build dashboard | `make lint-dashboard check-dashboard test-dashboard build-dashboard` |
| Typecheck infrastructure | `make check-infra test-infra` |
| Validate API contract drift | `make check-bruno check-api-contract` |
| Run the pre-cutover gate | `make check-pre-cutover-gate` |
| Inspect a configured target | `make infra-status` |
| Start local SST development | `make dev-infra` with explicit `TARGET=<name>` |
| Deploy infrastructure | `make deploy-infra` with optional `TARGET=<name>` |
| Remove infrastructure | `make remove-infra` (ephemeral) or `TARGET=<name> make remove-infra DESTROY=yes` (persistent) |
| Invite an administrator | `make invite-admin EMAIL=<email>` |
| Rotate the dashboard auth key | `make rotate-auth-key` |
| Run local OpenAPI documentation | `cd openapi && npm run docs` |

The Make-driven orchestrator under `infra/scripts/ops.mjs` is the only supported entrypoint for `sst dev`, `sst deploy`, and `sst remove`. Do not invoke `sst deploy` or `sst remove` directly.

## 📁 Repository layout

| Path | Purpose |
| --- | --- |
| `infra/` | SST app, lifecycle policy, and target configuration |
| `infra/targets/` | Per-target `*.target.json` files (one ignored file per target) |
| `infra/scripts/` | Internal infrastructure orchestrator and ephemeral cleanup helpers |
| `infra/stacks/bootstrap.ts` | API Gateway, dashboard, DynamoDB, queues, schedule, runtime wiring |
| `services/api-health` | Go Lambda behind `GET /api/health` |
| `services/monitor-api` | Go Lambda for monitor, incident, audit, admin, and authentication flows |
| `services/check-runtime` | Go runtime operating as scheduler and worker based on `RUNTIME_MODE` |
| `services/escalation-runtime` | Go runtime that resolves escalation policies and dispatches notification channels |
| `shared/` | Canonical Go domain modules wired by `go.work` |
| `apps/dashboard` | Next.js 15 App Router operator console |
| `openapi/` | Checked-in OpenAPI contract and local Swagger/Redoc tooling |
| `openspec/` | Spec-driven change workflow and merged capability specs |
| `.bruno/collections/` | Bruno API collection with domain-grouped requests and direct Cognito helpers |
| `docs/` | Lifecycle, authentication, and persistent-resource operations runbooks |
| `scripts/` | Repository-owned validators and operator helpers |

## 📚 Documentation boundaries

- [`README.md`](./README.md) — product overview, architecture summary, status, setup, and command reference.
- [`CONSTITUTION.md`](./CONSTITUTION.md) — engineering principles and policy statements.
- [`AGENTS.md`](./AGENTS.md) — workflow, commands, conventions, response envelope, error handling, and implementation guidance.
- [`openspec/specs/`](./openspec/specs) — merged behavioral and technical specifications.
- [`DESIGN.md`](./DESIGN.md) — product and interface direction.
- [`docs/`](./docs) — operational runbooks for lifecycle, authentication, and recovery.

## 📄 License

Licensed under Apache License 2.0. See [`LICENSE`](./LICENSE).