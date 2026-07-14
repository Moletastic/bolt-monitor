# Bolt Monitor API Collection

Collection documentation stored next to Bruno request files. Requests mirror every route wired in `infra/stacks/bootstrap.ts`.

## Domains

- `health`: API availability
- `search`: global resource search
- `channels`: notification channel CRUD and delivery test
- `policies`: escalation policy CRUD
- `services`: service CRUD, service incidents, and service audit
- `monitors`: service-scoped monitor CRUD, status, runs, manual run, incidents, audit, enable, and disable
- `incidents`: incident reads and state transitions
- `admin`: scheduler configuration

## Conventions

- Request names use `Verb Resource`.
- Route variables match API names: `serviceId`, `monitorId`, `runId`, `incidentId`, `channelId`, and `policyId`.
- Every request has exactly one `domain:<domain>` tag and one `operation:<operation>` tag.
- Every request docs block contains `Purpose`, `Setup`, and `Expected result`.
- Run `make check-bruno` after route or request changes.

## Variables

- `apiUrl`: base API URL, configured in ignored `environments/development.local.yml`
- `serviceId`: service identifier
- `monitorId`: monitor identifier under `serviceId`
- `runId`: manual run identifier
- `incidentId`: incident identifier
- `channelId`: notification channel identifier
- `policyId`: escalation policy identifier

## Example flow

1. Create or list a service.
2. Create a monitor under `serviceId`.
3. Read, run, update, disable, or enable the monitor.
4. Inspect monitor and service incidents/audit.
5. Test channel and escalation policy operations separately.

Create local environment file from `environments/development.example.yml`; never commit deployed URLs or credentials.
