# AGENTS.md

## Workflow
- This repo is spec-driven. Do not implement behavior that is not covered by an active OpenSpec change.
- OpenSpec source of truth lives under `openspec/`: active work goes in `openspec/changes/<name>/`, merged capabilities live in `openspec/specs/`, and archived changes live in `openspec/changes/archive/`.
- If a spec exists but the implementation plan is unclear, inspect `openspec status --change <name>` before editing code.

## Repo Shape
- `infra/` is the deployable SST app. This repo does not use CDK commands.
- `infra/stacks/bootstrap.ts` is the real wiring point for AWS resources and API routes.
- `services/api-health` is the simple Go Lambda behind `GET /api/health`.
- `services/monitor-api` is the Go Lambda for monitor CRUD, status, runs, incidents, and admin config.
- `shared/` holds the canonical Go domain modules; `go.work` wires them together for local multi-module development.
- `apps/dashboard` is a Next 15 App Router app that talks to the monitor API through server-side fetches and server actions.

## Verified Commands

All commands run from repo root via `make <target>`.

### Go
- `make test-go-all` — test all services and shared packages
- `make lint-go` — lint all Go code
- `make build-go` — build and zip Go handlers (api-health, check-runtime, monitor-api)
- Facade pattern: domain and service code depend on `bolt-monitor/shared/aws` interfaces such as `aws.DynamoDBAPI`, not `*dynamodb.Client`. Lambda entrypoints construct SDK clients through `aws.NewDynamoDBAPI` / `aws.NewSQSAPI`; repositories accept the facade.
- Rules pattern: compose validation with `shared/rules` and return typed errors with `details.field`:

```go
var builder rules.Builder[monitorconfig.Monitor]
builder.Add(rules.Field("intervalSeconds", func(m monitorconfig.Monitor) error {
	if !monitorconfig.IsAllowedIntervalSeconds(m.IntervalSeconds) {
		return errors.New(errors.CodeValidationFailed, map[string]any{"reason": "unsupported interval"})
	}
	return nil
}))
return builder.Build()(monitor)
```

### Dashboard
- `make lint-dashboard` — ESLint
- `make check-dashboard` — TypeScript type check
- `make test-dashboard` — Vitest unit and guard tests
- `make format-dashboard` — format with Prettier
- `make build-dashboard` — Next.js production build

### Infra
- `make check-infra` — TypeScript type check
- `make format-infra` — format with Prettier
- `make deploy-infra` — SST deploy to staging (uses `AWS_PROFILE=bolt-monitor`)

### JavaScript Package Manager
- `infra/` and `apps/dashboard` use `pnpm` with `pnpm-lock.yaml` committed and `packageManager` pinned per root.
- Install with `pnpm install --frozen-lockfile` from inside the relevant package root.
- Install-script execution is deny-by-default; only entries in the matching `.npmrc` `onlyBuiltDependencies` allowlist may run scripts. Update the allowlist and the install-script trust doc together when adding a new exception.
- `openapi/` remains on `npm` and is explicitly out of scope for this policy until a follow-on change moves it.

### Extras
- `make lint-all` — go + dashboard + infra lint
- `make bootstrap` — `go work sync`
- `make check-bruno` — validate Bruno coverage and request conventions against SST routes

### Bruno API Collection
- Bruno requests live under `.bruno/collections/` and cover every method/path route declared in `infra/stacks/bootstrap.ts`.
- Organize requests by API domain: `health`, `search`, `channels`, `policies`, `services`, `monitors`, `incidents`, and `admin`.
- Name requests `Verb Resource`; use exact route variables such as `serviceId`, `monitorId`, `incidentId`, `channelId`, and `policyId`.
- Every request has exactly one `domain:<domain>` tag and one `operation:<operation>` tag.
- Every request docs block includes `Purpose:`, `Setup:`, and `Expected result:`.
- When adding or changing an API route, update Bruno coverage and run `make check-bruno`.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) format.

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Example: `feat(monitor-api): add monitor audit endpoint`

## Response envelope

Every Lambda returns the same JSON shape:

```json
{
  "status": "success" | "error",
  "data": <T> | null,
  "reason": { "code": string, "details": Record<string, unknown> } | null,
  "message": string | null,
  "pagination": { "page": number, "size": number, "total": number, "items": unknown[] } | null
}
```

Optional fields are omitted from the JSON when not applicable; they are never emitted as `null`. Handlers construct the envelope through helpers, never via raw structs.

- Go (services + shared): `shared/api/response`
  - `response.Ok(data, message...)` — success envelope with `data`
  - `response.Err[T](code, details)` — failure envelope with `reason`
  - `response.OkPaginated(data, page, size, total)` — adds `pagination`
  - `Envelope[T].MarshalJSON` emits the shape above, omitting nil optional fields
- TypeScript (dashboard): `apps/dashboard/lib/api-response.ts`
  - `ApiResponse<T>` mirrors the Go envelope
  - `ok`, `err`, `okPaginated` factories plus `isSuccess` / `isError` type guards
  - `lib/errors.ts` owns `ApiErrorCode`, `ApiError`, `fromEnvelope`, and `messageFor`
  - `lib/api.ts` unwraps the envelope before returning data and surfaces `reason.code` via `ApiError`

## TypeScript error handling

- Use `Result<T, E>` from `apps/dashboard/lib/result.ts` for fallible dashboard helpers. Branch with `isOk` / `isErr`; reserve `unwrap` for tests or after an exhaustive `match`.
- Catching thrown values is restricted to `apps/dashboard/lib/io/**`. Add I/O adapters there and expose `Result<T, ApiError>` to the rest of `lib/**`.
- Server actions that still navigate with `redirect()` should call `runServerAction` at the API boundary, branch on `isErr`, and surface errors through `messageFor(error)` before redirecting.
- `ApiErrorCode` must exactly mirror `shared/errors/code.go`; `apps/dashboard/lib/errors.test.ts` fails if the Go and TypeScript registries drift.
- TypeScript `any` is a lint error. Use `unknown` and narrow with type guards, DOM `instanceof` checks, or schema validation before member access.
- Dashboard time handling uses `date-fns`; do not call or construct native `Date` outside `apps/dashboard/lib/clock.ts` and test/setup files. Use `parseISO`, `formatISO`, `compareDesc`, `differenceInMilliseconds`, and the `now()` clock wrapper instead.

Adding a new endpoint? Return one of the three constructors above and the parser on the dashboard side does the rest. Handlers route every error through `errors.Respond` (from `bolt-monitor/shared/errors`); the `Code` constants and the registry there are the single source of truth for `reason.code` values and their HTTP status mapping.

## Gotchas
- `infra/package.json` pins local SST dev to `sst dev --mode=mono`; keep that unless you intentionally change the TTY workaround.
- `infra/sst.config.ts` hard-codes AWS profile `bolt-monitor`. AWS commands from SST will use that profile unless the config is changed.
- Use explicit SST stage `staging` for normal local dev and deploy workflows to avoid recreating stray stage-specific resources.
- `services/monitor-api` requires `TABLE_NAME`; the SST stack injects it for the Lambda.
- The monitor API currently uses a single built-in tenant ID, `DEFAULT`.
- Route steps reference channels by `channelId`; configure channels under Integrations -> Channels.
- Monitor execution location is not operator-configurable; do not add dashboard region/probe-location pickers or hard-coded location identifiers.
- `apps/dashboard` requires `NEXT_PUBLIC_MONITOR_API_BASE_URL`; without it, server-rendered pages fail with an `ApiError`.
- Dashboard bootstrap assumptions (retained here for developer reference, removed from operator UI):
  - Single tenant context.
  - Service-first API is the source of truth.
  - Escalation policies do not own service-scoped business hours; the policy create action is responsible only for its own payload.
  - Destructive deletes go through `<ConfirmDialog>` (Radix AlertDialog); do not introduce `window.confirm` in operator flows.
- Router API usage (`dashboard-router-convention` spec): prefer `<Link>` from `next/link` for navigation and server actions or `<form action={...}>` for state changes that follow navigation. Reserve `useRouter`, `usePathname`, `router.push`, and `router.refresh()` for the polling provider's interval-driven revalidation; do not introduce new imperative router calls elsewhere.
