## 1. SST Dashboard Hosting

- [x] 1.1 Add a standalone `sst.aws.Nextjs` site for `apps/dashboard` in `infra/stacks/bootstrap.ts`
- [x] 1.2 Inject `NEXT_PUBLIC_MONITOR_API_BASE_URL` into the dashboard site from the deployed API URL
- [x] 1.3 Add `dashboardUrl` to SST stack outputs

## 2. Validation

- [x] 2.1 Run `cd infra && pnpm run check` to validate the SST stack shape
- [x] 2.2 Deploy the updated SST stack and capture the generated dashboard URL output
- [x] 2.3 Verify the deployed dashboard loads successfully and can reach the monitor API

## 3. Documentation

- [x] 3.1 Update root `README.md` deployment guidance to include SST-hosted dashboard deployment and generated URL output
- [x] 3.2 Update `apps/dashboard/README.md` to note SST deployment path and stack-managed `NEXT_PUBLIC_MONITOR_API_BASE_URL`
