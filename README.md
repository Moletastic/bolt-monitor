# ⚡ bolt-monitor

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev)
[![CI](https://github.com/Moletastic/bolt-monitor/actions/workflows/ci.yml/badge.svg)](https://github.com/Moletastic/bolt-monitor/actions/workflows/ci.yml)
[![Next.js](https://img.shields.io/badge/Next.js-15-000?logo=next.js)](https://nextjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-FF9900?logo=amazon-aws)](https://aws.amazon.com/lambda/)
[![SST](https://img.shields.io/badge/SST-Framework-E73358?logo=react)](https://sst.dev)

Work-in-progress uptime monitoring built around AWS serverless primitives. Lightweight health checks, monitor CRUD, status history, incident views, and an operator dashboard without running a long-lived control plane.

## ⚡ Why bolt-monitor

- AWS-native deployment model with small always-on footprint
- Go backends with clear service boundaries instead of one large monolith
- Operator-first dashboard for monitor management and runtime inspection
- OpenAPI contract checked into repo for local docs and API review
- Honest early-stage scope: useful for local development and staged validation, not yet a broad replacement for mature monitoring suites

## 🚀 Current Capabilities

- `GET /api/health` Lambda health endpoint
- Monitor CRUD and enable/disable flows in `services/monitor-api`
- Monitor status, recent runs, monitor audit trail, and monitor incident history endpoints
- Manual run trigger endpoint for monitors
- Incident list, incident detail, acknowledge, and resolve endpoints
- Scheduler config read/update endpoints
- Recurring runtime services in `services/check-runtime` for scheduler and worker modes
- Next.js dashboard for monitor workflows and module landing pages
- Local OpenAPI docs via Swagger UI and Redoc

## 📊 Project Status

bolt-monitor works as real software, not scaffold. Also still rough.

Known limitations today:

- Single built-in tenant ID: `DEFAULT`
- Single execution environment; regional probe selection is intentionally out of scope for now
- Authentication, RBAC, and multi-user access are intentionally outside the current single-operator scope
- SST mutations require an explicit validated deployment target; no stage, lifecycle class, profile, account, or region is inferred
- No production hardening claims around security, scaling policy, backup policy, or multi-region execution

## 🏗️ Architecture

```text
Browser
  |
  v
Next.js dashboard (`apps/dashboard`)
  |
  | server-side fetches / server actions
  v
API Gateway V2 (`infra/stacks/bootstrap.ts`)
  |
  +--> `GET /api/health`
  |      -> Go Lambda (`services/api-health`)
  |
  +--> monitor + incident + admin routes
         -> Go Lambda (`services/monitor-api`)
                    |
                    v
                 DynamoDB

EventBridge Cron
  |
  +--> scheduler mode -> Go Lambda (`services/check-runtime`)
  +--> worker mode    -> Go Lambda (`services/check-runtime`)
                    |
                    v
                 DynamoDB
```

## 📋 Prerequisites

- Node.js 22+
- pnpm 10.x (pinned per package root via `packageManager`; `infra/.npmrc` and `apps/dashboard/.npmrc`)
- Go 1.26+
- AWS credentials configured for target account
- AWS credentials for the declared target, supplied through an explicit profile or workload identity

The default JavaScript workflow for `infra/` and `apps/dashboard` is `pnpm`
with `pnpm-lock.yaml` committed. `npm install` against those roots is no
longer supported and will produce a drift warning.

## ⚡ Quick Start

### 1. Install infra dependencies

```bash
cd infra
pnpm install --frozen-lockfile
```

### 2. Install dashboard dependencies

```bash
cd apps/dashboard
pnpm install --frozen-lockfile
```

### 3. Validate main workspaces

```bash
make check-infra
```

```bash
make test-go-all
```

```bash
make lint-dashboard
```

### 4. Configure and start SST local development

Copy `infra/deployment-target.example.json` outside source control, replace its
example account, region, owner, and expiry values, then declare it explicitly.
`staging` is persistent only when approved in this file. Prefer a unique,
developer-owned ephemeral target for local work.

```bash
export SST_TARGET_CONFIG="$HOME/.config/bolt-monitor/deployment-target.json"
export SST_STAGE=dev-jane-20260715
export AWS_PROFILE=bolt-monitor
export SST_LIFECYCLE_ACTION=dev
node scripts/sst-lifecycle.mjs
```

The preflight prints non-secret target identity, then prints a target-bound
`SST_TARGET_CONFIRMATION` value if confirmation is missing. Set that exact
value and rerun. It rejects mismatched account/region or incomplete target
configuration before SST mutates AWS. Persistent remove and protection changes
also require `SST_DESTRUCTIVE_CONFIRMATION`.

SST prints resource outputs, including API URL.

### 5. Verify health endpoint

```bash
curl <api-url>/api/health
```

Expected response. Health stays public; versioned API route authentication lands in a follow-on change.

```json
{"status":"success","data":{"status":"ok"}}
```

### 6. Point dashboard at monitor API for local development

Dashboard needs `NEXT_PUBLIC_MONITOR_API_BASE_URL` set before server-rendered pages work.

```bash
export NEXT_PUBLIC_MONITOR_API_BASE_URL=<api-url>
cd apps/dashboard
pnpm run dev
```

For deployed hosting, SST now injects `NEXT_PUBLIC_MONITOR_API_BASE_URL` into the dashboard runtime automatically.

### 7. Open local API docs

```bash
cd openapi
npm install
npm run docs
```

Docs server runs at `http://127.0.0.1:4173/` with:

- Swagger UI: `http://127.0.0.1:4173/swagger.html`
- Redoc: `http://127.0.0.1:4173/redoc.html`

## 🔧 Environment Variables

| Variable | Used by | Required | Notes |
| --- | --- | --- | --- |
| `NEXT_PUBLIC_MONITOR_API_BASE_URL` | `apps/dashboard` | Yes for dashboard runtime | Base URL for server-side API fetches. Missing value throws `ApiError`. |
| `TABLE_NAME` | `services/monitor-api`, `services/check-runtime` | Yes in deployed/local Lambda runtime | Injected by SST stack when handlers are wired. |
| `RUNTIME_MODE` | `services/check-runtime` | Yes for runtime Lambda behavior | Set by SST cron jobs to `scheduler` or `worker`. |

## ⚡ Common Commands

| Intent | Command |
| --- | --- |
| Install infra deps | `cd infra && pnpm install --frozen-lockfile` |
| Typecheck infra | `make check-infra` |
| Start local infra | `make dev-infra` with explicit target environment |
| Review persistent changes | `docs/persistent-resource-operations.md` runbook; SST `4.14.1` has no safe preview command |
| Deploy infra | `make deploy-infra` with explicit target environment |
| Remove ephemeral infra | `make remove-infra` with explicit target environment |
| Test Go services/shared modules | `make test-go-all` |
| Install dashboard deps | `cd apps/dashboard && pnpm install --frozen-lockfile` |
| Run dashboard lint | `make lint-dashboard` |
| Run API route governance | `make check-bruno && make check-api-contract` |
| Run dashboard production build | `make build-dashboard` |
| Start dashboard dev server | `cd apps/dashboard && pnpm run dev` |
| Install OpenAPI docs deps | `cd openapi && npm install` |
| Run local API docs | `cd openapi && npm run docs` |

## 📁 Repository Layout

| Path | Purpose |
| --- | --- |
| `infra/` | SST app that defines API Gateway, dashboard, DynamoDB, queues, and runtime jobs |
| `infra/stacks/bootstrap.ts` | Main wiring for routes, dashboard, table, queues, scheduler, workers, and escalation runtime |
| `services/api-health` | Small Go Lambda behind `GET /api/health` |
| `services/monitor-api` | Go Lambda for monitor CRUD, status, runs, incidents, and admin config |
| `services/check-runtime` | Go runtime worker/scheduler service for recurring execution |
| `shared/` | Canonical Go domain modules used across services |
| `apps/dashboard` | Next 15 App Router operator dashboard |
| `openapi/` | Checked-in OpenAPI contract and local docs tooling |
| `openspec/` | Spec-driven change workflow and implementation artifacts |
| `DESIGN.md` | Product and design direction reference |

## 📝 Development Notes

- Repo is spec-driven. Active implementation work should map to an OpenSpec change.
- `go.work` wires local Go modules together across `services/` and `shared/`.
- `services/monitor-api` and `services/check-runtime` both depend on DynamoDB table injection from SST.
- Dashboard uses server-side fetches; if `NEXT_PUBLIC_MONITOR_API_BASE_URL` is unset, page rendering fails fast.
- Declare lifecycle target explicitly for every local, preview, deploy, and remove workflow. Do not infer `staging`; use approved persistent staging only for deliberate shared validation.
- SST deploys the dashboard as a standalone Next.js site and outputs `dashboardUrl` with the generated CloudFront hostname.
- Monitor execution location is not operator-configurable in the dashboard.
- `AWS_PROFILE` is optional but never silently defaulted. The target config records only a non-secret credential-source label.
- Route changes update SST wiring, `services/monitor-api/routes.go`, Bruno, OpenAPI, and affected merged OpenSpec specs. Run `make check-bruno check-api-contract`.
- Before protected-route refactoring, also run dashboard production build and portable stage/profile checks.
- Archive completed OpenSpec changes only after implementation validation. Archival is maintenance, not runtime or release-gate verification.

## 🔐 JavaScript Dependency Install Policy

`infra/` and `apps/dashboard` use `pnpm` with `pnpm-lock.yaml` committed and
`packageManager` pinned. Install-script execution is deny-by-default; only
packages listed in the matching `.npmrc` `onlyBuiltDependencies` allowlist
may run build or install scripts. New dependencies that need install scripts
must be reviewed and added to that allowlist with a justification, not
inherited silently from package-manager defaults. The full trust policy and
current approved exceptions live in
[`openspec/specs/javascript-dependency-install-security`](./openspec/specs/javascript-dependency-install-security/spec.md).

`openapi/` is intentionally out of scope for this policy and continues to use
`npm`. Treat migrating it as a separate follow-on change rather than expanding
the current policy silently.

## 🚀 Dashboard Deploy

Deploying the SST stack now also deploys `apps/dashboard`.

```bash
cd infra
pnpm exec sst deploy --stage staging
```

Expected outputs include:

- `apiUrl`
- `dashboardUrl`

`dashboardUrl` is the generated CloudFront hostname for the deployed operator console. No custom DNS is required for the first deployment.

The dashboard is currently designed for private or controlled deployments. Until authentication and RBAC land in a follow-on change, place deployed dashboard URLs behind environment-level access controls such as a private network, VPN, SSO proxy, or restricted ingress.

## 🗺️ Roadmap

Near-term focus areas, not promises:

- Expand dashboard from module landing pages into richer operational summary views
- Add authentication and authorization model
- Improve runtime execution visibility, failure analysis, and operator ergonomics
- Tighten deployment and production-readiness story around config, docs, and safety rails

## 📄 License

Licensed under Apache License 2.0. See [`LICENSE`](./LICENSE).
