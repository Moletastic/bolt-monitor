package main

import (
	"context"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type createMonitorStore interface {
	CreateMonitor(context.Context, monitorconfig.Monitor) (monitorconfig.Monitor, error)
}

type monitorLookup interface {
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
}

type updateMonitorStore interface {
	monitorLookup
	UpdateMonitor(context.Context, monitorconfig.Monitor) (monitorconfig.Monitor, error)
}

type deleteMonitorStore interface {
	DeleteMonitor(context.Context, string, string, string) (bool, error)
}

type setMonitorEnabledStore interface {
	SetMonitorEnabled(context.Context, string, string, string, bool) (monitorconfig.Monitor, bool, error)
}

type setMonitorMaintenanceStore interface {
	monitorLookup
	SetMonitorMaintenance(context.Context, string, string, string, bool) (resultstatus.MonitorStatus, bool, error)
}

type monitorStatusStore interface {
	GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error)
}

type listMonitorsStore interface {
	ListMonitors(context.Context, string, string) ([]monitorconfig.Monitor, error)
	monitorStatusStore
}

type monitorRunsStore interface {
	ListMonitorRunsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error)
}

type monitorAuditStore interface {
	ListMonitorAuditEventsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error)
}

type manualRunStore interface {
	monitorLookup
	RecordExecutionResult(context.Context, monitorconfig.Monitor, string, checkexecution.ExecutionResult) error
	ReserveManualIdempotency(context.Context, manualIdempotencyRecord) (manualIdempotencyRecord, error)
}

type createMonitorCommand struct {
	services            serviceLookup
	store               createMonitorStore
	validateDestination func(context.Context, monitorconfig.Monitor) error
	ids                 identifierGenerator
}

type updateMonitorCommand struct {
	store               updateMonitorStore
	validateDestination func(context.Context, monitorconfig.Monitor) error
}

type deleteMonitorCommand struct{ store deleteMonitorStore }
type setMonitorEnabledCommand struct{ store setMonitorEnabledStore }
type setMonitorMaintenanceCommand struct{ store setMonitorMaintenanceStore }
type listMonitorsQuery struct {
	services serviceLookup
	store    listMonitorsStore
}
type getMonitorQuery struct {
	monitors monitorLookup
	statuses monitorStatusStore
}
type monitorStatusQuery struct{ store monitorStatusStore }
type monitorRunsQuery struct{ store monitorRunsStore }
type monitorAuditQuery struct {
	monitors monitorLookup
	store    monitorAuditStore
}
type manualRunCommand struct {
	store    manualRunStore
	now      commandClock
	ids      identifierGenerator
	executor checkexecution.HTTPExecutor
}

type monitorOperations struct {
	create      createMonitorCommand
	update      updateMonitorCommand
	delete      deleteMonitorCommand
	enable      setMonitorEnabledCommand
	maintenance setMonitorMaintenanceCommand
	list        listMonitorsQuery
	get         getMonitorQuery
	status      monitorStatusQuery
	runs        monitorRunsQuery
	audit       monitorAuditQuery
	manualRun   manualRunCommand
}

func newMonitorOperations(services serviceLookup, create createMonitorStore, update updateMonitorStore, delete deleteMonitorStore, enable setMonitorEnabledStore, maintenance setMonitorMaintenanceStore, list listMonitorsStore, get monitorLookup, status monitorStatusStore, runs monitorRunsStore, audit monitorAuditStore, manual manualRunStore, now commandClock, ids identifierGenerator, executor checkexecution.HTTPExecutor, validateDestination func(context.Context, monitorconfig.Monitor) error) monitorOperations {
	return monitorOperations{
		create: createMonitorCommand{services: services, store: create, validateDestination: validateDestination, ids: ids},
		update: updateMonitorCommand{store: update, validateDestination: validateDestination},
		delete: deleteMonitorCommand{store: delete}, enable: setMonitorEnabledCommand{store: enable},
		maintenance: setMonitorMaintenanceCommand{store: maintenance}, list: listMonitorsQuery{services: services, store: list},
		get: getMonitorQuery{monitors: get, statuses: status}, status: monitorStatusQuery{store: status},
		runs: monitorRunsQuery{store: runs}, audit: monitorAuditQuery{monitors: get, store: audit},
		manualRun: manualRunCommand{store: manual, now: now, ids: ids, executor: executor},
	}
}

type createMonitorInput struct {
	TenantID  string
	ServiceID string
	Request   monitorconfig.CreateMonitorRequest
}

func (c createMonitorCommand) Execute(ctx context.Context, input createMonitorInput) (monitorconfig.Monitor, error) {
	if _, found, err := c.services.GetService(ctx, input.TenantID, input.ServiceID); err != nil {
		return monitorconfig.Monitor{}, err
	} else if !found {
		return monitorconfig.Monitor{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	targetURL := ""
	if input.Request.HTTP != nil {
		targetURL = input.Request.HTTP.Target
	}
	monitor, err := input.Request.ToMonitor(input.ServiceID, input.TenantID, c.ids.newMonitorID(string(input.Request.Type), targetURL, input.Request.Name))
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	if err := monitor.Validate(); err != nil {
		return monitorconfig.Monitor{}, err
	}
	if c.validateDestination != nil {
		if err := c.validateDestination(ctx, monitor); err != nil {
			return monitorconfig.Monitor{}, err
		}
	}
	return c.store.CreateMonitor(ctx, monitor)
}

type updateMonitorInput struct {
	TenantID, ServiceID, MonitorID string
	Request                        updateMonitorRequest
}

func (c updateMonitorCommand) Execute(ctx context.Context, input updateMonitorInput) (monitorconfig.Monitor, error) {
	current, found, err := c.store.GetMonitor(ctx, input.TenantID, input.ServiceID, input.MonitorID)
	if err != nil || !found {
		if err != nil {
			return monitorconfig.Monitor{}, err
		}
		return monitorconfig.Monitor{}, sharederrors.New(sharederrors.CodeMonitorNotFound, nil)
	}
	updated := applyMonitorPatch(current, input.Request)
	if err := updated.Validate(); err != nil {
		return monitorconfig.Monitor{}, err
	}
	if c.validateDestination != nil {
		if err := c.validateDestination(ctx, updated); err != nil {
			return monitorconfig.Monitor{}, err
		}
	}
	return c.store.UpdateMonitor(ctx, updated)
}

func applyMonitorPatch(monitor monitorconfig.Monitor, request updateMonitorRequest) monitorconfig.Monitor {
	updated := monitor
	if request.Name != nil {
		updated.Name = *request.Name
	}
	if request.IntervalSeconds != nil {
		updated.IntervalSeconds = *request.IntervalSeconds
	}
	if request.FailureThreshold != nil {
		updated.FailureThreshold = *request.FailureThreshold
	}
	if request.RecoveryThreshold != nil {
		updated.RecoveryThreshold = *request.RecoveryThreshold
	}
	if updated.FailureThreshold < 1 {
		updated.FailureThreshold = 1
	}
	if updated.RecoveryThreshold < 1 {
		updated.RecoveryThreshold = 1
	}
	if request.HTTP != nil {
		updated.HTTP = dynamodbrecord.CloneHTTPConfiguration(request.HTTP)
	}
	return updated
}

func (c deleteMonitorCommand) Execute(ctx context.Context, tenantID, serviceID, monitorID string) (bool, error) {
	return c.store.DeleteMonitor(ctx, tenantID, serviceID, monitorID)
}

func (c setMonitorEnabledCommand) Execute(ctx context.Context, tenantID, serviceID, monitorID string, enabled bool) (monitorconfig.Monitor, bool, error) {
	return c.store.SetMonitorEnabled(ctx, tenantID, serviceID, monitorID, enabled)
}

func (c setMonitorMaintenanceCommand) Execute(ctx context.Context, tenantID, serviceID, monitorID string, enabled bool) (resultstatus.MonitorStatus, bool, error) {
	if _, found, err := c.store.GetMonitor(ctx, tenantID, serviceID, monitorID); err != nil || !found {
		return resultstatus.MonitorStatus{}, found, err
	}
	return c.store.SetMonitorMaintenance(ctx, tenantID, serviceID, monitorID, enabled)
}

type monitorDetail struct {
	Monitor monitorconfig.Monitor
	Status  *resultstatus.MonitorStatus
}

func (q listMonitorsQuery) Execute(ctx context.Context, tenantID, serviceID string) ([]monitorDetail, bool, error) {
	if _, found, err := q.services.GetService(ctx, tenantID, serviceID); err != nil || !found {
		return nil, found, err
	}
	monitors, err := q.store.ListMonitors(ctx, tenantID, serviceID)
	if err != nil {
		return nil, false, err
	}
	details := make([]monitorDetail, 0, len(monitors))
	for _, monitor := range monitors {
		status, found, err := q.store.GetMonitorStatus(ctx, tenantID, serviceID, monitor.MonitorID)
		if err != nil {
			return nil, false, err
		}
		var value *resultstatus.MonitorStatus
		if found {
			value = &status
		}
		details = append(details, monitorDetail{Monitor: monitor, Status: value})
	}
	return details, true, nil
}

func (q getMonitorQuery) Execute(ctx context.Context, tenantID, serviceID, monitorID string) (monitorDetail, bool, error) {
	monitor, found, err := q.monitors.GetMonitor(ctx, tenantID, serviceID, monitorID)
	if err != nil || !found {
		return monitorDetail{}, found, err
	}
	status, foundStatus, err := q.statuses.GetMonitorStatus(ctx, tenantID, serviceID, monitorID)
	if err != nil {
		return monitorDetail{}, false, err
	}
	detail := monitorDetail{Monitor: monitor}
	if foundStatus {
		detail.Status = &status
	}
	return detail, true, nil
}

func (q monitorStatusQuery) Execute(ctx context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	return q.store.GetMonitorStatus(ctx, tenantID, serviceID, monitorID)
}

func (q monitorRunsQuery) Execute(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error) {
	return q.store.ListMonitorRunsPage(ctx, tenantID, serviceID, monitorID, limit, startKey)
}

func (q monitorAuditQuery) Execute(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[auditEventView], bool, error) {
	if _, found, err := q.monitors.GetMonitor(ctx, tenantID, serviceID, monitorID); err != nil || !found {
		return historyPage[auditEventView]{}, found, err
	}
	page, err := q.store.ListMonitorAuditEventsPage(ctx, tenantID, serviceID, monitorID, limit, startKey)
	return page, true, err
}

type manualRunResult struct {
	Record    manualRunRequestRecord
	Execution *checkexecution.ExecutionResult
	Existing  *manualIdempotencyRecord
}

func (c manualRunCommand) Execute(ctx context.Context, tenantID, serviceID, monitorID, idempotencyKey string) (manualRunResult, error) {
	now := c.now()
	fingerprint := manualRequestFingerprint(tenantID, serviceID, monitorID, idempotencyKey)
	reserved := newManualIdempotencyRecord(tenantID, serviceID, monitorID, idempotencyKey, fingerprint, c.ids.newRunID(now), now, manualIdempotencyRetentionSeconds)
	existing, err := c.store.ReserveManualIdempotency(ctx, reserved)
	if err != nil {
		return manualRunResult{}, err
	}
	if existing.Fingerprint != fingerprint {
		return manualRunResult{}, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "idempotency key reused with different request"})
	}
	if existing.RunID != reserved.RunID {
		return manualRunResult{Existing: &existing}, nil
	}
	monitor, found, err := c.store.GetMonitor(ctx, tenantID, serviceID, monitorID)
	if err != nil {
		return manualRunResult{}, err
	}
	if !found {
		return manualRunResult{}, sharederrors.New(sharederrors.CodeMonitorNotFound, nil)
	}
	if !monitor.Enabled {
		return manualRunResult{}, sharederrors.New(sharederrors.CodeMonitorDisabled, nil)
	}
	result := checkexecution.ExecuteHTTP(ctx, c.executor, checkexecution.ExecutionRequest{Monitor: monitor, RunID: reserved.RunID, Trigger: checkexecution.TriggerTypeManual})
	if err := c.store.RecordExecutionResult(ctx, monitor, reserved.RunID, result); err != nil {
		return manualRunResult{}, err
	}
	return manualRunResult{Record: manualRunRequestRecord{RunID: reserved.RunID, ServiceID: monitor.ServiceID, MonitorID: monitor.MonitorID, TenantID: monitor.TenantID, Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now.UTC().Format(time.RFC3339)}, Execution: &result}, nil
}
