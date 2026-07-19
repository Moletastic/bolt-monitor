package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
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
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), "SCHEDULER_CONFIG"))
	if err != nil {
		return checkexecution.SchedulerConfig{}, err
	}
	if !found {
		return checkexecution.SchedulerConfig{}, nil
	}
	var record dynamodbrecord.SchedulerConfigItemRecord
	if err := sharedaws.UnmarshalMap(item, &record); err != nil {
		return checkexecution.SchedulerConfig{}, err
	}
	return checkexecution.SchedulerConfig{RecurringEnabled: record.RecurringEnabled, StopControlMode: checkexecution.StopControlMode(record.StopControlMode)}, nil
}

func (r *dynamoRuntimeRepository) ListMonitors(ctx context.Context, tenantID string) ([]monitorconfig.Monitor, error) {
	return r.listMonitorsBounded(ctx, tenantID, sharedaws.PageOptions{Limit: 100})
}

func (r *dynamoRuntimeRepository) listMonitorsBounded(ctx context.Context, tenantID string, pageOpts sharedaws.PageOptions) ([]monitorconfig.Monitor, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	if pageOpts.Limit <= 0 {
		pageOpts.Limit = 100
	}
	monitors := make([]monitorconfig.Monitor, 0)
	seen := map[string]struct{}{}
	var serviceCursor map[string]sharedaws.AttributeValue
	for {
		pageOpts.Cursor = serviceCursor
		page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.TenantPK(tenantID), "SERVICE#", pageOpts)
		if err != nil {
			return nil, err
		}
		for _, item := range page.Items {
			var service dynamodbrecord.ServiceItemRecord
			if err := sharedaws.UnmarshalMap(item, &service); err != nil {
				return nil, err
			}
			if service.EntityType != dynamodbschema.EntityServiceRef {
				continue
			}
			var monitorCursor map[string]sharedaws.AttributeValue
			for {
				pageOpts.Cursor = monitorCursor
				serviceMonitors, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.ServicePK(tenantID, service.ServiceID), "MONITOR#", pageOpts)
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
					monitor := record.ToMonitor()
					key := monitor.TenantID + "/" + monitor.ServiceID + "/" + monitor.MonitorID
					if _, dup := seen[key]; dup {
						continue
					}
					seen[key] = struct{}{}
					monitors = append(monitors, monitor)
				}
				if !serviceMonitors.HasMore() {
					break
				}
				monitorCursor = serviceMonitors.NextKey
			}
		}
		if !page.HasMore() {
			break
		}
		serviceCursor = page.NextKey
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
	for _, request := range requests {
		runID := request.RunID
		if strings.TrimSpace(runID) == "" {
			runID = newRunID(now)
		}
		acceptedAt := request.AcceptedAt
		if acceptedAt.IsZero() {
			acceptedAt = now.UTC()
		}
		work := checkexecution.ExecutionWork{
			TenantID: request.Monitor.TenantID, ServiceID: request.Monitor.ServiceID, MonitorID: request.Monitor.MonitorID,
			RunID: runID, Trigger: request.Trigger, AcceptedAt: acceptedAt, RequestedAt: acceptedAt,
			ScheduleDefinitionVersion: request.ScheduleDefinitionVersion, ScheduledFor: request.ScheduledFor,
			Status: checkexecution.ExecutionWorkPending, PublicationState: checkexecution.PublicationPending,
		}
		if err := r.createExecutionWork(ctx, work); err != nil {
			return err
		}
	}
	return nil
}

func (r *dynamoRuntimeRepository) createExecutionWork(ctx context.Context, work checkexecution.ExecutionWork) error {
	bucket, shard := executionRecoveryBucket(work)
	records := []any{
		dynamodbrecord.ExecutionWorkItemRecordFromWork(work),
		dynamodbrecord.NewExecutionMarkerRecord(work, dynamodbrecord.ExecutionMarkerPublication, bucket, shard),
		dynamodbrecord.NewExecutionMarkerRecord(work, dynamodbrecord.ExecutionMarkerRecovery, bucket, shard),
	}
	items, err := conditionalPutItems(r.tableName, records...)
	if err != nil {
		return err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	if err == nil {
		return nil
	}
	if !sharedaws.IsConditionalCheckFailure(err) {
		return err
	}
	existing, found, loadErr := r.loadExecutionWork(ctx, work.TenantID, work.RunID)
	if loadErr != nil {
		return loadErr
	}
	if !found || !sameWorkIdentity(existing, work) {
		return checkexecution.Conflict("create-work", work.RunID)
	}
	return nil
}

func conditionalPutItems(tableName string, records ...any) ([]sharedaws.TransactWriteItem, error) {
	items := make([]sharedaws.TransactWriteItem, 0, len(records))
	for _, record := range records {
		item, err := sharedaws.MarshalMap(record)
		if err != nil {
			return nil, err
		}
		items = append(items, sharedaws.TransactWriteItem{Put: &sharedaws.Put{
			TableName: sharedaws.String(tableName), Item: item,
			ConditionExpression: sharedaws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}})
	}
	return items, nil
}

func executionRecoveryBucket(work checkexecution.ExecutionWork) (string, string) {
	acceptedAt := work.AcceptedAt
	if acceptedAt.IsZero() {
		acceptedAt = work.RequestedAt
	}
	sum := sha256.Sum256([]byte(work.RunID))
	return acceptedAt.UTC().Format("2006010215"), fmt.Sprintf("%02x", sum[0]%16)
}

func (r *dynamoRuntimeRepository) listExecutionMarkers(ctx context.Context, tenantID, kind, bucket, shard string, limit int32, cursor map[string]sharedaws.AttributeValue) ([]dynamodbrecord.ExecutionMarkerRecord, map[string]sharedaws.AttributeValue, error) {
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, fmt.Sprintf("RECOVERY#%s#%s#%s#%s", dynamodbschema.NormalizeToken(tenantID), dynamodbschema.NormalizeToken(kind), dynamodbschema.NormalizeToken(bucket), dynamodbschema.NormalizeToken(shard)), "MARKER#", sharedaws.PageOptions{Limit: limit, Cursor: cursor})
	if err != nil {
		return nil, nil, err
	}
	markers := make([]dynamodbrecord.ExecutionMarkerRecord, 0, len(page.Items))
	for _, item := range page.Items {
		var marker dynamodbrecord.ExecutionMarkerRecord
		if err := sharedaws.UnmarshalMap(item, &marker); err != nil {
			return nil, nil, err
		}
		if marker.EntityType == dynamodbschema.EntityExecutionMarker {
			markers = append(markers, marker)
		}
	}
	return markers, page.NextKey, nil
}

func (r *dynamoRuntimeRepository) ListPublicationMarkers(ctx context.Context, tenantID, bucketShard string, limit int32, cursor map[string]sharedaws.AttributeValue) ([]dynamodbrecord.ExecutionMarkerRecord, map[string]sharedaws.AttributeValue, error) {
	bucketShard = strings.TrimSpace(bucketShard)
	parts := strings.SplitN(bucketShard, "|", 2)
	bucket := ""
	shard := ""
	if len(parts) > 0 {
		bucket = parts[0]
	}
	if len(parts) > 1 {
		shard = parts[1]
	}
	return r.listExecutionMarkers(ctx, tenantID, dynamodbrecord.ExecutionMarkerPublication, bucket, shard, limit, cursor)
}

func (r *dynamoRuntimeRepository) ListDispatchPending(ctx context.Context, tenantID, bucketShard string, limit int32, cursor map[string]sharedaws.AttributeValue) ([]dynamodbrecord.DispatchPendingRecord, map[string]sharedaws.AttributeValue, error) {
	if err := r.requireTableName(); err != nil {
		return nil, nil, err
	}
	bucketShard = strings.TrimSpace(bucketShard)
	parts := strings.SplitN(bucketShard, "|", 2)
	bucket := ""
	shard := ""
	if len(parts) > 0 {
		bucket = parts[0]
	}
	if len(parts) > 1 {
		shard = parts[1]
	}
	pk := fmt.Sprintf("DISPATCH_PENDING#%s#%s#%s", dynamodbschema.NormalizeToken(tenantID), dynamodbschema.NormalizeToken(bucket), dynamodbschema.NormalizeToken(shard))
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, pk, "", sharedaws.PageOptions{Limit: limit, Cursor: cursor})
	if err != nil {
		return nil, nil, err
	}
	records := make([]dynamodbrecord.DispatchPendingRecord, 0, len(page.Items))
	for _, item := range page.Items {
		var record dynamodbrecord.DispatchPendingRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, nil, err
		}
		if record.EntityType == dynamodbschema.EntityDispatchPending {
			records = append(records, record)
		}
	}
	return records, page.NextKey, nil
}

func (r *dynamoRuntimeRepository) RemoveDispatchPending(ctx context.Context, tenantID, bucket, shard, eventID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	pk := fmt.Sprintf("DISPATCH_PENDING#%s#%s#%s", dynamodbschema.NormalizeToken(tenantID), dynamodbschema.NormalizeToken(bucket), dynamodbschema.NormalizeToken(shard))
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{
		TableName: sharedaws.String(r.tableName),
		Key:      sharedaws.NewPrimaryKey(pk, dynamodbschema.NormalizeToken(eventID)).AttributeMap(),
	})
	return err
}

func (r *dynamoRuntimeRepository) LoadTransitionOutbox(ctx context.Context, tenantID, eventID string) (dynamodbrecord.TransitionOutboxRecord, bool, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.TransitionOutboxRecord{}, false, err
	}
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), "TRANSITION_OUTBOX#"+dynamodbschema.NormalizeToken(eventID)))
	if err != nil || !found {
		return dynamodbrecord.TransitionOutboxRecord{}, found, err
	}
	var record dynamodbrecord.TransitionOutboxRecord
	if err := sharedaws.UnmarshalMap(item, &record); err != nil {
		return dynamodbrecord.TransitionOutboxRecord{}, false, err
	}
	return record, true, nil
}

func (r *dynamoRuntimeRepository) AcknowledgeExecutionPublication(ctx context.Context, work checkexecution.ExecutionWork) error {
	bucket, shard := executionRecoveryBucket(work)
	marker := dynamodbrecord.NewExecutionMarkerRecord(work, dynamodbrecord.ExecutionMarkerPublication, bucket, shard)
	markerKey := sharedaws.NewPrimaryKey(marker.PK, marker.SK).AttributeMap()
	workKey := sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(work.TenantID), "RUN_REQUEST#"+dynamodbschema.NormalizeToken(work.RunID)).AttributeMap()
	_, err := r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: []sharedaws.TransactWriteItem{
		{Update: &sharedaws.Update{
			TableName: sharedaws.String(r.tableName), Key: workKey,
			UpdateExpression: sharedaws.String("SET PublicationState = :acknowledged"),
			ConditionExpression: sharedaws.String("PublicationState = :pending"),
			ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
				":acknowledged": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.PublicationAcknowledged)},
				":pending": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.PublicationPending)},
			},
		}},
		{Delete: &sharedaws.Delete{TableName: sharedaws.String(r.tableName), Key: markerKey}},
	}})
	if err != nil {
		return err
	}
	return nil
}

func (r *dynamoRuntimeRepository) loadExecutionWork(ctx context.Context, tenantID, runID string) (checkexecution.ExecutionWork, bool, error) {
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), "RUN_REQUEST#"+dynamodbschema.NormalizeToken(runID)))
	if err != nil || !found {
		return checkexecution.ExecutionWork{}, found, err
	}
	var record dynamodbrecord.ExecutionWorkItemRecord
	if err := sharedaws.UnmarshalMap(item, &record); err != nil {
		return checkexecution.ExecutionWork{}, false, err
	}
	work, err := record.ToWork()
	return work, err == nil, err
}

func (r *dynamoRuntimeRepository) LoadExecutionWork(ctx context.Context, tenantID, runID string) (checkexecution.ExecutionWork, bool, error) {
	return r.loadExecutionWork(ctx, tenantID, runID)
}

func sameWorkIdentity(left, right checkexecution.ExecutionWork) bool {
	if !strings.EqualFold(left.TenantID, right.TenantID) || !strings.EqualFold(left.ServiceID, right.ServiceID) || !strings.EqualFold(left.MonitorID, right.MonitorID) || !strings.EqualFold(left.RunID, right.RunID) || left.Trigger != right.Trigger || left.ScheduleDefinitionVersion != right.ScheduleDefinitionVersion {
		return false
	}
	if left.ScheduledFor == nil || right.ScheduledFor == nil {
		return left.ScheduledFor == nil && right.ScheduledFor == nil
	}
	return left.ScheduledFor.Equal(*right.ScheduledFor)
}

func (r *dynamoRuntimeRepository) ListPendingExecutionWork(ctx context.Context, tenantID string, limit int32) ([]checkexecution.ExecutionWork, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	// Bounded worker claim: explicit limit prevents unbounded cost. Any
	// continuation state beyond the first page is reported as incomplete so
	// the worker can retry the next SQS message.
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.TenantPK(tenantID), "RUN_REQUEST#", sharedaws.PageOptions{
		Limit:   limit,
		Forward: true,
	})
	if err != nil {
		return nil, err
	}
	works := make([]checkexecution.ExecutionWork, 0, len(page.Items))
	for _, item := range page.Items {
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

const defaultExecutionWorkLeaseDuration = 60 * time.Second

func executionWorkLeaseDuration() time.Duration {
	seconds, err := strconv.Atoi(os.Getenv("WORK_LEASE_DURATION_SECONDS"))
	if err != nil || seconds <= 0 {
		return defaultExecutionWorkLeaseDuration
	}
	return time.Duration(seconds) * time.Second
}

func (r *dynamoRuntimeRepository) ClaimExecutionWork(ctx context.Context, work checkexecution.ExecutionWork, now time.Time) (checkexecution.ExecutionWork, bool, error) {
	if err := r.requireTableName(); err != nil {
		return checkexecution.ExecutionWork{}, false, err
	}
	startedAt := now.UTC()
	leaseUntil := startedAt.Add(executionWorkLeaseDuration())
	token := newFencingToken(now)
	key := sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(work.TenantID), "RUN_REQUEST#"+dynamodbschema.NormalizeToken(work.RunID)).AttributeMap()
	out, err := r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
		TableName: sharedaws.String(r.tableName), Key: key,
		UpdateExpression: sharedaws.String("SET #status = :inProgress, StartedAt = :startedAt, LeaseUntil = :leaseUntil, FencingToken = :token, AttemptCount = if_not_exists(AttemptCount, :zero) + :one"),
		ConditionExpression: sharedaws.String("#status = :pending OR (#status = :inProgress AND LeaseUntil < :now)"),
		ExpressionAttributeNames: map[string]string{"#status": "Status"},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pending": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkPending)},
			":inProgress": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkInProgress)},
			":startedAt": &sharedaws.AttributeValueMemberS{Value: startedAt.Format(time.RFC3339)},
			":leaseUntil": &sharedaws.AttributeValueMemberS{Value: leaseUntil.Format(time.RFC3339)},
			":token": &sharedaws.AttributeValueMemberS{Value: token},
			":now": &sharedaws.AttributeValueMemberS{Value: startedAt.Format(time.RFC3339)},
			":zero": &sharedaws.AttributeValueMemberN{Value: "0"},
			":one": &sharedaws.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: "ALL_NEW",
	})
	if err != nil {
		if sharedaws.IsConditionalCheckFailure(err) {
			return work, false, nil
		}
		return checkexecution.ExecutionWork{}, false, err
	}
	var record dynamodbrecord.ExecutionWorkItemRecord
	if err := sharedaws.UnmarshalMap(out.Attributes, &record); err != nil {
		return checkexecution.ExecutionWork{}, false, err
	}
	claimed, err := record.ToWork()
	return claimed, err == nil, err
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
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.MonitorPK(tenantID, serviceID, monitorID), "META"))
	if err != nil {
		return dynamodbrecord.MonitorItemRecord{}, false, err
	}
	if !found {
		return dynamodbrecord.MonitorItemRecord{}, false, nil
	}
	var record dynamodbrecord.MonitorItemRecord
	if err := sharedaws.UnmarshalMap(item, &record); err != nil {
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
	bucket, shard := executionRecoveryBucket(work)
	recoveryMarker := dynamodbrecord.NewExecutionMarkerRecord(work, dynamodbrecord.ExecutionMarkerRecovery, bucket, shard)
	_, err := r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: []sharedaws.TransactWriteItem{
		{Update: &sharedaws.Update{
			TableName: sharedaws.String(r.tableName),
			Key: sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(work.TenantID), "RUN_REQUEST#"+dynamodbschema.NormalizeToken(work.RunID)).AttributeMap(),
			UpdateExpression: sharedaws.String("SET #status = :skipped, CompletedAt = :completedAt, TerminalReason = :reason"),
			ConditionExpression: sharedaws.String("#status = :inProgress AND FencingToken = :token"),
			ExpressionAttributeNames: map[string]string{"#status": "Status"},
			ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
				":skipped": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkSkipped)},
				":inProgress": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkInProgress)},
				":completedAt": &sharedaws.AttributeValueMemberS{Value: completedAt.Format(time.RFC3339)},
				":reason": &sharedaws.AttributeValueMemberS{Value: reason},
				":token": &sharedaws.AttributeValueMemberS{Value: work.FencingToken},
			},
		}},
		{Delete: &sharedaws.Delete{
			TableName: sharedaws.String(r.tableName),
			Key:       sharedaws.NewPrimaryKey(recoveryMarker.PK, recoveryMarker.SK).AttributeMap(),
		}},
	}})
	if sharedaws.IsConditionalCheckFailure(err) {
		return checkexecution.LeaseLost("skip-work", work.RunID)
	}
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

	applyProjection := result.Trigger == checkexecution.TriggerTypeRecurring && result.ScheduledFor != nil && resultstatus.IsNewerRecurringObservation(currentStatus, *result.ScheduledFor, result.RunID)
	records := []any{run.ToRecord()}
	if applyProjection {
		updatedStatus.RecurringScheduledFor = result.ScheduledFor
		updatedStatus.RecurringRunID = result.RunID
		records = append(records, updatedStatus.ToRecord())
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
		if transition != "" {
			transitionType := transition
			eventID := checkexecution.TransitionID(work.RunID)
			transition = eventID
			outbox := dynamodbrecord.NewTransitionOutboxRecord(result.TenantID, result.ServiceID, result.MonitorID, eventID, work.RunID, incidentID, transitionType, work.ScheduleDefinitionVersion, formatScheduledFor(result.ScheduledFor), completedAt.Format(time.RFC3339))
			records = append(records, outbox)
			bucket, shard := executionRecoveryBucket(work)
			pending := dynamodbrecord.NewDispatchPendingRecord(result.TenantID, eventID, bucket, shard)
			records = append(records, pending)
		}
	} else {
		transition = ""
		incidentID = ""
	}
	items, err := marshalItems(r.tableName, records...)
	if err != nil {
		return "", "", err
	}
	identityItem, err := sharedaws.MarshalMap(run.IdentityRecord())
	if err != nil {
		return "", "", err
	}
	runIdentity := sharedaws.TransactWriteItem{Put: &sharedaws.Put{
		TableName: sharedaws.String(r.tableName), Item: identityItem,
		ConditionExpression: sharedaws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}}
	workKey := sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(work.TenantID), "RUN_REQUEST#"+dynamodbschema.NormalizeToken(work.RunID)).AttributeMap()
	completion := sharedaws.TransactWriteItem{Update: &sharedaws.Update{
		TableName: sharedaws.String(r.tableName), Key: workKey,
		UpdateExpression: sharedaws.String("SET #status = :completed, CompletedAt = :completedAt, LastError = :lastError"),
		ConditionExpression: sharedaws.String("#status = :inProgress AND FencingToken = :token"),
		ExpressionAttributeNames: map[string]string{"#status": "Status"},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":completed": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkCompleted)},
			":inProgress": &sharedaws.AttributeValueMemberS{Value: string(checkexecution.ExecutionWorkInProgress)},
			":completedAt": &sharedaws.AttributeValueMemberS{Value: completedAt.Format(time.RFC3339)},
			":lastError": &sharedaws.AttributeValueMemberS{Value: result.Error},
			":token": &sharedaws.AttributeValueMemberS{Value: work.FencingToken},
		},
	}}
	bucket, shard := executionRecoveryBucket(work)
	recoveryMarker := dynamodbrecord.NewExecutionMarkerRecord(work, dynamodbrecord.ExecutionMarkerRecovery, bucket, shard)
	markerDelete := sharedaws.TransactWriteItem{Delete: &sharedaws.Delete{
		TableName: sharedaws.String(r.tableName),
		Key: sharedaws.NewPrimaryKey(recoveryMarker.PK, recoveryMarker.SK).AttributeMap(),
	}}
	items = append([]sharedaws.TransactWriteItem{completion, runIdentity}, items...)
	items = append(items, markerDelete)
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	if err != nil {
		cancellations := sharedaws.TransactionCancellations(err)
		if len(cancellations) > 1 && cancellations[1].Code == "ConditionalCheckFailed" {
			return "", "", checkexecution.Duplicate("commit-result", work.RunID)
		}
		if len(cancellations) > 0 {
			return "", "", checkexecution.LeaseLost("complete-work", work.RunID)
		}
		return "", "", err
	}
	return transition, incidentID, nil
}

func (r *dynamoRuntimeRepository) getService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, bool, error) {
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.ServicePK(tenantID, serviceID), "META"))
	if err != nil {
		return monitorconfig.Service{}, false, err
	}
	if !found {
		return monitorconfig.Service{}, false, nil
	}
	var record dynamodbrecord.ServiceItemRecord
	if err := sharedaws.UnmarshalMap(item, &record); err != nil {
		return monitorconfig.Service{}, false, err
	}
	return monitorconfig.Service{TenantID: record.TenantID, ServiceID: record.ServiceID, LifecycleState: monitorconfig.ServiceLifecycle(record.LifecycleState), EscalationPolicyID: strings.TrimSpace(record.EscalationPolicyID), BusinessHours: dynamodbrecord.CloneBusinessHoursConfig(record.BusinessHours)}, true, nil
}

func (r *dynamoRuntimeRepository) getMonitorStatus(ctx context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.MonitorPK(tenantID, serviceID, monitorID), "STATUS"))
	if err != nil {
		return resultstatus.MonitorStatus{}, false, err
	}
	if !found {
		return resultstatus.MonitorStatus{}, false, nil
	}
	var record resultstatus.MonitorStatusRecord
	if err := sharedaws.UnmarshalMap(item, &record); err != nil {
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
					incidentRecords = buildIncidentRecords(current, "INCIDENT_RESOLVED", current.ResolvedAt, result.RunID, result.FinishedAt)
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
				incidentRecords = buildIncidentRecords(current, "INCIDENT_UPDATED", current.UpdatedAt, result.RunID, result.FinishedAt)
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
			incidentRecords = buildIncidentRecords(incident, "INCIDENT_OPENED", incident.IncidentID, result.RunID, result.FinishedAt)
			transition = "incident.down"
			incidentID = incident.IncidentID
		}
	}

	return incidentRecords, transition, incidentID, newStatus, nil
}

func (r *dynamoRuntimeRepository) getOpenIncident(ctx context.Context, tenantID, serviceID, monitorID string) (dynamodbrecord.IncidentRecord, bool, error) {
	// Bounded by a single page: the worker only needs the most recent open
	// incident for the monitor. If more than one page exists, the older
	// entries remain for later resolution but do not block the current tick.
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, dynamodbschema.MonitorPK(tenantID, serviceID, monitorID), "INCIDENT#", sharedaws.PageOptions{
		Limit:   20,
		Forward: false,
	})
	if err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(page.Items))
	for _, item := range page.Items {
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

func buildIncidentRecords(incident dynamodbrecord.IncidentRecord, action, changeValue string, runID string, now time.Time) []any {
	auditID := newAuditID(now)
	activityID := checkexecution.TransitionID(runID)
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

func formatScheduledFor(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
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
