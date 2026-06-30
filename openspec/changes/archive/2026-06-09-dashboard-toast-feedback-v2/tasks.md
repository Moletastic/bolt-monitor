## 1. Fix use-toast.ts Hook

- [x] 1.1 Remove `state` from useEffect dependency array in useToast hook
- [x] 1.2 Verify effect runs only once on mount

## 2. Fix actions.ts Server Actions

- [x] 2.1 Move redirect() outside try-catch in createServiceAction
- [x] 2.2 Move redirect() outside try-catch in updateServiceAction
- [x] 2.3 Move redirect() outside try-catch in createMonitorAction
- [x] 2.4 Move redirect() outside try-catch in updateMonitorAction
- [x] 2.5 Move redirect() outside try-catch in toggleMonitorAction
- [x] 2.6 Move redirect() outside try-catch in triggerManualRunAction
- [x] 2.7 Move redirect() outside try-catch in acknowledgeIncidentAction
- [x] 2.8 Move redirect() outside try-catch in resolveIncidentAction
- [x] 2.9 Move redirect() outside try-catch in updateSchedulerConfigAction

## 3. Fix toast-watcher.tsx Component

- [x] 3.1 Use window.location.search directly (not useSearchParams)
- [x] 3.2 Use empty dependency array so effect runs once on mount
- [x] 3.3 Handle all param types: created, updated, error, run

## 4. Add ServiceListStatusToast Component

- [ ] 4.1 Create component for status change detection
- [ ] 4.2 Use sessionStorage to track notified services
- [ ] 4.3 Show DOWN toast when service goes DOWN
- [ ] 4.4 Show UP toast when service recovers

## 5. Integrate Components

- [x] 5.1 Ensure Toaster is in root layout
- [x] 5.2 Ensure ToastWatcher is in (monitoring) layout
- [ ] 5.3 Add ServiceListStatusToast to services list page

## 6. Verify and Test

- [ ] 6.1 Create a service, verify toast appears
- [ ] 6.2 Create a monitor, verify toast appears
- [ ] 6.3 Wait for status change, verify toast appears
- [ ] 6.4 Trigger an error, verify error toast appears

## 7. Build and Deploy

- [ ] 7.1 Run `make lint-dashboard`
- [ ] 7.2 Run `make check-dashboard`
- [ ] 7.3 Run `make build-dashboard`
- [ ] 7.4 Deploy to staging