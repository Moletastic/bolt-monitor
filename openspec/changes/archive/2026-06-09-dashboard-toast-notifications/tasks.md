## 1. Add Toaster to Root Layout

- [x] 1.1 Import `Toaster` from `@/components/ui/toaster` in `app/layout.tsx`
- [x] 1.2 Add `<Toaster />` after children in body

## 2. Create ToastWatcher Component

- [x] 2.1 Create `components/toast-watcher.tsx` client component
- [x] 2.2 Use `useSearchParams` to read `?created`, `?updated`, `?error` params
- [x] 2.3 Call `toast()` with appropriate variant based on params
- [x] 2.4 Handle cleanup on unmount

## 3. Integrate ToastWatcher into Monitoring Layout

- [x] 3.1 Import and add `ToastWatcher` to `(monitoring)/layout.tsx`
- [x] 3.2 Ensure it wraps all monitoring pages

## 4. Add Status Change Toasts to Services List

- [x] 4.1 Create `components/service-list-status-toast.tsx` for status change detection
- [x] 4.2 Track previous status vs current status using sessionStorage
- [x] 4.3 Show destructive toast when service goes DOWN
- [x] 4.4 Show success toast when service recovers to UP

## 5. Add Toasts to Service Detail Page

- [x] 5.1 Add `ServiceListStatusToast` to service detail page for status tracking
- [ ] 5.2 Show toast on status transitions (via ServiceListStatusToast)

## 6. Add Toasts to Monitor Detail Page

- [x] 6.1 Show toast on `?created` success (via ToastWatcher)
- [x] 6.2 Show toast on `?updated` success (via ToastWatcher)
- [ ] 6.3 Add status change detection for monitor status

## 7. Add Toasts to Incidents Pages

- [x] 7.1 ToastWatcher handles success/error params on incident pages

## 8. Verify and Test

- [ ] 8.1 Create a service, verify toast appears
- [ ] 8.2 Create a monitor, verify toast appears
- [ ] 8.3 Wait for status change, verify toast appears
- [ ] 8.4 Trigger an error, verify error toast appears

## 9. Build and Deploy

- [x] 9.1 Run `make lint-dashboard`
- [x] 9.2 Run `make check-dashboard`
- [x] 9.3 Run `make build-dashboard`
- [x] 9.4 Deploy to staging