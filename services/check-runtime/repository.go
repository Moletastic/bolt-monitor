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
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type dynamoAPI = sharedaws.DynamoDBAPI

type dynamoRuntimeRepository struct {
	client    dynamoAPI
	tableName string
	now       func() time.Time
}

func newDynamoRuntimeRepository(client dynamoAPI, tableName string) *dynamoRuntimeRepository {
	return &dynamoRuntimeRepository{client: client, tableName: tableName, now: time.Now}
}

func (r *dynamoRuntimeRepository) GetSchedulerConfig(ctx context.Context, tenantID string) (checkexecution.SchedulerConfig, error) {
	if err := r.requireTableName(); err != nil {
		return checkexecution.SchedulerConfig{}, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "SCHEDULER_CONFIG"},
		},
	})
	if err != nil {
		return checkexecution.SchedulerConfig{}, err
	}
	if len(out.Item) == 0 {
		return checkexecution.SchedulerConfig{}, nil
	}
	var record dynamodbrecord.SchedulerConfigItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return checkexecution.SchedulerConfig{}, err
	}
	return checkexecution.SchedulerConfig{RecurringEnabled: record.RecurringEnabled, StopControlMode: checkexecution.StopControlMode(record.StopControlMode)}, nil
}

func (r *dynamoRuntimeRepository) ListMonitors(ctx context.Context, tenantID string) ([]monitorconfig.Monitor, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	serviceRefs, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "SERVICE#"},
		},
	})
	if err != nil {
		return nil, err
	}
	monitors := make([]monitorconfig.Monitor, 0)
	for _, item := range serviceRefs.Items {
		var service dynamodbrecord.ServiceItemRecord
		if err := sharedaws.UnmarshalMap(item, &service); err != nil {
			return nil, err
		}
		if service.EntityType != dynamodbschema.EntityServiceRef {
			continue
		}
		serviceMonitors, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
			TableName:              sharedaws.String(r.tableName),
			KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
			ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
				":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.ServicePK(tenantID, service.ServiceID)},
				":prefix": &sharedaws.AttributeValueMemberS{Value: "MONITOR#"},
			},
		})
		if err != nil {
			return nil, err
		}
		for _, serviceMonitorItem := range serviceMonitors.Items {
			var record dynamodbrecord.MonitorItemRecord
			if err := sharedaws.UnmarshalMap(serviceMonitorItem, &record); err != nil {
				return nil, err
			}
			if record.EntityType != dynamodbschema.EntityServiceMonitorRef {
				continue
			}
			monitors = append(monitors, record.ToMonitor())
		}
	}
	return monitors, nil
}

func (r *dynamoRuntimeRepository) GetLastExecution(ctx context.Context, tenantID, serviceID, monitorID string) (*time.Time, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	record, found, err := r.getMonitorRecord(ctx, tenantID, serviceID, monitorID)
	if err != nil || !found || strings.TrimSpace(record.LastExecutionAt) == "" {
		return nil, err
	}
	lastExecution, err := time.Parse(time.RFC3339, record.LastExecutionAt)
	if err != nil {
		return nil, err
	}
	return &lastExecution, nil
}

func (r *dynamoRuntimeRepository) RecordLastExecution(ctx context.Context, tenantID, serviceID, monitorID string, lastExec time.Time) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	record, found, err := r.getMonitorRecord(ctx, tenantID, serviceID, monitorID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("monitor %s/%s not found", serviceID, monitorID)
	}
	record.LastExecutionAt = lastExec.UTC().Format(time.RFC3339)
	items, err := marshalItems(r.tableName, record)
	if err != nil {
		return err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	return err
}

func (r *dynamoRuntimeRepository) EnqueueExecutionRequests(ctx context.Context, requests []checkexecution.ExecutionRequest, now time.Time) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	acceptedAt := now.UTC().Format(time.RFC3339)
	records := make([]any, 0, len(requests))
	for _, request := range requests {
		runID := request.RunID
		if strings.TrimSpace(runID) == "" {
			runID = newRunID(now)
		}
		records = append(records, dynamodbrecord.NewExecutionWorkItemRecord(request.Monitor.TenantID, request.Monitor.ServiceID, request.Monitor.MonitorID, runID, request.Trigger, acceptedAt, checkexecution.ExecutionWorkPending, nil, nil, ""))
	}
	items, err := marshalItems(r.tableName, records...)
	if err != nil {
		return err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	return err
}

func (r *dynamoRuntimeRepository) ListPendingExecutionWork(ctx context.Context, tenantID string, limit int32) ([]checkexecution.ExecutionWork, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "RUN_REQUEST#"},
		},
		Limit: sharedaws.Int32(limit),
	})
	if err != nil {
		return nil, err
	}
	works := make([]checkexecution.ExecutionWork, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.ExecutionWorkItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityExecutionWork {
			continue
		}
		work, err := record.ToWork()
		if err != nil {
			return nil, err
		}
		if work.Status == checkexecution.ExecutionWorkPending {
			works = append(works, work)
		}
	}
	sortWorksByRequestedAt(works)
	return works, nil
}

func (r *dynamoRuntimeRepository) ClaimExecutionWork(ctx context.Context, work checkexecution.ExecutionWork, now time.Time) (bool, error) {
	if err := r.requireTableName(); err != nil {
		return false, err
	}
	startedAt := now.UTC()
	updated := work
	updated.Status = checkexecution.ExecutionWorkInProgress
	updated.StartedAt = &startedAt
	record := dynamodbrecord.ExecutionWorkItemRecordFromWork(updated)
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		return false, err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: []sharedaws.TransactWriteItem{{Put: &sharedaws.Put{
		TableName:           sharedaws.String(r.tableName),
		Item:                item,
		ConditionExpression: sharedaws.String("#status = :pending"),
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
		},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pending": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkPending)},
		},
	}}}})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceled") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *dynamoRuntimeRepository) GetMonitor(ctx context.Context, tenantID, serviceID, monitorID string) (monitorconfig.Monitor, bool, error) {
	if err := r.requireTableName(); err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	record, found, err := r.getMonitorRecord(ctx, tenantID, serviceID, monitorID)
	if err != nil {
		return monitorconfig.Monitor{}, false, err
	}
	if !found {
		return monitorconfig.Monitor{}, false, nil
	}
	monitor := record.ToMonitor()
	if !strings.EqualFold(monitor.TenantID, tenantID) || !strings.EqualFold(monitor.ServiceID, serviceID) {
		return monitorconfig.Monitor{}, false, nil
	}
	return monitor, true, nil
}

func (r *dynamoRuntimeRepository) getMonitorRecord(ctx context.Context, tenantID, serviceID, monitorID string) (dynamodbrecord.MonitorItemRecord, bool, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.MonitorPK(tenantID, serviceID, monitorID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil {
		return dynamodbrecord.MonitorItemRecord{}, false, err
	}
	if len(out.Item) == 0 {
		return dynamodbrecord.MonitorItemRecord{}, false, nil
	}
	var record dynamodbrecord.MonitorItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return dynamodbrecord.MonitorItemRecord{}, false, err
	}
	return record, true, nil
}

func (r *dynamoRuntimeRepository) GetService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, bool, error) {
	return r.getService(ctx, tenantID, serviceID)
}

func (r *dynamoRuntimeRepository) MarkExecutionWorkSkipped(ctx context.Context, work checkexecution.ExecutionWork, now time.Time, reason string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	completedAt := now.UTC()
	updated := work
	updated.Status = checkexecution.ExecutionWorkSkipped
	updated.CompletedAt = &completedAt
	updated.LastError = reason
	items, err := marshalItems(r.tableName, dynamodbrecord.ExecutionWorkItemRecordFromWork(updated))
	if err != nil {
		return err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	return err
}

func (r *dynamoRuntimeRepository) RecordExecutionResult(ctx context.Context, monitor monitorconfig.Monitor, work checkexecution.ExecutionWork, result checkexecution.ExecutionResult) (string, string, error) {
	if err := r.requireTableName(); err != nil {
		return "", "", err
	}
	completedAt := result.FinishedAt.UTC()
	updatedWork := work
	updatedWork.Status = checkexecution.ExecutionWorkCompleted
	updatedWork.CompletedAt = &completedAt
	updatedWork.LastError = result.Error
	run := resultstatus.NewCheckRun(result, completedAt)

	currentStatus, statusFound, err := r.getMonitorStatus(ctx, result.TenantID, result.ServiceID, result.MonitorID)
	if err != nil {
		return "", "", err
	}
	if !statusFound {
		currentStatus = resultstatus.NewMonitorStatus(result)
	}

	openIncident, incidentFound, err := r.getOpenIncident(ctx, result.TenantID, result.ServiceID, result.MonitorID)
	if err != nil {
		return "", "", err
	}

	thresholdConfig := resultstatus.ThresholdConfig{
		FailureThreshold:  monitor.FailureThreshold,
		RecoveryThreshold: monitor.RecoveryThreshold,
	}
	incidentRecords, transition, incidentID, updatedStatus, err := r.incidentRecordsForResult(monitor, result, currentStatus, thresholdConfig, openIncident, incidentFound)
	if err != nil {
		return "", "", err
	}

	records := []any{dynamodbrecord.ExecutionWorkItemRecordFromWork(updatedWork), run.ToRecord(), updatedStatus.ToRecord()}
	records = append(records, incidentRecords...)
	service, found, err := r.getService(ctx, result.TenantID, result.ServiceID)
	if err != nil {
		return "", "", err
	}
	if found {
		serviceStatus, err := r.buildServiceStatusRecord(ctx, service, updatedStatus)
		if err != nil {
			return "", "", err
		}
		records = append(records, serviceStatus)
	}
	items, err := marshalItems(r.tableName, records...)
	if err != nil {
		return "", "", err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	if err != nil {
		return "", "", err
	}
	return transition, incidentID, nil
}

func (r *dynamoRuntimeRepository) getService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, bool, error) {
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
	return monitorconfig.Service{TenantID: record.TenantID, ServiceID: record.ServiceID, LifecycleState: monitorconfig.ServiceLifecycle(record.LifecycleState), EscalationPolicyID: strings.TrimSpace(record.EscalationPolicyID), BusinessHours: dynamodbrecord.CloneBusinessHoursConfig(record.BusinessHours)}, true, nil
}

func (r *dynamoRuntimeRepository) getMonitorStatus(ctx context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
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
	return resultstatus.MonitorStatus{ServiceID: record.ServiceID, MonitorID: record.MonitorID, TenantID: record.TenantID, CurrentStatus: record.CurrentStatus, ConsecutiveFailures: record.ConsecutiveFailures, ConsecutiveSuccesses: record.ConsecutiveSuccesses, LastCheckedAt: lastCheckedAt, LastDurationMs: record.LastDurationMs, LastError: record.LastError, LastFailureCode: record.LastFailureCode, LastOutcome: checkexecution.Outcome(strings.ToLower(firstNonEmpty(record.LastOutcome, "unknown")))}, true, nil
}

func (r *dynamoRuntimeRepository) GetMonitorStatus(ctx context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	return r.getMonitorStatus(ctx, tenantID, serviceID, monitorID)
}

func (r *dynamoRuntimeRepository) buildServiceStatusRecord(ctx context.Context, service monitorconfig.Service, latestStatus resultstatus.MonitorStatus) (dynamodbrecord.ServiceStatusRecord, error) {
	monitors, err := r.ListMonitors(ctx, service.TenantID)
	if err != nil {
		return dynamodbrecord.ServiceStatusRecord{}, err
	}
	serviceMonitors := make([]monitorconfig.Monitor, 0)
	statusByMonitor := map[string]resultstatus.MonitorStatus{statusKey(latestStatus.ServiceID, latestStatus.MonitorID): latestStatus}
	for _, monitor := range monitors {
		if !strings.EqualFold(monitor.ServiceID, service.ServiceID) {
			continue
		}
		serviceMonitors = append(serviceMonitors, monitor)
		key := statusKey(monitor.ServiceID, monitor.MonitorID)
		if _, ok := statusByMonitor[key]; ok {
			continue
		}
		status, found, err := r.getMonitorStatus(ctx, service.TenantID, monitor.ServiceID, monitor.MonitorID)
		if err != nil {
			return dynamodbrecord.ServiceStatusRecord{}, err
		}
		if found {
			statusByMonitor[key] = status
		}
	}
	rollup := deriveServiceRollup(service.LifecycleState, serviceMonitors, statusByMonitor)
	updatedAt := latestStatus.LastCheckedAt.UTC().Format(time.RFC3339)
	item := dynamodbschema.ServiceStatusItem(service.TenantID, service.ServiceID, rollup, updatedAt)
	return dynamodbrecord.ServiceStatusRecord{PK: item.PK, SK: item.SK, EntityType: item.EntityType, TenantID: strings.ToUpper(service.TenantID), ServiceID: strings.ToLower(service.ServiceID), LifecycleState: string(service.LifecycleState), RollupStatus: rollup, MonitorCount: len(serviceMonitors), EnabledMonitorCount: countEnabledMonitors(serviceMonitors), UpdatedAt: updatedAt, GSI2PK: item.GSI2PK, GSI2SK: item.GSI2SK}, nil
}

func (r *dynamoRuntimeRepository) incidentRecordsForResult(monitor monitorconfig.Monitor, result checkexecution.ExecutionResult, currentStatus resultstatus.MonitorStatus, thresholdConfig resultstatus.ThresholdConfig, current dynamodbrecord.IncidentRecord, found bool) ([]any, string, string, resultstatus.MonitorStatus, error) {
	isManual := result.Trigger == checkexecution.TriggerTypeManual

	currentState := domainvalues.MonitorStateFromStored(currentStatus.CurrentStatus)
	if currentState == "" {
		currentState = domainvalues.MonitorStateUp
	}

	failureThreshold := thresholdConfig.FailureThreshold
	if failureThreshold < 1 {
		failureThreshold = 1
	}
	recoveryThreshold := thresholdConfig.RecoveryThreshold
	if recoveryThreshold < 1 {
		recoveryThreshold = 1
	}

	newStatus := currentStatus
	newStatus.LastCheckedAt = result.FinishedAt.UTC()
	newStatus.LastDurationMs = result.DurationMs
	newStatus.LastError = result.Error
	newStatus.LastFailureCode = result.FailureCode
	newStatus.LastOutcome = result.Outcome

	var incidentRecords []any
	var transition string
	var incidentID string

	if isManual {
		newStatus.CurrentStatus = currentState.Stored()
		return incidentRecords, "", "", newStatus, nil
	}

	if result.Outcome == checkexecution.OutcomeSuccess {
		newStatus.ConsecutiveFailures = 0
		newStatus.ConsecutiveSuccesses++

		switch currentState {
		case domainvalues.MonitorStateUp:
			newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()

		case domainvalues.MonitorStateDegraded:
			newStatus.ConsecutiveFailures = 0
			newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()

		case domainvalues.MonitorStateDown:
			newStatus.CurrentStatus = domainvalues.MonitorStateRecovering.Stored()

		case domainvalues.MonitorStateRecovering:
			if newStatus.ConsecutiveSuccesses >= recoveryThreshold {
				newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()
				newStatus.ConsecutiveSuccesses = 0
				if found {
					current.Status = incidentStatusResolved
					current.ResolvedAt = result.FinishedAt.UTC().Format(time.RFC3339)
					current.UpdatedAt = current.ResolvedAt
					incidentRecords = buildIncidentRecords(current, "INCIDENT_RESOLVED", current.ResolvedAt, result.FinishedAt)
					transition = "incident.up"
					incidentID = current.IncidentID
				}
			} else {
				newStatus.CurrentStatus = domainvalues.MonitorStateRecovering.Stored()
			}

		case domainvalues.MonitorStateMaintenance:
			newStatus.CurrentStatus = domainvalues.MonitorStateMaintenance.Stored()

		default:
			newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()
		}
	} else {
		newStatus.ConsecutiveSuccesses = 0
		newStatus.ConsecutiveFailures++

		switch currentState {
		case domainvalues.MonitorStateUp:
			if failureThreshold == 1 {
				newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			} else {
				newStatus.CurrentStatus = domainvalues.MonitorStateDegraded.Stored()
			}

		case domainvalues.MonitorStateDegraded:
			if newStatus.ConsecutiveFailures >= failureThreshold {
				newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			} else {
				newStatus.CurrentStatus = domainvalues.MonitorStateDegraded.Stored()
			}

		case domainvalues.MonitorStateDown:
			newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			if found {
				current.Summary = incidentSummary(monitor, result)
				current.UpdatedAt = result.FinishedAt.UTC().Format(time.RFC3339)
				incidentRecords = buildIncidentRecords(current, "INCIDENT_UPDATED", current.UpdatedAt, result.FinishedAt)
			}

		case domainvalues.MonitorStateRecovering:
			newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			newStatus.ConsecutiveSuccesses = 0

		case domainvalues.MonitorStateMaintenance:
			newStatus.CurrentStatus = domainvalues.MonitorStateMaintenance.Stored()

		default:
			newStatus.CurrentStatus = domainvalues.MonitorStateDegraded.Stored()
		}
	}

	if newStatus.CurrentStatus == domainvalues.MonitorStateDown.Stored() && currentState != domainvalues.MonitorStateDown {
		if !found {
			summary := incidentSummary(monitor, result)
			incident := dynamodbrecord.IncidentRecord{
				IncidentID: newIncidentID(result.FinishedAt),
				ServiceID:  strings.ToLower(result.ServiceID),
				MonitorID:  strings.ToLower(result.MonitorID),
				TenantID:   strings.ToUpper(result.TenantID),
				Type:       "monitoring",
				Summary:    summary,
				Status:     incidentStatusOpen,
				OpenedAt:   result.FinishedAt.UTC().Format(time.RFC3339),
				UpdatedAt:  result.FinishedAt.UTC().Format(time.RFC3339),
				Origin:     "system",
			}
			incidentRecords = buildIncidentRecords(incident, "INCIDENT_OPENED", incident.IncidentID, result.FinishedAt)
			transition = "incident.down"
			incidentID = incident.IncidentID
		}
	}

	return incidentRecords, transition, incidentID, newStatus, nil
}

func (r *dynamoRuntimeRepository) getOpenIncident(ctx context.Context, tenantID, serviceID, monitorID string) (dynamodbrecord.IncidentRecord, bool, error) {
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.MonitorPK(tenantID, serviceID, monitorID)},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "INCIDENT#"},
		},
	})
	if err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.IncidentItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return dynamodbrecord.IncidentRecord{}, false, err
		}
		if record.EntityType != dynamodbschema.EntityIncident || record.TenantID != strings.ToUpper(tenantID) {
			continue
		}
		incident := record.ToIncident()
		if incident.Status == incidentStatusOpen || incident.Status == incidentStatusAcknowledged {
			incidents = append(incidents, incident)
		}
	}
	if len(incidents) == 0 {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].OpenedAt > incidents[j].OpenedAt })
	return incidents[0], true, nil
}

func (r *dynamoRuntimeRepository) requireTableName() error {
	if strings.TrimSpace(r.tableName) == "" {
		return fmt.Errorf("TABLE_NAME is required")
	}
	return nil
}

const (
	incidentStatusOpen         = "open"
	incidentStatusAcknowledged = "acknowledged"
	incidentStatusResolved     = "resolved"
)

func buildIncidentRecords(incident dynamodbrecord.IncidentRecord, action, changeValue string, now time.Time) []any {
	auditID := newAuditID(now)
	activityID := newActivityID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, incident.TenantID, action, incident.ServiceID, incident.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "incident", "", changeValue)
	activity := dynamodbrecord.NewIncidentActivityRecord(incident.TenantID, incident.IncidentID, activityID, action, now)
	return []any{dynamodbrecord.NewIncidentMonitorItemRecord(incident), dynamodbrecord.NewIncidentRefItemRecord(incident), dynamodbrecord.NewIncidentMetaItemRecord(incident), activity, auditEvent, change}
}

func incidentSummary(monitor monitorconfig.Monitor, result checkexecution.ExecutionResult) string {
	summary := fmt.Sprintf("%s failed", monitor.Name)
	if result.Outcome == checkexecution.OutcomeSuccess {
		return summary
	}
	if result.Error != "" {
		return fmt.Sprintf("%s: %s", summary, result.Error)
	}
	if result.StatusCode != nil {
		return fmt.Sprintf("%s: status %d", summary, *result.StatusCode)
	}
	return summary
}

func marshalItems(tableName string, records ...any) ([]sharedaws.TransactWriteItem, error) {
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

func deriveServiceRollup(lifecycle monitorconfig.ServiceLifecycle, monitors []monitorconfig.Monitor, statuses map[string]resultstatus.MonitorStatus) string {
	switch lifecycle {
	case monitorconfig.ServiceLifecycleDraft:
		return "draft"
	case monitorconfig.ServiceLifecycleArchived:
		return "archived"
	}
	enabled := make([]monitorconfig.Monitor, 0)
	for _, monitor := range monitors {
		if monitor.Enabled {
			enabled = append(enabled, monitor)
		}
	}
	if len(enabled) == 0 {
		return "paused"
	}
	upCount := 0
	downCount := 0
	unknownCount := 0
	for _, monitor := range enabled {
		status, ok := statuses[statusKey(monitor.ServiceID, monitor.MonitorID)]
		if !ok {
			unknownCount++
			continue
		}
		switch strings.ToLower(status.CurrentStatus) {
		case "up":
			upCount++
		case "down":
			downCount++
		default:
			unknownCount++
		}
	}
	if unknownCount == len(enabled) {
		return "unknown"
	}
	if upCount == len(enabled) {
		return "up"
	}
	if downCount == len(enabled) {
		return "down"
	}
	return "degraded"
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

func statusKey(serviceID, monitorID string) string {
	return strings.ToLower(serviceID) + "/" + strings.ToLower(monitorID)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
