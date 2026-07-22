package main

import (
	"context"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

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
