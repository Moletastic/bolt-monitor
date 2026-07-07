package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

var errMissingTableName = sharederrors.New(sharederrors.CodeInternal, nil)
var errIncidentNotActionable = sharederrors.New(sharederrors.CodeIncidentNotActionable, nil)
var errServiceAlreadyExists = sharederrors.New(sharederrors.CodeServiceAlreadyExists, nil)
var errMonitorAlreadyExists = sharederrors.New(sharederrors.CodeMonitorAlreadyExists, nil)

// var errServiceActiveRequiresMonitors = sharederrors.New(sharederrors.CodeServiceActive, nil)
var errCannotDeleteActiveService = sharederrors.New(sharederrors.CodeServiceActive, nil)
var errCannotDeleteLastMonitorFromActiveService = sharederrors.New(sharederrors.CodeLastMonitor, nil)

const (
	entityIncidentRef     = "IncidentRef"
	entitySchedulerConfig = "SchedulerConfig"

	rollupDraft    = "draft"
	rollupArchived = "archived"
	rollupPaused   = "paused"
	rollupUnknown  = "unknown"
	rollupUp       = "up"
	rollupDown     = "down"
	rollupDegraded = "degraded"

	incidentStatusOpen         = "open"
	incidentStatusAcknowledged = "acknowledged"
	incidentStatusResolved     = "resolved"
)

type dynamoAPI = sharedaws.DynamoDBAPI

type dynamoMonitorRepository struct {
	client    dynamoAPI
	tableName string
	now       func() time.Time
}

func newDynamoMonitorRepository(client dynamoAPI, tableName string) *dynamoMonitorRepository {
	return &dynamoMonitorRepository{client: client, tableName: tableName, now: time.Now}
}

func (r *dynamoMonitorRepository) CreateService(ctx context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Service{}, err
	}
	if _, found, err := r.GetService(ctx, service.TenantID, service.ServiceID); err != nil {
		return monitorconfig.Service{}, err
	} else if found {
		return monitorconfig.Service{}, errServiceAlreadyExists
	}
	now := r.now().UTC().Format(time.RFC3339)
	service.CreatedAt = now
	service.UpdatedAt = now
	service.MonitorCount = 0
	service.EnabledCount = 0
	service.RollupStatus = deriveServiceRollup(service.LifecycleState, nil)
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewServiceItemRecord(service),
		dynamodbrecord.NewServiceRefItemRecord(service),
		dynamodbrecord.NewServiceStatusItemRecord(service, now),
	)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Service{}, err
	}
	return service, nil
}

func (r *dynamoMonitorRepository) ListServices(ctx context.Context, tenantID string) ([]monitorconfig.Service, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "SERVICE#")
	if err != nil {
		return nil, err
	}
	services := make([]monitorconfig.Service, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.ServiceItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityServiceRef {
			continue
		}
		service := record.ToService()
		status, found, err := r.GetServiceStatus(ctx, tenantID, service.ServiceID)
		if err != nil {
			return nil, err
		}
		if found {
			service.MonitorCount = status.MonitorCount
			service.EnabledCount = status.EnabledMonitorCount
			service.RollupStatus = status.RollupStatus
			if service.LifecycleState != monitorconfig.ServiceLifecycleArchived {
				if status.EnabledMonitorCount > 0 {
					service.LifecycleState = monitorconfig.ServiceLifecycleActive
				} else {
					service.LifecycleState = monitorconfig.ServiceLifecycleDraft
				}
			}
		}
		summaries, err := r.serviceMonitorSummaries(ctx, tenantID, service.ServiceID)
		if err != nil {
			return nil, err
		}
		service.MonitorSummaries = summaries
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].ServiceID < services[j].ServiceID })
	return services, nil
}

func (r *dynamoMonitorRepository) serviceMonitorSummaries(ctx context.Context, tenantID, serviceID string) ([]monitorconfig.MonitorSummary, error) {
	monitors, err := r.ListMonitors(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	statuses := make(map[string]resultstatus.MonitorStatus, len(monitors))
	for _, monitor := range monitors {
		status, found, err := r.GetMonitorStatus(ctx, tenantID, serviceID, monitor.MonitorID)
		if err != nil {
			return nil, err
		}
		if found {
			statuses[monitorStatusKey(monitor.ServiceID, monitor.MonitorID)] = status
		}
	}
	return buildMonitorSummaries(monitors, statuses), nil
}

func (r *dynamoMonitorRepository) GetService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, bool, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Service{}, false, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.ServicePK(tenantID, serviceID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil {
		return monitorconfig.Service{}, false, err
	}
	if len(out.Item) == 0 {
		return monitorconfig.Service{}, false, nil
	}
	var record dynamodbrecord.ServiceItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return monitorconfig.Service{}, false, err
	}
	service := record.ToService()
	if !strings.EqualFold(service.TenantID, tenantID) {
		return monitorconfig.Service{}, false, nil
	}
	status, found, err := r.GetServiceStatus(ctx, tenantID, serviceID)
	if err != nil {
		return monitorconfig.Service{}, false, err
	}
	if found {
		service.MonitorCount = status.MonitorCount
		service.EnabledCount = status.EnabledMonitorCount
		service.RollupStatus = status.RollupStatus
		if service.LifecycleState != monitorconfig.ServiceLifecycleArchived {
			if status.EnabledMonitorCount > 0 {
				service.LifecycleState = monitorconfig.ServiceLifecycleActive
			} else {
				service.LifecycleState = monitorconfig.ServiceLifecycleDraft
			}
		}
	}
	return service, true, nil
}

func (r *dynamoMonitorRepository) UpdateService(ctx context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Service{}, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	service.UpdatedAt = now
	statusRecord, err := r.buildServiceStatusRecord(ctx, service, now, nil, nil)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	service.MonitorCount = statusRecord.MonitorCount
	service.EnabledCount = statusRecord.EnabledMonitorCount
	service.RollupStatus = statusRecord.RollupStatus
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewServiceItemRecord(service),
		dynamodbrecord.NewServiceRefItemRecord(service),
		statusRecord,
	)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Service{}, err
	}
	return service, nil
}

func (r *dynamoMonitorRepository) DeleteService(ctx context.Context, tenantID, serviceID string) (bool, error) {
	service, found, err := r.GetService(ctx, tenantID, serviceID)
	if err != nil || !found {
		return found, err
	}
	if service.LifecycleState == monitorconfig.ServiceLifecycleActive {
		return true, errCannotDeleteActiveService
	}
	monitors, err := r.ListMonitors(ctx, tenantID, serviceID)
	if err != nil {
		return true, err
	}
	deleteSet := newDeleteKeySet()
	servicePartitionItems, err := r.queryPartition(ctx, dynamodbschema.ServicePK(tenantID, serviceID), "")
	if err != nil {
		return true, err
	}
	deleteSet.addItems(servicePartitionItems)
	deleteSet.add(dynamodbschema.TenantPK(tenantID), dynamodbschema.ServiceRefSK(serviceID))
	for _, monitor := range monitors {
		if err := r.collectMonitorDeleteKeys(ctx, tenantID, serviceID, monitor.MonitorID, deleteSet); err != nil {
			return true, err
		}
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, tenantID, "SERVICE_DELETED", serviceID, "")
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "service", service.Name, "")
	putItems, err := marshalPutItems(r.tableName, auditEvent, change)
	if err != nil {
		return true, err
	}
	if err := r.deleteKeysAndPut(ctx, deleteSet.list(), putItems); err != nil {
		return true, err
	}
	return true, nil
}

func (r *dynamoMonitorRepository) GetServiceStatus(ctx context.Context, tenantID, serviceID string) (dynamodbrecord.ServiceStatusRecord, bool, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.ServiceStatusRecord{}, false, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.ServicePK(tenantID, serviceID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "STATUS"},
		},
	})
	if err != nil {
		return dynamodbrecord.ServiceStatusRecord{}, false, err
	}
	if len(out.Item) == 0 {
		return dynamodbrecord.ServiceStatusRecord{}, false, nil
	}
	var record dynamodbrecord.ServiceStatusRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return dynamodbrecord.ServiceStatusRecord{}, false, err
	}
	return record, true, nil
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
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, []monitorconfig.Monitor{monitor}, map[string]resultstatus.MonitorStatus{monitorStatusKey(monitor.ServiceID, monitor.MonitorID): monitorStatusRecordToDomain(statusRecord)})
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
			ServiceID:     strings.ToLower(serviceID),
			MonitorID:     strings.ToLower(monitorID),
			TenantID:      strings.ToUpper(tenantID),
			CurrentStatus: strings.ToUpper(rollupUnknown),
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

func (r *dynamoMonitorRepository) ArchiveService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Service{}, err
	}
	service, found, err := r.GetService(ctx, tenantID, serviceID)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if !found {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	now := r.now().UTC().Format(time.RFC3339)
	service.LifecycleState = monitorconfig.ServiceLifecycleArchived
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, nil, nil)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, tenantID, "SERVICE_ARCHIVED", serviceID, "")
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewServiceItemRecord(service),
		serviceStatus,
		auditEvent,
	)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Service{}, err
	}
	return service, nil
}

func (r *dynamoMonitorRepository) ReactivateService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Service{}, err
	}
	service, found, err := r.GetService(ctx, tenantID, serviceID)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if !found {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	if service.LifecycleState != monitorconfig.ServiceLifecycleArchived {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotArchived, nil)
	}
	now := r.now().UTC().Format(time.RFC3339)
	if service.EnabledCount > 0 {
		service.LifecycleState = monitorconfig.ServiceLifecycleActive
	} else {
		service.LifecycleState = monitorconfig.ServiceLifecycleDraft
	}
	serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now, nil, nil)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	auditID := newAuditID(r.now())
	auditEvent := dynamodbrecord.NewAuditEventRecord(r.now(), auditID, tenantID, "SERVICE_REACTIVATED", serviceID, "")
	items, err := marshalPutItems(r.tableName,
		dynamodbrecord.NewServiceItemRecord(service),
		serviceStatus,
		auditEvent,
	)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return monitorconfig.Service{}, err
	}
	return service, nil
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
		ServiceID:           record.ServiceID,
		MonitorID:           record.MonitorID,
		TenantID:            record.TenantID,
		CurrentStatus:       record.CurrentStatus,
		LastCheckedAt:       lastCheckedAt,
		LastDurationMs:      record.LastDurationMs,
		LastProbeLocationID: record.LastProbeLocationID,
		LastError:           record.LastError,
		LastOutcome:         checkexecution.Outcome(strings.ToLower(firstNonEmpty(record.LastOutcome, rollupUnknown))),
	}, true, nil
}

func (r *dynamoMonitorRepository) ListMonitorRuns(ctx context.Context, tenantID, serviceID, monitorID string, limit int32) ([]resultstatus.CheckRun, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.MonitorPK(tenantID, serviceID, monitorID)},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "RUN#"},
		},
		ScanIndexForward: sharedaws.Bool(false),
		Limit:            sharedaws.Int32(limit),
	})
	if err != nil {
		return nil, err
	}
	runs := make([]resultstatus.CheckRun, 0, len(out.Items))
	for _, item := range out.Items {
		var record resultstatus.CheckRunRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		startedAt, err := time.Parse(time.RFC3339, record.StartedAt)
		if err != nil {
			return nil, err
		}
		finishedAt, err := time.Parse(time.RFC3339, record.FinishedAt)
		if err != nil {
			return nil, err
		}
		runs = append(runs, resultstatus.CheckRun{
			ServiceID:       record.ServiceID,
			MonitorID:       record.MonitorID,
			TenantID:        record.TenantID,
			RunID:           record.RunID,
			Type:            record.Type,
			ProbeLocationID: record.ProbeLocationID,
			Trigger:         checkexecution.TriggerType(strings.ToLower(record.Trigger)),
			StartedAt:       startedAt,
			FinishedAt:      finishedAt,
			DurationMs:      record.DurationMs,
			Outcome:         checkexecution.Outcome(strings.ToLower(record.Outcome)),
			StatusCode:      record.StatusCode,
			Error:           record.Error,
			TTL:             record.TTL,
		})
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt.After(runs[j].StartedAt) })
	return runs, nil
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
	works := dynamodbrecord.NewExecutionWorkItemRecords(monitor, checkexecution.TriggerTypeManual, runID, acceptedAt)
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, monitor.TenantID, "MONITOR_RUN_REQUESTED", monitor.ServiceID, monitor.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "run", "", runID)
	records := make([]any, 0, len(works)+2)
	for _, work := range works {
		records = append(records, work)
	}
	records = append(records, auditEvent, change)
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
		serviceStatus, err := r.buildServiceStatusRecord(ctx, service, now.UTC().Format(time.RFC3339), nil, map[string]resultstatus.MonitorStatus{monitorStatusKey(monitor.ServiceID, monitor.MonitorID): monitorStatus})
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

func (r *dynamoMonitorRepository) ListIncidents(ctx context.Context, tenantID, status string) ([]dynamodbrecord.IncidentRecord, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "INCIDENT#")
	if err != nil {
		return nil, err
	}
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.IncidentItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != entityIncidentRef {
			continue
		}
		incident := record.ToIncident()
		if matchesIncidentFilter(incident.Status, status) {
			incidents = append(incidents, incident)
		}
	}
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].UpdatedAt > incidents[j].UpdatedAt })
	return incidents, nil
}

func (r *dynamoMonitorRepository) GetIncident(ctx context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	if len(out.Item) == 0 {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	var record dynamodbrecord.IncidentItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	incident := record.ToIncident()
	if !strings.EqualFold(incident.TenantID, tenantID) {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	return incident, true, nil
}

func (r *dynamoMonitorRepository) ListIncidentActivities(ctx context.Context, tenantID, incidentID string) ([]dynamodbrecord.IncidentActivityRecord, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.IncidentPK(incidentID), "ACTIVITY#")
	if err != nil {
		return nil, err
	}
	activities := make([]dynamodbrecord.IncidentActivityRecord, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.IncidentActivityRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityIncidentActivity || !strings.EqualFold(record.TenantID, tenantID) {
			continue
		}
		activities = append(activities, record)
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].Timestamp < activities[j].Timestamp })
	return activities, nil
}

func (r *dynamoMonitorRepository) ListMonitorIncidents(ctx context.Context, tenantID, serviceID, monitorID string) ([]dynamodbrecord.IncidentRecord, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.MonitorPK(tenantID, serviceID, monitorID), "INCIDENT#")
	if err != nil {
		return nil, err
	}
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.IncidentItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityIncident {
			continue
		}
		incidents = append(incidents, record.ToIncident())
	}
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].OpenedAt > incidents[j].OpenedAt })
	return incidents, nil
}

func (r *dynamoMonitorRepository) AcknowledgeIncident(ctx context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, found, err := r.GetIncident(ctx, tenantID, incidentID)
	if err != nil || !found {
		return incident, found, err
	}
	if incident.Status != incidentStatusOpen {
		return dynamodbrecord.IncidentRecord{}, true, errIncidentNotActionable
	}
	incident.Status = incidentStatusAcknowledged
	incident.AcknowledgedAt = now.UTC().Format(time.RFC3339)
	incident.UpdatedAt = incident.AcknowledgedAt
	if err := r.writeIncident(ctx, incident, "INCIDENT_ACKNOWLEDGED", now, incident.AcknowledgedAt); err != nil {
		return dynamodbrecord.IncidentRecord{}, true, err
	}
	return incident, true, nil
}

func (r *dynamoMonitorRepository) ResolveIncident(ctx context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, found, err := r.GetIncident(ctx, tenantID, incidentID)
	if err != nil || !found {
		return incident, found, err
	}
	if incident.Status == incidentStatusResolved {
		return dynamodbrecord.IncidentRecord{}, true, errIncidentNotActionable
	}
	incident.Status = incidentStatusResolved
	incident.ResolvedAt = now.UTC().Format(time.RFC3339)
	incident.UpdatedAt = incident.ResolvedAt
	if err := r.writeIncident(ctx, incident, "INCIDENT_RESOLVED", now, incident.ResolvedAt); err != nil {
		return dynamodbrecord.IncidentRecord{}, true, err
	}
	return incident, true, nil
}

func (r *dynamoMonitorRepository) GetSchedulerConfig(ctx context.Context, tenantID string) (dynamodbrecord.SchedulerConfigRecord, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "SCHEDULER_CONFIG"},
		},
	})
	if err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	if len(out.Item) == 0 {
		return dynamodbrecord.SchedulerConfigRecord{Config: checkexecution.SchedulerConfig{}}, nil
	}
	var record dynamodbrecord.SchedulerConfigItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	return record.ToSchedulerConfig(), nil
}

func (r *dynamoMonitorRepository) UpdateSchedulerConfig(ctx context.Context, tenantID string, config checkexecution.SchedulerConfig, now time.Time) (dynamodbrecord.SchedulerConfigRecord, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	record := dynamodbrecord.NewSchedulerConfigItemRecord(tenantID, config, now)
	items, err := marshalPutItems(r.tableName, record)
	if err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	return record.ToSchedulerConfig(), nil
}

func (r *dynamoMonitorRepository) ListMonitorAuditEvents(ctx context.Context, tenantID, serviceID, monitorID string) ([]auditEventView, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "AUDIT#")
	if err != nil {
		return nil, err
	}
	resourceID := dynamodbrecord.MonitorAuditResourceID(serviceID, monitorID)
	eventsList := make([]auditEventView, 0)
	for _, item := range out {
		var record dynamodbrecord.AuditEventRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityAuditEvent || record.ResourceID != resourceID {
			continue
		}
		eventsList = append(eventsList, auditEventView{
			AuditID:    record.AuditID,
			ServiceID:  record.ServiceID,
			MonitorID:  record.MonitorID,
			EventType:  record.Action,
			OccurredAt: record.Timestamp,
			Actor:      record.Actor,
			Origin:     record.Origin,
		})
	}
	sort.Slice(eventsList, func(i, j int) bool { return eventsList[i].OccurredAt > eventsList[j].OccurredAt })
	return eventsList, nil
}

func (r *dynamoMonitorRepository) ListServiceAuditEvents(ctx context.Context, tenantID, serviceID string) ([]auditEventView, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "AUDIT#")
	if err != nil {
		return nil, err
	}
	resourceID := dynamodbrecord.MonitorAuditResourceID(serviceID, "")
	eventsList := make([]auditEventView, 0)
	for _, item := range out {
		var record dynamodbrecord.AuditEventRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityAuditEvent || record.ResourceID != resourceID {
			continue
		}
		eventsList = append(eventsList, auditEventView{
			AuditID:    record.AuditID,
			ServiceID:  record.ServiceID,
			MonitorID:  record.MonitorID,
			EventType:  record.Action,
			OccurredAt: record.Timestamp,
			Actor:      record.Actor,
			Origin:     record.Origin,
		})
	}
	sort.Slice(eventsList, func(i, j int) bool { return eventsList[i].OccurredAt > eventsList[j].OccurredAt })
	return eventsList, nil
}

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
		byID[monitorStatusKey(monitor.ServiceID, monitor.MonitorID)] = monitor
	}
	for _, monitor := range monitorsOverride {
		if _, skip := excluded[strings.ToLower(monitor.MonitorID)]; skip {
			continue
		}
		byID[monitorStatusKey(monitor.ServiceID, monitor.MonitorID)] = monitor
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
		key := monitorStatusKey(monitor.ServiceID, monitor.MonitorID)
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

func (r *dynamoMonitorRepository) queryPartition(ctx context.Context, pk, prefix string) ([]map[string]sharedaws.AttributeValue, error) {
	input := &sharedaws.DynamoDBQueryInput{
		TableName: sharedaws.String(r.tableName),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk": &sharedaws.AttributeValueMemberS{Value: pk},
		},
	}
	if prefix == "" {
		input.KeyConditionExpression = sharedaws.String("PK = :pk")
	} else {
		input.KeyConditionExpression = sharedaws.String("PK = :pk AND begins_with(SK, :prefix)")
		input.ExpressionAttributeValues[":prefix"] = &sharedaws.AttributeValueMemberS{Value: prefix}
	}
	out, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (r *dynamoMonitorRepository) deleteKeysAndPut(ctx context.Context, keys []ddbKey, puts []sharedaws.TransactWriteItem) error {
	if len(puts) > 25 {
		return fmt.Errorf("too many put items for transaction: %d", len(puts))
	}
	if len(keys) == 0 {
		return r.writeTransaction(ctx, puts)
	}
	maxDeletesPerTransaction := 25
	if len(puts) > 0 {
		maxDeletesPerTransaction = 25 - len(puts)
	}
	if maxDeletesPerTransaction <= 0 {
		return fmt.Errorf("too many put items for transaction with deletes: %d", len(puts))
	}
	for start := 0; start < len(keys); start += maxDeletesPerTransaction {
		end := start + maxDeletesPerTransaction
		if end > len(keys) {
			end = len(keys)
		}
		items := make([]sharedaws.TransactWriteItem, 0, end-start+1)
		for _, key := range keys[start:end] {
			items = append(items, sharedaws.TransactWriteItem{Delete: &sharedaws.Delete{
				TableName: sharedaws.String(r.tableName),
				Key: map[string]sharedaws.AttributeValue{
					"PK": &sharedaws.AttributeValueMemberS{Value: key.PK},
					"SK": &sharedaws.AttributeValueMemberS{Value: key.SK},
				},
			}})
		}
		if end == len(keys) {
			items = append(items, puts...)
		}
		if err := r.writeTransaction(ctx, items); err != nil {
			return err
		}
	}
	return nil
}

func (r *dynamoMonitorRepository) writeTransaction(ctx context.Context, items []sharedaws.TransactWriteItem) error {
	if len(items) == 0 {
		return nil
	}
	_, err := r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	return err
}

func (r *dynamoMonitorRepository) writeIncident(ctx context.Context, incident dynamodbrecord.IncidentRecord, action string, now time.Time, changeValue string) error {
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, incident.TenantID, action, incident.ServiceID, incident.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "incident", "", changeValue)
	activity := dynamodbrecord.NewIncidentActivityRecord(incident.TenantID, incident.IncidentID, newActivityID(now), action, now)
	items, err := marshalPutItems(
		r.tableName,
		dynamodbrecord.NewIncidentMonitorItemRecord(incident),
		dynamodbrecord.NewIncidentRefItemRecord(incident),
		dynamodbrecord.NewIncidentMetaItemRecord(incident),
		activity,
		auditEvent,
		change,
	)
	if err != nil {
		return err
	}
	return r.writeTransaction(ctx, items)
}

func (r *dynamoMonitorRepository) requireTableName() error {
	if strings.TrimSpace(r.tableName) == "" {
		return errMissingTableName
	}
	return nil
}

func newDefaultMonitorStatusRecord(monitor monitorconfig.Monitor, now string) resultstatus.MonitorStatusRecord {
	status := resultstatus.MonitorStatus{
		ServiceID:           monitor.ServiceID,
		MonitorID:           monitor.MonitorID,
		TenantID:            monitor.TenantID,
		CurrentStatus:       strings.ToUpper(rollupUnknown),
		LastCheckedAt:       mustParseTime(now),
		LastDurationMs:      0,
		LastProbeLocationID: "",
		LastError:           "",
		LastOutcome:         checkexecution.Outcome(rollupUnknown),
	}
	return status.ToRecord()
}

func monitorStatusRecordToDomain(record resultstatus.MonitorStatusRecord) resultstatus.MonitorStatus {
	lastCheckedAt, _ := time.Parse(time.RFC3339, firstNonEmpty(record.LastCheckedAt, record.UpdatedAt))
	return resultstatus.MonitorStatus{
		ServiceID:           record.ServiceID,
		MonitorID:           record.MonitorID,
		TenantID:            record.TenantID,
		CurrentStatus:       record.CurrentStatus,
		LastCheckedAt:       lastCheckedAt,
		LastDurationMs:      record.LastDurationMs,
		LastProbeLocationID: record.LastProbeLocationID,
		LastError:           record.LastError,
		LastOutcome:         checkexecution.Outcome(strings.ToLower(firstNonEmpty(record.LastOutcome, rollupUnknown))),
	}
}

type ddbKey struct {
	PK string
	SK string
}

type deleteKeySet struct {
	seen map[string]ddbKey
}

func newDeleteKeySet() *deleteKeySet {
	return &deleteKeySet{seen: map[string]ddbKey{}}
}

func (s *deleteKeySet) add(pk, sk string) {
	if strings.TrimSpace(pk) == "" || strings.TrimSpace(sk) == "" {
		return
	}
	s.seen[pk+"\x00"+sk] = ddbKey{PK: pk, SK: sk}
}

func (s *deleteKeySet) addItems(items []map[string]sharedaws.AttributeValue) {
	for _, item := range items {
		s.addItem(item)
	}
}

func (s *deleteKeySet) addItem(item map[string]sharedaws.AttributeValue) {
	pk, ok1 := item["PK"].(*sharedaws.AttributeValueMemberS)
	sk, ok2 := item["SK"].(*sharedaws.AttributeValueMemberS)
	if ok1 && ok2 {
		s.add(pk.Value, sk.Value)
	}
}

func (s *deleteKeySet) list() []ddbKey {
	out := make([]ddbKey, 0, len(s.seen))
	for _, key := range s.seen {
		out = append(out, key)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PK == out[j].PK {
			return out[i].SK < out[j].SK
		}
		return out[i].PK < out[j].PK
	})
	return out
}

func marshalPutItems(tableName string, records ...any) ([]sharedaws.TransactWriteItem, error) {
	items := make([]sharedaws.TransactWriteItem, 0, len(records))
	for _, record := range records {
		item, err := sharedaws.MarshalMap(record)
		if err != nil {
			return nil, err
		}
		items = append(items, sharedaws.TransactWriteItem{Put: &sharedaws.Put{TableName: sharedaws.String(tableName), Item: item}})
	}
	return items, nil
}

func matchesIncidentFilter(incidentStatus, filter string) bool {
	switch strings.ToLower(filter) {
	case "", "all":
		return true
	case "open":
		return incidentStatus == incidentStatusOpen || incidentStatus == incidentStatusAcknowledged
	case "closed":
		return incidentStatus == incidentStatusResolved
	default:
		return true
	}
}

func deriveServiceRollup(lifecycle monitorconfig.ServiceLifecycle, summaries []monitorconfig.MonitorSummary) string {
	enabled := make([]monitorconfig.MonitorSummary, 0)
	for _, summary := range summaries {
		if summary.Enabled {
			enabled = append(enabled, summary)
		}
	}
	if len(enabled) == 0 {
		return rollupPaused
	}
	upCount := 0
	downCount := 0
	unknownCount := 0
	for _, summary := range enabled {
		switch strings.ToLower(strings.TrimSpace(summary.CurrentStatus)) {
		case "up":
			upCount++
		case "down":
			downCount++
		default:
			unknownCount++
		}
	}
	if unknownCount == len(enabled) {
		return rollupUnknown
	}
	if upCount == len(enabled) {
		return rollupUp
	}
	if downCount == len(enabled) {
		return rollupDown
	}
	return rollupDegraded
}

func buildMonitorSummaries(monitors []monitorconfig.Monitor, statuses map[string]resultstatus.MonitorStatus) []monitorconfig.MonitorSummary {
	summaries := make([]monitorconfig.MonitorSummary, 0, len(monitors))
	for _, monitor := range monitors {
		summary := monitorconfig.MonitorSummary{
			TenantID:        monitor.TenantID,
			ServiceID:       monitor.ServiceID,
			MonitorID:       monitor.MonitorID,
			Name:            monitor.Name,
			Type:            monitor.Type,
			Enabled:         monitor.Enabled,
			IntervalSeconds: monitor.IntervalSeconds,
			ProbeLocations:  append([]string(nil), monitor.ProbeLocations...),
		}
		if status, ok := statuses[monitorStatusKey(monitor.ServiceID, monitor.MonitorID)]; ok {
			summary.CurrentStatus = strings.ToLower(status.CurrentStatus)
			summary.LastCheckedAt = status.LastCheckedAt.UTC().Format(time.RFC3339)
			summary.LastDurationMs = status.LastDurationMs
			summary.LastProbeLocationID = status.LastProbeLocationID
			summary.LastError = status.LastError
			summary.UpdatedAt = summary.LastCheckedAt
		}
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].MonitorID < summaries[j].MonitorID })
	return summaries
}

func buildServiceCardMetrics(monitors []monitorconfig.Monitor, statuses map[string]resultstatus.MonitorStatus, runsByMonitor map[string][]resultstatus.CheckRun) serviceCardMetricsResponse {
	metrics := serviceCardMetricsResponse{
		MonitorCount: len(monitors),
		Trend:        []serviceCardTrendPoint{},
	}
	if len(monitors) == 0 {
		metrics.State = serviceCardMetricStateNoMonitors
		return metrics
	}
	for _, status := range statuses {
		if strings.EqualFold(status.CurrentStatus, string(resultstatus.MonitorStateUp)) {
			metrics.UpMonitorCount++
		}
	}
	successDurations := []int64{}
	for _, monitor := range monitors {
		for _, run := range runsByMonitor[monitor.MonitorID] {
			metrics.SampleCount++
			success := run.Outcome == checkexecution.OutcomeSuccess
			if success {
				metrics.SuccessCount++
				successDurations = append(successDurations, run.DurationMs)
			}
			metrics.Trend = append(metrics.Trend, serviceCardTrendPoint{
				MonitorID:  run.MonitorID,
				StartedAt:  run.StartedAt.UTC().Format(time.RFC3339),
				DurationMs: run.DurationMs,
				Outcome:    string(run.Outcome),
				Success:    success,
			})
		}
	}
	if metrics.SampleCount == 0 {
		metrics.State = serviceCardMetricStateNoData
		return metrics
	}
	metrics.State = serviceCardMetricStateReady
	uptime := float64(metrics.SuccessCount) / float64(metrics.SampleCount) * 100
	metrics.RecentUptimePct = &uptime
	if len(successDurations) > 0 {
		avg := averageInt64(successDurations)
		p99 := percentileNearestRank(successDurations, 99)
		metrics.AvgLatencyMs = &avg
		metrics.P99LatencyMs = &p99
	}
	sort.Slice(metrics.Trend, func(i, j int) bool { return metrics.Trend[i].StartedAt < metrics.Trend[j].StartedAt })
	return metrics
}

func averageInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	var total int64
	for _, value := range values {
		total += value
	}
	return total / int64(len(values))
}

func percentileNearestRank(values []int64, percentile int) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := (percentile*len(sorted)+99)/100 - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func countEnabledMonitors(monitors []monitorconfig.Monitor) int {
	count := 0
	for _, monitor := range monitors {
		if monitor.Enabled {
			count++
		}
	}
	return count
}

func monitorStatusKey(serviceID, monitorID string) string {
	return strings.ToLower(strings.TrimSpace(serviceID)) + "/" + strings.ToLower(strings.TrimSpace(monitorID))
}

func mustParseTime(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339, value)
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func (r *dynamoMonitorRepository) CreateNotificationChannel(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.NotificationChannel{}, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	channel.CreatedAt = now
	channel.UpdatedAt = now
	av, err := sharedaws.MarshalMap(newNotificationChannelItemRecord(channel))
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	return channel, nil
}

func (r *dynamoMonitorRepository) ListNotificationChannels(ctx context.Context, tenantID string) ([]escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "NOTIFICATION_CHANNEL#")
	if err != nil {
		return nil, err
	}
	channels := make([]escalation.NotificationChannel, 0, len(out))
	for _, item := range out {
		var record notificationChannelItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityNotificationChannel {
			continue
		}
		channels = append(channels, record.toNotificationChannel())
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].Name < channels[j].Name })
	return channels, nil
}

func (r *dynamoMonitorRepository) GetNotificationChannel(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: notificationChannelSK(channelID)}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record notificationChannelItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	channel := record.toNotificationChannel()
	return &channel, nil
}

func (r *dynamoMonitorRepository) UpdateNotificationChannel(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.NotificationChannel{}, err
	}
	existing, err := r.GetNotificationChannel(ctx, channel.TenantID, channel.ChannelID)
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	if existing != nil && strings.TrimSpace(existing.CreatedAt) != "" {
		channel.CreatedAt = existing.CreatedAt
	}
	channel.UpdatedAt = r.now().UTC().Format(time.RFC3339)
	av, err := sharedaws.MarshalMap(newNotificationChannelItemRecord(channel))
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	return channel, nil
}

func (r *dynamoMonitorRepository) DeleteNotificationChannel(ctx context.Context, tenantID, channelID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: notificationChannelSK(channelID)}}})
	return err
}

func (r *dynamoMonitorRepository) RecordNotificationChannelTestAudit(ctx context.Context, tenantID, channelID, channelType, outcome, reason string, now time.Time) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, tenantID, "NOTIFICATION_CHANNEL_TEST_SENT", "notification-channel", channelID)
	records := []any{
		auditEvent,
		dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "channelType", "", channelType),
		dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "outcome", "", outcome),
	}
	if strings.TrimSpace(reason) != "" {
		records = append(records, dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "reason", "", reason))
	}
	items, err := marshalPutItems(r.tableName, records...)
	if err != nil {
		return err
	}
	return r.writeTransaction(ctx, items)
}

func (r *dynamoMonitorRepository) ChannelsReferencedByRoutes(ctx context.Context, tenantID, channelID string) ([]routeReference, error) {
	policies, err := r.ListEscalationPolicies(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	references := []routeReference{}
	for _, policy := range policies {
		if policyReferencesChannel(policy, channelID) {
			references = append(references, routeReference{PolicyID: policy.PolicyID, Name: policy.Name})
		}
	}
	return references, nil
}

func policyReferencesChannel(policy escalation.EscalationPolicy, channelID string) bool {
	needle := strings.TrimSpace(channelID)
	for _, path := range []escalation.EscalationPath{policy.BusinessHoursPath, policy.OffHoursPath} {
		for _, step := range path.Steps {
			if strings.EqualFold(step.ChannelID, needle) {
				return true
			}
		}
	}
	return false
}

func (r *dynamoMonitorRepository) CreateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	policy.CreatedAt = now
	policy.UpdatedAt = now
	av, err := sharedaws.MarshalMap(newEscalationPolicyItemRecord(policy))
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return policy, nil
}

func (r *dynamoMonitorRepository) ListEscalationPolicies(ctx context.Context, tenantID string) ([]escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "ESCALATION_POLICY#")
	if err != nil {
		return nil, err
	}
	policies := make([]escalation.EscalationPolicy, 0, len(out))
	for _, item := range out {
		var record escalationPolicyItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityEscalationPolicy {
			continue
		}
		policy := record.toEscalationPolicy()
		if err := r.MigrateRouteInlineChannels(ctx, &policy); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	sort.Slice(policies, func(i, j int) bool { return policies[i].PolicyID < policies[j].PolicyID })
	return policies, nil
}

func (r *dynamoMonitorRepository) GetEscalationPolicy(ctx context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: escalationPolicySK(policyID)}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record escalationPolicyItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	policy := record.toEscalationPolicy()
	if err := r.MigrateRouteInlineChannels(ctx, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *dynamoMonitorRepository) MigrateRouteInlineChannels(ctx context.Context, policy *escalation.EscalationPolicy) error {
	if policy == nil {
		return nil
	}
	migrated := false
	stepIndex := 0
	paths := []*escalation.EscalationPath{&policy.BusinessHoursPath, &policy.OffHoursPath}
	for _, path := range paths {
		for i := range path.Steps {
			step := &path.Steps[i]
			if strings.TrimSpace(step.ChannelID) != "" || len(step.Channels) == 0 {
				stepIndex++
				continue
			}
			legacy := step.Channels[0]
			channelID := fmt.Sprintf("%s#%s#%d", strings.ToUpper(strings.TrimSpace(policy.TenantID)), strings.ToUpper(strings.TrimSpace(policy.PolicyID)), stepIndex)
			channel := escalation.NotificationChannel{
				TenantID:  policy.TenantID,
				ChannelID: channelID,
				Name:      fmt.Sprintf("Migrated channel %d", stepIndex+1),
				Type:      legacy.Type,
				Target:    legacy.Target,
				Config:    append(json.RawMessage(nil), legacy.Config...),
				CreatedAt: policy.CreatedAt,
				UpdatedAt: policy.UpdatedAt,
			}
			if _, err := r.CreateNotificationChannel(ctx, channel); err != nil {
				return err
			}
			step.ChannelID = channelID
			step.Channels = nil
			migrated = true
			stepIndex++
		}
	}
	if !migrated {
		return nil
	}
	av, err := sharedaws.MarshalMap(newEscalationPolicyItemRecord(*policy))
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	return err
}

func (r *dynamoMonitorRepository) UpdateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	existing, err := r.GetEscalationPolicy(ctx, policy.TenantID, policy.PolicyID)
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	if existing != nil && strings.TrimSpace(existing.CreatedAt) != "" {
		policy.CreatedAt = existing.CreatedAt
	}
	policy.UpdatedAt = r.now().UTC().Format(time.RFC3339)
	av, err := sharedaws.MarshalMap(newEscalationPolicyItemRecord(policy))
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return policy, nil
}

func (r *dynamoMonitorRepository) DeleteEscalationPolicy(ctx context.Context, tenantID, policyID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: escalationPolicySK(policyID)}}})
	return err
}

func (r *dynamoMonitorRepository) ServiceReferencesEscalationPolicy(ctx context.Context, tenantID, policyID string) (bool, error) {
	services, err := r.ListServices(ctx, tenantID)
	if err != nil {
		return false, err
	}
	for _, service := range services {
		if strings.EqualFold(service.EscalationPolicyID, policyID) {
			return true, nil
		}
	}
	return false, nil
}

func (r *dynamoMonitorRepository) GetEscalationState(ctx context.Context, tenantID, incidentID string) (*escalation.EscalationState, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)}, "SK": &sharedaws.AttributeValueMemberS{Value: "ESCALATION_STATE"}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record escalationStateItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	state := record.toEscalationState()
	return &state, nil
}

type escalationStateItemRecord struct {
	PK           string                      `dynamodbav:"PK"`
	SK           string                      `dynamodbav:"SK"`
	EntityType   string                      `dynamodbav:"EntityType"`
	TenantID     string                      `dynamodbav:"TenantID"`
	IncidentID   string                      `dynamodbav:"IncidentID"`
	PolicyID     string                      `dynamodbav:"PolicyID"`
	ServiceID    string                      `dynamodbav:"ServiceID"`
	MonitorID    string                      `dynamodbav:"MonitorID"`
	CurrentStep  int                         `dynamodbav:"CurrentStep"`
	StepsFired   []int                       `dynamodbav:"StepsFired"`
	SelectedPath string                      `dynamodbav:"SelectedPath"`
	ScheduledFor string                      `dynamodbav:"ScheduledFor,omitempty"`
	Status       escalation.EscalationStatus `dynamodbav:"Status"`
	CreatedAt    string                      `dynamodbav:"CreatedAt"`
	UpdatedAt    string                      `dynamodbav:"UpdatedAt"`
}

func (r escalationStateItemRecord) toEscalationState() escalation.EscalationState {
	return escalation.EscalationState{
		TenantID:     r.TenantID,
		IncidentID:   r.IncidentID,
		PolicyID:     r.PolicyID,
		ServiceID:    r.ServiceID,
		MonitorID:    r.MonitorID,
		CurrentStep:  r.CurrentStep,
		StepsFired:   append([]int(nil), r.StepsFired...),
		SelectedPath: r.SelectedPath,
		ScheduledFor: r.ScheduledFor,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

type escalationPolicyItemRecord struct {
	PK                string                    `dynamodbav:"PK"`
	SK                string                    `dynamodbav:"SK"`
	EntityType        string                    `dynamodbav:"EntityType"`
	TenantID          string                    `dynamodbav:"TenantID"`
	PolicyID          string                    `dynamodbav:"PolicyID"`
	Name              string                    `dynamodbav:"Name"`
	Description       string                    `dynamodbav:"Description,omitempty"`
	BusinessHoursPath escalation.EscalationPath `dynamodbav:"BusinessHoursPath"`
	OffHoursPath      escalation.EscalationPath `dynamodbav:"OffHoursPath"`
	CreatedAt         string                    `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt         string                    `dynamodbav:"UpdatedAt,omitempty"`
}

type notificationChannelItemRecord struct {
	PK         string                 `dynamodbav:"PK"`
	SK         string                 `dynamodbav:"SK"`
	EntityType string                 `dynamodbav:"EntityType"`
	TenantID   string                 `dynamodbav:"TenantID"`
	ChannelID  string                 `dynamodbav:"ChannelID"`
	Name       string                 `dynamodbav:"Name"`
	Type       escalation.ChannelType `dynamodbav:"Type"`
	Target     string                 `dynamodbav:"Target"`
	Config     []byte                 `dynamodbav:"Config,omitempty"`
	CreatedAt  string                 `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt  string                 `dynamodbav:"UpdatedAt,omitempty"`
}

func escalationPolicySK(policyID string) string {
	return "ESCALATION_POLICY#" + strings.ToUpper(strings.TrimSpace(policyID))
}

func notificationChannelSK(channelID string) string {
	return "NOTIFICATION_CHANNEL#" + strings.ToUpper(strings.TrimSpace(channelID))
}

func newNotificationChannelItemRecord(channel escalation.NotificationChannel) notificationChannelItemRecord {
	item := dynamodbschema.NotificationChannelItem(channel.TenantID, channel.ChannelID)
	return notificationChannelItemRecord{
		PK:         item.PK,
		SK:         item.SK,
		EntityType: item.EntityType,
		TenantID:   strings.ToUpper(channel.TenantID),
		ChannelID:  strings.ToUpper(strings.TrimSpace(channel.ChannelID)),
		Name:       strings.TrimSpace(channel.Name),
		Type:       channel.Type,
		Target:     strings.TrimSpace(channel.Target),
		Config:     append([]byte(nil), channel.Config...),
		CreatedAt:  channel.CreatedAt,
		UpdatedAt:  channel.UpdatedAt,
	}
}

func (r notificationChannelItemRecord) toNotificationChannel() escalation.NotificationChannel {
	return escalation.NotificationChannel{
		TenantID:  r.TenantID,
		ChannelID: r.ChannelID,
		Name:      r.Name,
		Type:      r.Type,
		Target:    r.Target,
		Config:    append(json.RawMessage(nil), r.Config...),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func newEscalationPolicyItemRecord(policy escalation.EscalationPolicy) escalationPolicyItemRecord {
	return escalationPolicyItemRecord{
		PK:                dynamodbschema.TenantPK(policy.TenantID),
		SK:                escalationPolicySK(policy.PolicyID),
		EntityType:        dynamodbschema.EntityEscalationPolicy,
		TenantID:          strings.ToUpper(policy.TenantID),
		PolicyID:          strings.ToUpper(strings.TrimSpace(policy.PolicyID)),
		Name:              strings.TrimSpace(policy.Name),
		Description:       strings.TrimSpace(policy.Description),
		BusinessHoursPath: cloneEscalationPath(policy.BusinessHoursPath),
		OffHoursPath:      cloneEscalationPath(policy.OffHoursPath),
		CreatedAt:         policy.CreatedAt,
		UpdatedAt:         policy.UpdatedAt,
	}
}

func (r escalationPolicyItemRecord) toEscalationPolicy() escalation.EscalationPolicy {
	return escalation.EscalationPolicy{
		TenantID:          r.TenantID,
		PolicyID:          r.PolicyID,
		Name:              r.Name,
		Description:       r.Description,
		BusinessHoursPath: cloneEscalationPath(r.BusinessHoursPath),
		OffHoursPath:      cloneEscalationPath(r.OffHoursPath),
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}
