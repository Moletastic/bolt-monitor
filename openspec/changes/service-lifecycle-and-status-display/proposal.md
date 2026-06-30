## Why

The dashboard is showing incorrect states for services and monitors:
1. Draft services are showing monitors running in background (scheduler should skip draft services)
2. Monitor status shows "unknown" instead of actual status (up/down)
3. Duration shows "N/A" instead of actual duration
4. Service lifecycle state is manually editable in frontend, but should be state-managed by backend

## What Changes

- **Modified**: Scheduler to skip monitors on draft services (monitors only execute on active services)
- **Modified**: Dashboard to correctly display monitor status and duration from API response
- **Modified**: Service lifecycle state to auto-transition Draft → Active when first monitor is enabled
- **Modified**: Service lifecycle state to only allow Active → Archived manually (not automatic)

## Capabilities

### New Capabilities
- `service-lifecycle-auto-transition`: Service transitions from Draft to Active automatically when first monitor is enabled

### Modified Capabilities
- `check-runtime-scheduler-mode`: Scheduler SHALL skip monitors where the parent service lifecycleState is "draft"
- `monitor-status-read-api`: Ensure status and duration fields are properly exposed for dashboard display
- `monitor-configuration`: Clarify that `intervalSeconds` handling should respect service lifecycle

## Impact

- **Code**: `services/check-runtime/runtime.go` - add service lifecycle check in scheduler
- **Code**: `apps/dashboard` - fix status and duration display in UI components
- **Code**: `services/monitor-api` - add automatic lifecycle transition when monitor enabled
