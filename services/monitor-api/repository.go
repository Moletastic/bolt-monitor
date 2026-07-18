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
	"bolt-monitor/shared/domainvalues"
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

type historyPage[T any] struct {
	Items   []T
	NextKey map[string]sharedaws.AttributeValue
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
	if err := r.replaceSearchIndex(ctx, service.TenantID, searchResourceService, service.ServiceID, "", buildServiceSearchRecords(service)); err != nil {
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
			statuses[monitorStatusMapKey(monitor)] = status
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
	if err := r.replaceSearchIndex(ctx, service.TenantID, searchResourceService, service.ServiceID, "", buildServiceSearchRecords(service)); err != nil {
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
	if err := r.deleteSearchIndex(ctx, tenantID, searchResourceService, serviceID, ""); err != nil {
		return true, err
	}
	for _, monitor := range monitors {
		if err := r.deleteSearchIndex(ctx, tenantID, searchResourceMonitor, monitor.MonitorID, serviceID); err != nil {
			return true, err
		}
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
	if err := r.replaceSearchIndex(ctx, service.TenantID, searchResourceService, service.ServiceID, "", buildServiceSearchRecords(service)); err != nil {
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
	if err := r.replaceSearchIndex(ctx, service.TenantID, searchResourceService, service.ServiceID, "", buildServiceSearchRecords(service)); err != nil {
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
	page, err := r.ListMonitorAuditEventsPage(ctx, tenantID, serviceID, monitorID, 20, nil)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *dynamoMonitorRepository) ListMonitorAuditEventsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	if err := r.requireTableName(); err != nil {
		return historyPage[auditEventView]{}, err
	}
	resource := dynamodbschema.AuditResourceItem(tenantID, serviceID, monitorID, "cursor", "cursor")
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		IndexName:              sharedaws.String(dynamodbschema.GSIAuditByResource),
		KeyConditionExpression: sharedaws.String("GSI3PK = :pk AND begins_with(GSI3SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: resource.GSI3PK},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "AUDIT#"},
		},
		ScanIndexForward:  sharedaws.Bool(false),
		Limit:             sharedaws.Int32(limit),
		ExclusiveStartKey: startKey,
	})
	if err != nil {
		return historyPage[auditEventView]{}, err
	}
	eventsList := make([]auditEventView, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.AuditEventRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return historyPage[auditEventView]{}, err
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
	return historyPage[auditEventView]{Items: eventsList, NextKey: out.LastEvaluatedKey}, nil
}

func (r *dynamoMonitorRepository) ListServiceAuditEvents(ctx context.Context, tenantID, serviceID string) ([]auditEventView, error) {
	page, err := r.ListServiceAuditEventsPage(ctx, tenantID, serviceID, 20, nil)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *dynamoMonitorRepository) ListServiceAuditEventsPage(ctx context.Context, tenantID, serviceID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	if err := r.requireTableName(); err != nil {
		return historyPage[auditEventView]{}, err
	}
	resource := dynamodbschema.AuditResourceItem(tenantID, serviceID, "", "cursor", "cursor")
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		IndexName:              sharedaws.String(dynamodbschema.GSIAuditByResource),
		KeyConditionExpression: sharedaws.String("GSI3PK = :pk AND begins_with(GSI3SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: resource.GSI3PK},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "AUDIT#"},
		},
		ScanIndexForward:  sharedaws.Bool(false),
		Limit:             sharedaws.Int32(limit),
		ExclusiveStartKey: startKey,
	})
	if err != nil {
		return historyPage[auditEventView]{}, err
	}
	eventsList := make([]auditEventView, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.AuditEventRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return historyPage[auditEventView]{}, err
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
	return historyPage[auditEventView]{Items: eventsList, NextKey: out.LastEvaluatedKey}, nil
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

func (r *dynamoMonitorRepository) queryPartition(ctx context.Context, pk, prefix string) ([]map[string]sharedaws.AttributeValue, error) {
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, pk, prefix, sharedaws.PageOptions{})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
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
		}
		if status, ok := statuses[monitorStatusMapKey(monitor)]; ok {
			summary.CurrentStatus = strings.ToLower(status.CurrentStatus)
			summary.LastCheckedAt = status.LastCheckedAt.UTC().Format(time.RFC3339)
			summary.LastDurationMs = status.LastDurationMs
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

func monitorStatusMapKey(monitor monitorconfig.Monitor) string {
	return domainvalues.MustMonitorRef(
		domainvalues.TenantID(monitor.TenantID),
		domainvalues.ServiceID(monitor.ServiceID),
		domainvalues.MonitorID(monitor.MonitorID),
	).StatusMapKey()
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
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewNotificationChannelItemRecord(channel))
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	if err := r.replaceSearchIndex(ctx, channel.TenantID, searchResourceChannel, channel.ChannelID, "", buildChannelSearchRecords(channel)); err != nil {
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
		var record dynamodbrecord.NotificationChannelItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityNotificationChannel {
			continue
		}
		channels = append(channels, record.ToNotificationChannel())
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].Name < channels[j].Name })
	return channels, nil
}

func (r *dynamoMonitorRepository) GetNotificationChannel(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.NotificationChannelSK(channelID)}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record dynamodbrecord.NotificationChannelItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	channel := record.ToNotificationChannel()
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
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewNotificationChannelItemRecord(channel))
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	if err := r.replaceSearchIndex(ctx, channel.TenantID, searchResourceChannel, channel.ChannelID, "", buildChannelSearchRecords(channel)); err != nil {
		return escalation.NotificationChannel{}, err
	}
	return channel, nil
}

func (r *dynamoMonitorRepository) DeleteNotificationChannel(ctx context.Context, tenantID, channelID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.NotificationChannelSK(channelID)}}})
	if err != nil {
		return err
	}
	return r.deleteSearchIndex(ctx, tenantID, searchResourceChannel, channelID, "")
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
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPolicyItemRecord(policy))
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	if err := r.replaceSearchIndex(ctx, policy.TenantID, searchResourcePolicy, policy.PolicyID, "", buildPolicySearchRecords(policy)); err != nil {
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
		var record dynamodbrecord.EscalationPolicyItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityEscalationPolicy {
			continue
		}
		policy := record.ToEscalationPolicy()
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
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.EscalationPolicySK(policyID)}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record dynamodbrecord.EscalationPolicyItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	policy := record.ToEscalationPolicy()
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
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPolicyItemRecord(*policy))
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
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPolicyItemRecord(policy))
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	if err := r.replaceSearchIndex(ctx, policy.TenantID, searchResourcePolicy, policy.PolicyID, "", buildPolicySearchRecords(policy)); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return policy, nil
}

func (r *dynamoMonitorRepository) DeleteEscalationPolicy(ctx context.Context, tenantID, policyID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.EscalationPolicySK(policyID)}}})
	if err != nil {
		return err
	}
	return r.deleteSearchIndex(ctx, tenantID, searchResourcePolicy, policyID, "")
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
	var record dynamodbrecord.EscalationStateItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	state := record.ToEscalationState()
	return &state, nil
}
