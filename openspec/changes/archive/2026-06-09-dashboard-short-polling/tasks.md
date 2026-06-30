## 1. Create PollingProvider Component

- [x] 1.1 Create `components/polling-provider.tsx` client component
- [x] 1.2 Implement `setInterval` with `router.refresh()` every 5 seconds
- [x] 1.3 Add `visibilitychange` listener to pause when tab hidden
- [x] 1.4 Add immediate refresh when tab becomes visible
- [x] 1.5 Handle cleanup on unmount

## 2. Integrate PollingProvider into Dashboard Layout

- [x] 2.1 Create `(monitoring)` route group with PollingProvider layout
- [x] 2.2 Ensure it only applies to monitoring routes (services, incidents)
- [x] 2.3 Verify polling doesn't affect non-monitoring pages (forms, admin)

## 3. Implement Toast Notifications

- [x] 3.1 Add Toaster to root layout
- [x] 3.2 Create ToastWatcher with useSearchParams + Suspense
- [x] 3.3 Create ServiceListStatusToast for status change detection
- [x] 3.4 Fix use-toast hook dependency array
- [x] 3.5 Fix redirect() in try-catch in actions.ts

## 4. Build and Deploy

- [x] 4.1 Run `make lint-dashboard` to check for linting issues
- [x] 4.2 Run `make check-dashboard` for TypeScript checks
- [x] 4.3 Build the dashboard: `make build-dashboard`
- [x] 4.4 Deploy to staging: `make deploy-infra`