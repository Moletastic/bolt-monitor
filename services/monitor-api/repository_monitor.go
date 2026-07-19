package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/domainvalues"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

// MonitorStore is the narrow interface used by monitor-lifecycle handlers and
// by the check-execution vertical slice.
type MonitorStore interface {
	CreateMonitor(context.Context, monitorconfig.Monitor) (monitorconfig.Monitor, error)
	ListMonitors(context.Context, string, string) ([]monitorconfig.Monitor, error)
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
	UpdateMonitor(context.Context, monitorconfig.Monitor) (monitorconfig.Monitor, error)
	DeleteMonitor(context.Context, string, string, string) (bool, error)
	SetMonitorEnabled(context.Context, string, string, string, bool) (monitorconfig.Monitor, bool, error)
	SetMonitorMaintenance(context.Context, string, string, string, bool) (resultstatus.MonitorStatus, bool, error)
	GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error)
	ListMonitorRuns(context.Context, string, string, string, int32) ([]resultstatus.CheckRun, error)
	ListMonitorRunsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error)
	GetServiceCardMetrics(context.Context, string, string) (serviceCardMetricsResponse, error)
	CreateManualRun(context.Context, monitorconfig.Monitor, time.Time) (manualRunRequestRecord, error)
	RecordExecutionResult(context.Context, monitorconfig.Monitor, string, checkexecution.ExecutionResult) error
	ListMonitorIncidents(context.Context, string, string, string) ([]dynamodbrecord.IncidentRecord, error)
	ListMonitorIncidentsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], error)
	ListServiceIncidents(context.Context, string, string, int32) ([]dynamodbrecord.IncidentRecord, error)
	GetMonitorByRef(context.Context, domainvalues.MonitorRef) (monitorconfig.Monitor, bool, error)
}

// executionStore is the minimal storage surface that the worker Lambda (and
// future monitor-execution adapters) needs. It is narrower than MonitorStore
// so the worker cannot accidentally touch handler-only operations.
type executionStore interface {
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
	RecordExecutionResult(context.Context, monitorconfig.Monitor, string, checkexecution.ExecutionResult) error
	ReserveManualIdempotency(context.Context, manualIdempotencyRecord) (manualIdempotencyRecord, error)
	LoadManualIdempotency(context.Context, string, string, string, string) (manualIdempotencyRecord, bool, error)
}

func (r *dynamoMonitorRepository) CreateMonitor(ctx context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Monitor{}, err
	}
	if _, found, err := r.GetMonitor(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID); err != nil {
		return monitorconfig.Monitor{}, err
	} else if found {
		return monitorconfig.Monitor{}, errMonitorAlreadyExists
	}
	now := r.now().UTC().Format(time.RFC3339)
	statusRecord := newDefaultMonitorStatusRecord(monitor, now)
	service, found, err := r.GetService(ctx, monitor.TenantID, monitor.ServiceID)
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	if !found {
		return monitorconfig.Monitor{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, []monitorconfig.Monitor{monitor}, map[string]resultstatus.MonitorStatus{monitorStatusMapKey(monitor): monitorStatusRecordToDomain(statusRecord)})
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, monitor.TenantID, "MONITOR_CREATED", monitor.ServiceID, monitor.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "monitor", "", monitor.Name)
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewMonitorItemRecord(monitor),
		dynamodbrecord.NewServiceMonitorRefItemRecord(monitor),
		statusRecord,
		serviceStatus,
		auditEvent,
		change,
	)
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Monitor{}, err
	}
	if err := r.replaceSearchIndex(ctx, monitor.TenantID, searchResourceMonitor, monitor.MonitorID, monitor.ServiceID, buildMonitorSearchRecords(monitor, service.Name)); err != nil {
		return monitorconfig.Monitor{}, err
	}
	return monitor, nil
}

func (r *dynamoMonitorRepository) ListMonitors(ctx context.Context, tenantID, serviceID string) ([]monitorconfig.Monitor, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.ServicePK(tenantID, serviceID), "MONITOR#")
	if err != nil {
		return nil, err
	}
	monitors := make([]monitorconfig.Monitor, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.MonitorItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityServiceMonitorRef {
			continue
		}
		monitors = append(monitors, record.ToMonitor())
	}
	sort.Slice(monitors, func(i, j int) bool { return monitors[i].MonitorID < monitors[j].MonitorID })
	return monitors, nil
}

func (r *dynamoMonitorRepository) GetMonitor(ctx context.Context, tenantID, serviceID, monitorID string) (monitorconfig.Monitor, bool, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.MonitorPK(tenantID, serviceID, monitorID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	if len(out.Item) == 0 {
		return monitorconfig.Monitor{}, false, nil
	}
	var record dynamodbrecord.MonitorItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	monitor := record.ToMonitor()
	if !strings.EqualFold(monitor.TenantID, tenantID) || !strings.EqualFold(monitor.ServiceID, serviceID) {
		return monitorconfig.Monitor{}, false, nil
	}
	return monitor, true, nil
}

// GetMonitorByRef is the MonitorRef-based vertical slice entry point. It
// performs the same exact-key read as GetMonitor but accepts the composite
// reference so callers that already have a MonitorRef do not need to split it.
func (r *dynamoMonitorRepository) GetMonitorByRef(ctx context.Context, ref domainvalues.MonitorRef) (monitorconfig.Monitor, bool, error) {
	return r.GetMonitor(ctx, string(ref.Tenant), string(ref.Service), string(ref.Monitor))
}

func (r *dynamoMonitorRepository) UpdateMonitor(ctx context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Monitor{}, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	service, found, err := r.GetService(ctx, monitor.TenantID, monitor.ServiceID)
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	if !found {
		return monitorconfig.Monitor{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, []monitorconfig.Monitor{monitor}, nil)
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, monitor.TenantID, "MONITOR_UPDATED", monitor.ServiceID, monitor.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "monitor", "", monitor.Name)
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewMonitorItemRecord(monitor),
		dynamodbrecord.NewServiceMonitorRefItemRecord(monitor),
		serviceStatus,
		auditEvent,
		change,
	)
	if err != nil {
		return monitorconfig.Monitor{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Monitor{}, err
	}
	if err := r.replaceSearchIndex(ctx, monitor.TenantID, searchResourceMonitor, monitor.MonitorID, monitor.ServiceID, buildMonitorSearchRecords(monitor, service.Name)); err != nil {
		return monitorconfig.Monitor{}, err
	}
	return monitor, nil
}

func (r *dynamoMonitorRepository) DeleteMonitor(ctx context.Context, tenantID, serviceID, monitorID string) (bool, error) {
	monitor, found, err := r.GetMonitor(ctx, tenantID, serviceID, monitorID)
	if err != nil || !found {
		return found, err
	}
	service, found, err := r.GetService(ctx, tenantID, serviceID)
	if err != nil || !found {
		return false, err
	}
	monitors, err := r.ListMonitors(ctx, tenantID, serviceID)
	if err != nil {
		return true, err
	}
	if service.LifecycleState == monitorconfig.ServiceLifecycleActive && len(monitors) == 1 {
		return true, errCannotDeleteLastMonitorFromActiveService
	}
	deleteSet := newDeleteKeySet()
	if err := r.collectMonitorDeleteKeys(ctx, tenantID, serviceID, monitorID, deleteSet); err != nil {
		return true, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, nil, nil, monitor.MonitorID)
	if err != nil {
		return true, err
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, tenantID, "MONITOR_DELETED", serviceID, monitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "monitor", monitor.Name, "")
	putItems, err := marshalPutItems(r.tableName, serviceStatus, auditEvent, change)
	if err != nil {
		return true, err
	}
	if err := r.deleteKeysAndPut(ctx, deleteSet.list(), putItems); err != nil {
		return true, err
	}
	if err := r.deleteSearchIndex(ctx, tenantID, searchResourceMonitor, monitorID, serviceID); err != nil {
		return true, err
	}
	return true, nil
}

func (r *dynamoMonitorRepository) SetMonitorEnabled(ctx context.Context, tenantID, serviceID, monitorID string, enabled bool) (monitorconfig.Monitor, bool, error) {
	monitor, found, err := r.GetMonitor(ctx, tenantID, serviceID, monitorID)
	if err != nil || !found {
		return monitor, found, err
	}
	monitor.Enabled = enabled
	now := r.now().UTC().Format(time.RFC3339)
	service, found, err := r.GetService(ctx, tenantID, serviceID)
	if err != nil || !found {
		return monitorconfig.Monitor{}, false, err
	}
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, []monitorconfig.Monitor{monitor}, nil)
	if err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	service.LifecycleState = monitorconfig.ServiceLifecycle(serviceStatus.LifecycleState)
	action := "MONITOR_DISABLED"
	if enabled {
		action = "MONITOR_ENABLED"
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, tenantID, action, serviceID, monitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "enabled", boolString(!enabled), boolString(enabled))
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewMonitorItemRecord(monitor),
		dynamodbrecord.NewServiceMonitorRefItemRecord(monitor),
		dynamodbrecord.NewServiceItemRecord(service),
		serviceStatus,
		auditEvent,
		change,
	)
	if err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	if err := r.replaceSearchIndex(ctx, monitor.TenantID, searchResourceMonitor, monitor.MonitorID, monitor.ServiceID, buildMonitorSearchRecords(monitor, service.Name)); err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	return monitor, true, nil
}

func (r *dynamoMonitorRepository) SetMonitorMaintenance(ctx context.Context, tenantID, serviceID, monitorID string, enabled bool) (resultstatus.MonitorStatus, bool, error) {
	if err := r.requireTableName(); err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	now := r.now().UTC()
	nowStr := now.Format(time.RFC3339)

	status, found, err := r.GetMonitorStatus(ctx, tenantID, serviceID, monitorID)
	if err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	if !found {
		status = resultstatus.MonitorStatus{
			ServiceID:     dynamodbschema.NormalizeField(serviceID),
			MonitorID:     dynamodbschema.NormalizeField(monitorID),
			TenantID:      dynamodbschema.NormalizeToken(tenantID),
			CurrentStatus: domainvalues.MonitorStateUnknown.Stored(),
			LastCheckedAt: now,
			LastOutcome:   checkexecution.Outcome(rollupUnknown),
		}
	}

	if enabled {
		status.CurrentStatus = string(resultstatus.MonitorStateMaintenance)
		status.ConsecutiveFailures = 0
		status.ConsecutiveSuccesses = 0
	} else {
		status.CurrentStatus = string(resultstatus.MonitorStateUp)
		status.ConsecutiveFailures = 0
		status.ConsecutiveSuccesses = 0
	}
	status.LastCheckedAt = now

	statusRecord := status.ToRecord()
	statusRecord.UpdatedAt = nowStr

	items, err := marshalPutItems(r.tableName, statusRecord)
	if err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	return status, true, nil
}

func (r *dynamoMonitorRepository) GetMonitorStatus(ctx context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	if err := r.requireTableName(); err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.MonitorPK(tenantID, serviceID, monitorID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "STATUS"},
		},
	})
	if err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	if len(out.Item) == 0 {
		return resultstatus.MonitorStatus{}, false, nil
	}
	var record resultstatus.MonitorStatusRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	lastCheckedAt, err := time.Parse(time.RFC3339, firstNonEmpty(record.LastCheckedAt, record.UpdatedAt))
	if err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	return resultstatus.MonitorStatus{
		ServiceID:       record.ServiceID,
		MonitorID:       record.MonitorID,
		TenantID:        record.TenantID,
		CurrentStatus:   record.CurrentStatus,
		LastCheckedAt:   lastCheckedAt,
		LastDurationMs:  record.LastDurationMs,
		LastError:       record.LastError,
		LastFailureCode: record.LastFailureCode,
		LastOutcome:     checkexecution.Outcome(strings.ToLower(firstNonEmpty(record.LastOutcome, rollupUnknown))),
	}, true, nil
}

func (r *dynamoMonitorRepository) ListMonitorRuns(ctx context.Context, tenantID, serviceID, monitorID string, limit int32) ([]resultstatus.CheckRun, error) {
	page, err := r.ListMonitorRunsPage(ctx, tenantID, serviceID, monitorID, limit, nil)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *dynamoMonitorRepository) ListMonitorRunsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error) {
	if err := r.requireTableName(); err != nil {
		return historyPage[resultstatus.CheckRun]{}, err
	}
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.MonitorPK(tenantID, serviceID, monitorID), "RUN#", sharedaws.PageOptions{
		Limit:   limit,
		Forward: false,
		Cursor:  startKey,
	})
	if err != nil {
		return historyPage[resultstatus.CheckRun]{}, err
	}
	runs := make([]resultstatus.CheckRun, 0, len(page.Items))
	for _, item := range page.Items {
		var record resultstatus.CheckRunRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return historyPage[resultstatus.CheckRun]{}, err
		}
		startedAt, err := time.Parse(time.RFC3339, record.StartedAt)
		if err != nil {
			return historyPage[resultstatus.CheckRun]{}, err
		}
		finishedAt, err := time.Parse(time.RFC3339, record.FinishedAt)
		if err != nil {
			return historyPage[resultstatus.CheckRun]{}, err
		}
		runs = append(runs, resultstatus.CheckRun{
			ServiceID:   record.ServiceID,
			MonitorID:   record.MonitorID,
			TenantID:    record.TenantID,
			RunID:       record.RunID,
			Type:        record.Type,
			Trigger:     checkexecution.TriggerType(strings.ToLower(record.Trigger)),
			StartedAt:   startedAt,
			FinishedAt:  finishedAt,
			DurationMs:  record.DurationMs,
			Outcome:     checkexecution.Outcome(strings.ToLower(record.Outcome)),
			StatusCode:  record.StatusCode,
			Error:       record.Error,
			FailureCode: record.FailureCode,
			TTL:         record.TTL,
		})
	}
	return historyPage[resultstatus.CheckRun]{Items: runs, NextKey: page.NextKey}, nil
}

func (r *dynamoMonitorRepository) GetServiceCardMetrics(ctx context.Context, tenantID, serviceID string) (serviceCardMetricsResponse, error) {
	monitors, err := r.ListMonitors(ctx, tenantID, serviceID)
	if err != nil {
		return serviceCardMetricsResponse{}, err
	}
	statuses := make(map[string]resultstatus.MonitorStatus, len(monitors))
	runsByMonitor := make(map[string][]resultstatus.CheckRun, len(monitors))
	for _, monitor := range monitors {
		status, found, err := r.GetMonitorStatus(ctx, tenantID, serviceID, monitor.MonitorID)
		if err != nil {
			return serviceCardMetricsResponse{}, err
		}
		if found {
			statuses[monitor.MonitorID] = status
		}
		runs, err := r.ListMonitorRuns(ctx, tenantID, serviceID, monitor.MonitorID, 20)
		if err != nil {
			return serviceCardMetricsResponse{}, err
		}
		runsByMonitor[monitor.MonitorID] = runs
	}
	return buildServiceCardMetrics(monitors, statuses, runsByMonitor), nil
}

func (r *dynamoMonitorRepository) CreateManualRun(ctx context.Context, monitor monitorconfig.Monitor, now time.Time) (manualRunRequestRecord, error) {
	if err := r.requireTableName(); err != nil {
		return manualRunRequestRecord{}, err
	}
	acceptedAt := now.UTC().Format(time.RFC3339)
	runID := newRunID(now)
	work := dynamodbrecord.NewExecutionWorkItemRecord(monitor.TenantID, monitor.ServiceID, monitor.MonitorID, runID, checkexecution.TriggerTypeManual, acceptedAt, checkexecution.ExecutionWorkPending, nil, nil, "")
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, monitor.TenantID, "MONITOR_RUN_REQUESTED", monitor.ServiceID, monitor.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "run", "", runID)
	records := []any{work, auditEvent, change}
	items, err := marshalPutItems(r.tableName, records...)
	if err != nil {
		return manualRunRequestRecord{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return manualRunRequestRecord{}, err
	}
	return manualRunRequestRecord{
		RunID:      runID,
		ServiceID:  monitor.ServiceID,
		MonitorID:  monitor.MonitorID,
		TenantID:   monitor.TenantID,
		Trigger:    checkexecution.TriggerTypeManual,
		AcceptedAt: acceptedAt,
	}, nil
}

func (r *dynamoMonitorRepository) RecordExecutionResult(ctx context.Context, monitor monitorconfig.Monitor, runID string, result checkexecution.ExecutionResult) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	now := r.now()
	checkRun := resultstatus.NewCheckRun(result, now)
	monitorStatus := resultstatus.NewMonitorStatus(result)
	checkRunRecord := checkRun.ToRecord()
	monitorStatusRecord := monitorStatus.ToRecord()
	records := []any{checkRunRecord, monitorStatusRecord}
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, monitor.TenantID, "MONITOR_CHECK_EXECUTED", monitor.ServiceID, monitor.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "run", "", runID)
	records = append(records, auditEvent, change)
	openIncidents, err := r.ListMonitorIncidents(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID)
	if err != nil {
		return err
	}
	var openIncident *dynamodbrecord.IncidentRecord
	for _, inc := range openIncidents {
		if inc.Status == incidentStatusOpen || inc.Status == incidentStatusAcknowledged {
			openIncident = &inc
			break
		}
	}
	if result.Outcome != checkexecution.OutcomeSuccess && openIncident == nil {
		incidentID := newIncidentID(now)
		incident := dynamodbrecord.IncidentRecord{
			IncidentID: incidentID,
			ServiceID:  monitor.ServiceID,
			MonitorID:  monitor.MonitorID,
			TenantID:   monitor.TenantID,
			Summary:    fmt.Sprintf("Monitor check failed: %s", result.Outcome),
			Status:     incidentStatusOpen,
			OpenedAt:   now.UTC().Format(time.RFC3339),
			UpdatedAt:  now.UTC().Format(time.RFC3339),
			Origin:     "monitoring",
		}
		records = append(records,
			dynamodbrecord.NewIncidentMonitorItemRecord(incident),
			dynamodbrecord.NewIncidentRefItemRecord(incident),
			dynamodbrecord.NewIncidentMetaItemRecord(incident),
		)
	} else if result.Outcome == checkexecution.OutcomeSuccess && openIncident != nil {
		openIncident.Status = incidentStatusResolved
		openIncident.ResolvedAt = now.UTC().Format(time.RFC3339)
		openIncident.UpdatedAt = openIncident.ResolvedAt
		if err := r.writeIncident(ctx, *openIncident, "INCIDENT_RESOLVED", now, openIncident.ResolvedAt); err != nil {
			return err
		}
	}
	service, found, err := r.GetService(ctx, monitor.TenantID, monitor.ServiceID)
	if err != nil {
		return err
	}
	if found {
		serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now.UTC().Format(time.RFC3339), nil, map[string]resultstatus.MonitorStatus{monitorStatusMapKey(monitor): monitorStatus})
		if err != nil {
			return err
		}
		records = append(records, serviceStatus)
	}
	items, err := marshalPutItems(r.tableName, records...)
	if err != nil {
		return err
	}
	return r.writeTransaction(ctx, items)
}

func (r *dynamoMonitorRepository) ListMonitorIncidents(ctx context.Context, tenantID, serviceID, monitorID string) ([]dynamodbrecord.IncidentRecord, error) {
	page, err := r.ListMonitorIncidentsPage(ctx, tenantID, serviceID, monitorID, 20, nil)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *dynamoMonitorRepository) ListMonitorIncidentsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], error) {
	if err := r.requireTableName(); err != nil {
		return historyPage[dynamodbrecord.IncidentRecord]{}, err
	}
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.MonitorPK(tenantID, serviceID, monitorID), "INCIDENT#", sharedaws.PageOptions{
		Limit:   limit,
		Forward: false,
		Cursor:  startKey,
	})
	if err != nil {
		return historyPage[dynamodbrecord.IncidentRecord]{}, err
	}
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(page.Items))
	for _, item := range page.Items {
		var record dynamodbrecord.IncidentItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return historyPage[dynamodbrecord.IncidentRecord]{}, err
		}
		if record.EntityType != dynamodbschema.EntityIncident {
			continue
		}
		incidents = append(incidents, record.ToIncident())
	}
	return historyPage[dynamodbrecord.IncidentRecord]{Items: incidents, NextKey: page.NextKey}, nil
}

func (r *dynamoMonitorRepository) ListServiceIncidents(ctx context.Context, tenantID, serviceID string, limit int32) ([]dynamodbrecord.IncidentRecord, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.TenantPK(tenantID), "INCIDENT#", sharedaws.PageOptions{
		Limit:   limit,
		Forward: false,
	})
	if err != nil {
		return nil, err
	}
	normalized := strings.ToLower(strings.TrimSpace(serviceID))
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(page.Items))
	for _, item := range page.Items {
		var record dynamodbrecord.IncidentItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbrecord.IncidentRefEntityType {
			continue
		}
		if !strings.EqualFold(record.ServiceID, normalized) {
			continue
		}
		incidents = append(incidents, record.ToIncident())
	}
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].OpenedAt > incidents[j].OpenedAt })
	return incidents, nil
}

// collectMonitorDeleteKeys is shared between monitor deletion and service
// deletion. It scans the monitor partition and selects the items that share
// the monitor's lifecycle.
func (r *dynamoMonitorRepository) collectMonitorDeleteKeys(ctx context.Context, tenantID, serviceID, monitorID string, deleteSet *deleteKeySet) error {
	monitorPK := dynamodbschema.MonitorPK(tenantID, serviceID, monitorID)
	monitorItems, err := r.queryPartition(ctx, monitorPK, "")
	if err != nil {
		return err
	}
	for _, item := range monitorItems {
		if shouldDeleteWithMonitorConfig(item) {
			deleteSet.addItem(item)
		}
	}
	deleteSet.add(dynamodbschema.ServicePK(tenantID, serviceID), dynamodbschema.ServiceMonitorRefSK(monitorID))
	return nil
}

func shouldDeleteWithMonitorConfig(item map[string]sharedaws.AttributeValue) bool {
	sk, ok := item["SK"].(*sharedaws.AttributeValueMemberS)
	if !ok {
		return false
	}
	return sk.Value == "META" || sk.Value == "STATUS" || sk.Value == "ALERT_STATE"
}

// buildServiceStatusRecord composes the canonical service-status item from
// the current monitor inventory and any status overrides supplied by the
// caller.
func (r *dynamoMonitorRepository) buildServiceStatusRecord(ctx context.Context, service monitorconfig.Service, updatedAt string, monitorsOverride []monitorconfig.Monitor, statusOverride map[string]resultstatus.MonitorStatus, excludedMonitorIDs ...string) (dynamodbrecord.ServiceStatusRecord, error) {
	monitors, err := r.ListMonitors(ctx, service.TenantID, service.ServiceID)
	if err != nil {
		return dynamodbrecord.ServiceStatusRecord{}, err
	}
	excluded := map[string]struct{}{}
	for _, monitorID := range excludedMonitorIDs {
		excluded[strings.ToLower(strings.TrimSpace(monitorID))] = struct{}{}
	}
	byID := map[string]monitorconfig.Monitor{}
	for _, monitor := range monitors {
		if _, skip := excluded[strings.ToLower(monitor.MonitorID)]; skip {
			continue
		}
		byID[monitorStatusMapKey(monitor)] = monitor
	}
	for _, monitor := range monitorsOverride {
		if _, skip := excluded[strings.ToLower(monitor.MonitorID)]; skip {
			continue
		}
		byID[monitorStatusMapKey(monitor)] = monitor
	}
	merged := make([]monitorconfig.Monitor, 0, len(byID))
	for _, monitor := range byID {
		merged = append(merged, monitor)
	}
	statuses := map[string]resultstatus.MonitorStatus{}
	for key, status := range statusOverride {
		statuses[key] = status
	}
	for _, monitor := range merged {
		key := monitorStatusMapKey(monitor)
		if _, ok := statuses[key]; ok {
			continue
		}
		status, found, err := r.GetMonitorStatus(ctx, service.TenantID, monitor.ServiceID, monitor.MonitorID)
		if err != nil {
			return dynamodbrecord.ServiceStatusRecord{}, err
		}
		if found {
			statuses[key] = status
		}
	}
	enabledCount := countEnabledMonitors(merged)
	if service.LifecycleState != monitorconfig.ServiceLifecycleArchived {
		if enabledCount > 0 {
			service.LifecycleState = monitorconfig.ServiceLifecycleActive
		} else {
			service.LifecycleState = monitorconfig.ServiceLifecycleDraft
		}
	}
	rollup := deriveServiceRollup(service.LifecycleState, buildMonitorSummaries(merged, statuses))
	return dynamodbrecord.ServiceStatusRecord{
		PK:                  dynamodbschema.ServicePK(service.TenantID, service.ServiceID),
		SK:                  "STATUS",
		EntityType:          dynamodbschema.EntityServiceStatus,
		TenantID:            strings.ToUpper(service.TenantID),
		ServiceID:           strings.ToLower(service.ServiceID),
		LifecycleState:      string(service.LifecycleState),
		RollupStatus:        rollup,
		MonitorCount:        len(merged),
		EnabledMonitorCount: enabledCount,
		UpdatedAt:           updatedAt,
		GSI2PK:              dynamodbschema.TenantPK(service.TenantID),
		GSI2SK:              dynamodbschema.ServiceStatusItem(service.TenantID, service.ServiceID, rollup, updatedAt).GSI2SK,
	}, nil
}

// newDefaultMonitorStatusRecord materializes the placeholder status record
// written when a monitor is first created.
func newDefaultMonitorStatusRecord(monitor monitorconfig.Monitor, now string) resultstatus.MonitorStatusRecord {
	status := resultstatus.MonitorStatus{
		ServiceID:      monitor.ServiceID,
		MonitorID:      monitor.MonitorID,
		TenantID:       monitor.TenantID,
		CurrentStatus:  domainvalues.MonitorStateUnknown.Stored(),
		LastCheckedAt:  mustParseTime(now),
		LastDurationMs: 0,
		LastError:      "",
		LastOutcome:    checkexecution.Outcome(rollupUnknown),
	}
	return status.ToRecord()
}

// monitorStatusRecordToDomain converts a persisted status record back into
// the domain representation with parsed timestamps.
func monitorStatusRecordToDomain(record resultstatus.MonitorStatusRecord) resultstatus.MonitorStatus {
	lastCheckedAt, _ := time.Parse(time.RFC3339, firstNonEmpty(record.LastCheckedAt, record.UpdatedAt))
	return resultstatus.MonitorStatus{
		ServiceID:      record.ServiceID,
		MonitorID:      record.MonitorID,
		TenantID:       record.TenantID,
		CurrentStatus:  record.CurrentStatus,
		LastCheckedAt:  lastCheckedAt,
		LastDurationMs: record.LastDurationMs,
		LastError:      record.LastError,
		LastOutcome:    checkexecution.Outcome(strings.ToLower(firstNonEmpty(record.LastOutcome, rollupUnknown))),
	}
}

// monitorStatusLookup is a small adapter used by ServiceStore to keep its
// dependency surface tight.
