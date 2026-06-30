## Context

The current system has several data consistency and display issues:

1. **Scheduler skips all enabled monitors regardless of service lifecycle** - A monitor on a "draft" service still executes because the monitor's `enabled` flag is true, even though the service is not ready.

2. **Dashboard shows "unknown" status** - The monitor status endpoint returns data, but the dashboard UI isn't displaying the status correctly. The `currentStatus` field should show "up", "down", or other states, not "unknown".

3. **Duration shows N/A** - The dashboard isn't displaying the `durationMs` field from the API response.

4. **Service lifecycle is manually editable** - The frontend shows a dropdown to change lifecycle state, but this should be state-managed:
   - Draft → Active: Automatic when first monitor is enabled
   - Active → Archived: Manual (user action)

## Goals / Non-Goals

**Goals:**
- Scheduler skips monitors on draft services
- Dashboard correctly displays monitor status (up/down/unknown)
- Dashboard correctly displays duration in milliseconds
- Service lifecycle auto-transitions Draft → Active when first monitor is enabled
- Service lifecycle only manually transitions Active → Archived

**Non-Goals:**
- Changing the underlying status model (CheckRun, MonitorStatus records)
- Modifying how status is calculated (outcome derivation)
- Adding new status types

## Decisions

### Decision 1: Scheduler Checks Service LifecycleState

**Choice**: When scheduler evaluates a monitor, it first checks the parent service's `lifecycleState`. If "draft", skip the monitor.

**Rationale**: Draft services are not yet production-ready. Monitors on draft services should not execute automatically, even if the monitor's `enabled` flag is true.

**Implementation**:
```go
func (h runtimeHandler) runScheduler(ctx context.Context) (runtimeSummary, error) {
    // ... existing config and monitor reading ...

    for _, monitor := range monitors {
        // NEW: Check service lifecycle
        service, found, err := h.repo.GetService(ctx, monitor.TenantID, monitor.ServiceID)
        if err != nil {
            return summary, err
        }
        if found && service.LifecycleState == monitorconfig.ServiceLifecycleDraft {
            continue // skip monitors on draft services
        }

        // ... rest of existing logic ...
    }
}
```

### Decision 2: Dashboard Status Display Fix

**Choice**: Investigate and fix the dashboard component that displays monitor status. Ensure it reads `status.currentStatus` from the API response.

**Rationale**: The API returns correct data (as seen in runs API), but the dashboard listing or status component isn't displaying it correctly.

**Likely Issues**:
- Dashboard may be reading wrong field path
- Status display component may have wrong conditional logic
- Default value of "unknown" being shown when actual status exists

### Decision 3: Dashboard Duration Display Fix

**Choice**: Fix the dashboard component that displays duration. Ensure it reads `status.lastDurationMs` and formats it as milliseconds.

**Rationale**: Duration should be displayed when available, not as N/A.

### Decision 4: Service Lifecycle Auto-Transition

**Choice**: When a monitor is enabled on a draft service, automatically transition the service to "active".

**Implementation**:
```go
// In monitor-api handler when enabling a monitor
func (h monitorHandler) enableMonitor(ctx context.Context, serviceID, monitorID string) {
    // ... existing enable logic ...

    // NEW: If service is draft, transition to active
    service, found, err := h.repo.GetService(ctx, h.tenantID, serviceID)
    if err != nil {
        return serverError(err)
    }
    if found && service.LifecycleState == monitorconfig.ServiceLifecycleDraft {
        // Transition service to active
        if err := h.repo.UpdateServiceLifecycle(ctx, h.tenantID, serviceID, monitorconfig.ServiceLifecycleActive); err != nil {
            return serverError(err)
        }
    }
}
```

### Decision 5: Active → Archived is Manual Only

**Choice**: Service lifecycle transitions from Active to Archived happen only via explicit user action (archive button), not automatically.

**Rationale**: Archiving is a business decision that should be intentional, not triggered by system state.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Dashboard fix may be complex if status field path is wrong | Investigate API response first, then trace through UI components |
| Service lifecycle transition during enable may cause race condition | Use DynamoDB transactions to ensure atomicity |
| Existing draft services with running monitors will stop | This is expected behavior - draft services shouldn't have running monitors |

## Migration Plan

1. **Deploy scheduler fix** - monitors on draft services stop executing
2. **Verify** - check that draft services don't have running monitors
3. **Fix dashboard status display** - ensure status shows correctly
4. **Fix dashboard duration display** - ensure duration shows correctly
5. **Deploy lifecycle auto-transition** - when enabling first monitor on draft service, service becomes active
6. **Remove manual lifecycle dropdown** - if currently exposed in UI

## Open Questions

1. **Should we also check service lifecycle when running manual executions?** Currently manual runs via API execute regardless of service state. Should they also be blocked for draft services?

2. **What happens if a service is manually set to Archived?** Should all monitors on that service be automatically disabled?

3. **Should we show a warning in the dashboard for draft services?** Help users understand why their monitors aren't executing.
