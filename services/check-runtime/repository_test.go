package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type fakeDynamoClient struct {
	transactInput    *sharedaws.DynamoDBTransactWriteItemsInput
	queryOutput      *sharedaws.DynamoDBQueryOutput
	getItemOutput    *sharedaws.DynamoDBGetItemOutput
	transactErr      error
	updateErr        error
	updateOutputs    []*sharedaws.DynamoDBUpdateItemOutput
	store            map[string]map[string]sharedaws.AttributeValue
	updateItemInputs []*sharedaws.DynamoDBUpdateItemInput
}

func (f *fakeDynamoClient) GetItem(_ context.Context, input *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	if f.getItemOutput != nil {
		return f.getItemOutput, nil
	}
	if f.store != nil {
		pk, _ := input.Key["PK"].(*sharedaws.AttributeValueMemberS)
		sk, _ := input.Key["SK"].(*sharedaws.AttributeValueMemberS)
		if row, ok := f.store[pk.Value+"/"+sk.Value]; ok {
			return &sharedaws.DynamoDBGetItemOutput{Item: row}, nil
		}
	}
	return &sharedaws.DynamoDBGetItemOutput{}, nil
}

func (f *fakeDynamoClient) Query(context.Context, *sharedaws.DynamoDBQueryInput) (*sharedaws.DynamoDBQueryOutput, error) {
	if f.queryOutput != nil {
		return f.queryOutput, nil
	}
	return &sharedaws.DynamoDBQueryOutput{}, nil
}

func (f *fakeDynamoClient) TransactWriteItems(_ context.Context, input *sharedaws.DynamoDBTransactWriteItemsInput) (*sharedaws.DynamoDBTransactWriteItemsOutput, error) {
	f.transactInput = input
	if f.transactErr != nil {
		return nil, f.transactErr
	}
	return &sharedaws.DynamoDBTransactWriteItemsOutput{}, nil
}

func (f *fakeDynamoClient) PutItem(context.Context, *sharedaws.DynamoDBPutItemInput) (*sharedaws.DynamoDBPutItemOutput, error) {
	return &sharedaws.DynamoDBPutItemOutput{}, nil
}

func TestEnqueueExecutionRequestsConditionallyCreatesWork(t *testing.T) {
	client := &fakeDynamoClient{}
	repo := newDynamoRuntimeRepository(client, "table-name")
	acceptedAt := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	request := checkexecution.ExecutionRequest{Monitor: testMonitor("https://example.com", true), RunID: "RUN_1", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: acceptedAt}
	if _, err := repo.EnqueueExecutionRequests(context.Background(), []checkexecution.ExecutionRequest{request}, acceptedAt); err != nil {
		t.Fatalf("EnqueueExecutionRequests returned error: %v", err)
	}
	if client.transactInput == nil || len(client.transactInput.TransactItems) != 3 {
		t.Fatalf("transaction = %#v", client.transactInput)
	}
	for _, item := range client.transactInput.TransactItems {
		if sharedaws.ToString(item.Put.ConditionExpression) != "attribute_not_exists(PK) AND attribute_not_exists(SK)" {
			t.Fatalf("condition = %v", item.Put.ConditionExpression)
		}
	}
}

func (f *fakeDynamoClient) UpdateItem(_ context.Context, input *sharedaws.DynamoDBUpdateItemInput) (*sharedaws.DynamoDBUpdateItemOutput, error) {
	f.updateItemInputs = append(f.updateItemInputs, input)
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if len(f.updateOutputs) > 0 {
		out := f.updateOutputs[0]
		if len(f.updateOutputs) > 1 {
			f.updateOutputs = f.updateOutputs[1:]
		}
		return out, nil
	}
	return &sharedaws.DynamoDBUpdateItemOutput{}, nil
}

func (f *fakeDynamoClient) DeleteItem(context.Context, *sharedaws.DynamoDBDeleteItemInput) (*sharedaws.DynamoDBDeleteItemOutput, error) {
	return &sharedaws.DynamoDBDeleteItemOutput{}, nil
}

func (f *fakeDynamoClient) Scan(context.Context, *sharedaws.DynamoDBScanInput) (*sharedaws.DynamoDBScanOutput, error) {
	return &sharedaws.DynamoDBScanOutput{}, nil
}

func TestEnqueueExecutionRequestsIsNoOpWhenWorkAlreadyExists(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	existing := dynamodbrecord.ExecutionWorkItemRecordFromWork(checkexecution.ExecutionWork{
		TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now,
	})
	item, err := sharedaws.MarshalMap(existing)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	store := map[string]map[string]sharedaws.AttributeValue{
		sharedaws.NewPrimaryKey("TENANT#DEFAULT", "RUN_REQUEST#RUN_1").PK + "/" + sharedaws.NewPrimaryKey("TENANT#DEFAULT", "RUN_REQUEST#RUN_1").SK: item,
	}
	client := &fakeDynamoClient{store: store, transactErr: sharedaws.NewConditionalCheckFailedError()}
	repo := newDynamoRuntimeRepository(client, "table-name")
	request := checkexecution.ExecutionRequest{Monitor: testMonitor("https://example.com", true), RunID: "RUN_1", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now}
	if _, err := repo.EnqueueExecutionRequests(context.Background(), []checkexecution.ExecutionRequest{request}, now); err != nil {
		t.Fatalf("EnqueueExecutionRequests returned error: %v", err)
	}
}

func TestEnqueueExecutionRequestsRejectsIdentityConflict(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	conflicting := dynamodbrecord.ExecutionWorkItemRecordFromWork(checkexecution.ExecutionWork{
		TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now,
	})
	item, err := sharedaws.MarshalMap(conflicting)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	store := map[string]map[string]sharedaws.AttributeValue{
		sharedaws.NewPrimaryKey("TENANT#DEFAULT", "RUN_REQUEST#RUN_1").PK + "/" + sharedaws.NewPrimaryKey("TENANT#DEFAULT", "RUN_REQUEST#RUN_1").SK: item,
	}
	client := &fakeDynamoClient{store: store, transactErr: sharedaws.NewConditionalCheckFailedError()}
	repo := newDynamoRuntimeRepository(client, "table-name")
	request := checkexecution.ExecutionRequest{Monitor: testMonitor("https://example.com", true), RunID: "RUN_1", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now}
	if _, err := repo.EnqueueExecutionRequests(context.Background(), []checkexecution.ExecutionRequest{request}, now); err == nil || !strings.Contains(err.Error(), "immutable_identity_conflict") {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestClaimExecutionWorkReportsActiveLeaseConflict(t *testing.T) {
	client := &fakeDynamoClient{updateErr: sharedaws.NewConditionalCheckFailedError()}
	repo := newDynamoRuntimeRepository(client, "table-name")
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now}
	if _, claimed, err := repo.ClaimExecutionWork(context.Background(), work, now); err != nil || claimed {
		t.Fatalf("ClaimExecutionWork = (%v, %v)", err, claimed)
	}
}

func TestMarkExecutionWorkSkippedDetectsLeaseLoss(t *testing.T) {
	client := &fakeDynamoClient{transactErr: sharedaws.NewConditionalCheckFailedError()}
	repo := newDynamoRuntimeRepository(client, "table-name")
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, RunID: "RUN_1", AcceptedAt: now, FencingToken: "TOKEN"}
	err := repo.MarkExecutionWorkSkipped(context.Background(), work, now, "monitor disabled")
	if err == nil || !strings.Contains(err.Error(), "lease_lost") {
		t.Fatalf("expected lease lost, got %v", err)
	}
}

func TestRecordExecutionResultDetectsDuplicate(t *testing.T) {
	client := &fakeDynamoClient{transactErr: sharedaws.NewTransactionCanceledException([]string{"None", "ConditionalCheckFailed", "None"})}
	repo := newDynamoRuntimeRepository(client, "table-name")
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	monitor := testMonitor("https://example.com", true)
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now, FencingToken: "TOKEN"}
	result := checkexecution.ExecutionResult{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Outcome: checkexecution.OutcomeSuccess, FinishedAt: now}
	if _, _, err := repo.RecordExecutionResult(context.Background(), monitor, work, result); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate, got %v", err)
	}
}

func TestExecutionSafetyConfigAcceptsValidSettings(t *testing.T) {
	t.Setenv("WORKER_LAMBDA_TIMEOUT_SECONDS", "45")
	t.Setenv("EXECUTION_QUEUE_VISIBILITY_TIMEOUT_SECONDS", "60")
	t.Setenv("WORK_LEASE_DURATION_SECONDS", "60")
	t.Setenv("MAX_OUTBOUND_EXECUTION_SECONDS", "30")
	t.Setenv("RESULT_COMMIT_BUFFER_SECONDS", "10")
	if _, _, _, _, _, err := executionSafetyConfig(os.Getenv); err != nil {
		t.Fatalf("executionSafetyConfig returned error: %v", err)
	}
}

func TestExecutionSafetyConfigRejectsShortWorkerTimeout(t *testing.T) {
	t.Setenv("WORKER_LAMBDA_TIMEOUT_SECONDS", "40")
	t.Setenv("EXECUTION_QUEUE_VISIBILITY_TIMEOUT_SECONDS", "60")
	t.Setenv("WORK_LEASE_DURATION_SECONDS", "60")
	t.Setenv("MAX_OUTBOUND_EXECUTION_SECONDS", "30")
	t.Setenv("RESULT_COMMIT_BUFFER_SECONDS", "10")
	if _, _, _, _, _, err := executionSafetyConfig(os.Getenv); err == nil {
		t.Fatal("expected worker timeout violation")
	}
}

func TestExecutionSafetyConfigRejectsShortLease(t *testing.T) {
	t.Setenv("WORKER_LAMBDA_TIMEOUT_SECONDS", "45")
	t.Setenv("EXECUTION_QUEUE_VISIBILITY_TIMEOUT_SECONDS", "60")
	t.Setenv("WORK_LEASE_DURATION_SECONDS", "30")
	t.Setenv("MAX_OUTBOUND_EXECUTION_SECONDS", "30")
	t.Setenv("RESULT_COMMIT_BUFFER_SECONDS", "10")
	if _, _, _, _, _, err := executionSafetyConfig(os.Getenv); err == nil {
		t.Fatal("expected lease violation")
	}
}

func TestExecutionSafetyConfigRejectsShortVisibility(t *testing.T) {
	t.Setenv("WORKER_LAMBDA_TIMEOUT_SECONDS", "45")
	t.Setenv("EXECUTION_QUEUE_VISIBILITY_TIMEOUT_SECONDS", "45")
	t.Setenv("WORK_LEASE_DURATION_SECONDS", "60")
	t.Setenv("MAX_OUTBOUND_EXECUTION_SECONDS", "30")
	t.Setenv("RESULT_COMMIT_BUFFER_SECONDS", "10")
	if _, _, _, _, _, err := executionSafetyConfig(os.Getenv); err == nil {
		t.Fatal("expected visibility violation")
	}
}

func TestRecordExecutionResultSkipsOlderRecurringProjection(t *testing.T) {
	cursor := time.Date(2026, 7, 19, 12, 5, 0, 0, time.UTC)
	record := resultstatus.MonitorStatus{
		TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http",
		CurrentStatus: "DOWN", ConsecutiveFailures: 1, ConsecutiveSuccesses: 0,
		LastCheckedAt: cursor, LastOutcome: checkexecution.OutcomeFailure,
		RecurringScheduledFor: &cursor, RecurringRunID: "RUN_FUTURE",
	}
	statusItem, err := sharedaws.MarshalMap(record.ToRecord())
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	client := &fakeDynamoClient{getItemOutput: &sharedaws.DynamoDBGetItemOutput{Item: statusItem}}
	repo := newDynamoRuntimeRepository(client, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, FailureThreshold: 1, RecoveryThreshold: 1}
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_OLD", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: cursor.Add(-time.Minute), FencingToken: "TOKEN"}
	older := cursor.Add(-time.Minute)
	result := checkexecution.ExecutionResult{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_OLD", Outcome: checkexecution.OutcomeSuccess, Trigger: checkexecution.TriggerTypeRecurring, ScheduledFor: &older, FinishedAt: older}

	if _, _, err := repo.RecordExecutionResult(context.Background(), monitor, work, result); err != nil {
		t.Fatalf("RecordExecutionResult returned error: %v", err)
	}
	if client.transactInput == nil {
		t.Fatal("transaction not captured")
	}
	for i, item := range client.transactInput.TransactItems {
		if item.Put != nil {
			if entity, _ := item.Put.Item["EntityType"].(*sharedaws.AttributeValueMemberS); entity != nil && entity.Value == dynamodbschema.EntityMonitorStatus {
				t.Fatalf("MonitorStatus projection should be skipped for older recurring key (index %d)", i)
			}
			if entity, _ := item.Put.Item["EntityType"].(*sharedaws.AttributeValueMemberS); entity != nil && entity.Value == dynamodbschema.EntityServiceStatus {
				t.Fatalf("ServiceStatus projection should be skipped for older recurring key (index %d)", i)
			}
		}
	}
}

func TestRecordExecutionResultWritesEqualTransitionAndOutboxIdentities(t *testing.T) {
	client := &fakeDynamoClient{}
	repo := newDynamoRuntimeRepository(client, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, FailureThreshold: 1, RecoveryThreshold: 1}
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_T1", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: time.Now(), FencingToken: "TOKEN"}
	scheduledFor := time.Now()
	result := checkexecution.ExecutionResult{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_T1", Outcome: checkexecution.OutcomeFailure, Trigger: checkexecution.TriggerTypeRecurring, ScheduledFor: &scheduledFor, FinishedAt: time.Now()}

	if _, _, err := repo.RecordExecutionResult(context.Background(), monitor, work, result); err != nil {
		t.Fatalf("RecordExecutionResult returned error: %v", err)
	}
	if client.transactInput == nil {
		t.Fatal("transaction not captured")
	}
	var outboxID, activityID string
	for _, item := range client.transactInput.TransactItems {
		if item.Put == nil {
			continue
		}
		entity, _ := item.Put.Item["EntityType"].(*sharedaws.AttributeValueMemberS)
		if entity == nil {
			continue
		}
		switch entity.Value {
		case dynamodbschema.EntityTransitionOutbox:
			if v, _ := item.Put.Item["TransitionID"].(*sharedaws.AttributeValueMemberS); v != nil {
				outboxID = v.Value
			}
		case dynamodbschema.EntityIncidentActivity:
			if v, _ := item.Put.Item["ActivityID"].(*sharedaws.AttributeValueMemberS); v != nil {
				activityID = v.Value
			}
		}
	}
	if outboxID == "" || activityID == "" || outboxID != activityID {
		t.Fatalf("transition identity must match activity identity: outbox=%q activity=%q", outboxID, activityID)
	}
}

func TestExecutionMaxConcurrencyAcceptsPositiveSetting(t *testing.T) {
	t.Setenv("EXECUTION_EVENT_SOURCE_MAX_CONCURRENCY", "8")
	if got := executionMaxConcurrency(os.Getenv); got != 8 {
		t.Fatalf("executionMaxConcurrency = %d, want 8", got)
	}
}

func TestExecutionMaxConcurrencyFallsBackToDefault(t *testing.T) {
	t.Setenv("EXECUTION_EVENT_SOURCE_MAX_CONCURRENCY", "")
	if got := executionMaxConcurrency(os.Getenv); got != defaultMaxConcurrency {
		t.Fatalf("executionMaxConcurrency = %d, want %d", got, defaultMaxConcurrency)
	}
}

func TestCanonicalInvariantsAcrossRecurringLifecycle(t *testing.T) {
	client := &fakeDynamoClient{}
	repo := newDynamoRuntimeRepository(client, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, FailureThreshold: 1, RecoveryThreshold: 1}
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	scheduled := now
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_INV", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now, ScheduleDefinitionVersion: "v1", ScheduledFor: &scheduled, FencingToken: "TOKEN"}
	failure := checkexecution.ExecutionResult{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_INV", Outcome: checkexecution.OutcomeFailure, Trigger: checkexecution.TriggerTypeRecurring, ScheduledFor: &scheduled, FinishedAt: now}
	success := checkexecution.ExecutionResult{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_INV", Outcome: checkexecution.OutcomeSuccess, Trigger: checkexecution.TriggerTypeRecurring, ScheduledFor: &scheduled, FinishedAt: now.Add(time.Minute)}

	if _, _, err := repo.RecordExecutionResult(context.Background(), monitor, work, failure); err != nil {
		t.Fatalf("first commit: %v", err)
	}
	recurrenceID := checkexecution.TransitionID(work.RunID)
	var sawOutbox, sawActivity, sawRun bool
	for _, item := range client.transactInput.TransactItems {
		if item.Put == nil {
			continue
		}
		entity, _ := item.Put.Item["EntityType"].(*sharedaws.AttributeValueMemberS)
		if entity == nil {
			continue
		}
		switch entity.Value {
		case dynamodbschema.EntityTransitionOutbox:
			if v, _ := item.Put.Item["TransitionID"].(*sharedaws.AttributeValueMemberS); v != nil && v.Value != recurrenceID {
				t.Fatalf("transition identity drift: outbox=%q recurrence=%q", v.Value, recurrenceID)
			}
			sawOutbox = true
		case dynamodbschema.EntityIncidentActivity:
			if v, _ := item.Put.Item["ActivityID"].(*sharedaws.AttributeValueMemberS); v != nil && v.Value != recurrenceID {
				t.Fatalf("activity identity drift: activity=%q recurrence=%q", v.Value, recurrenceID)
			}
			sawActivity = true
		case dynamodbschema.EntityCheckRun:
			sawRun = true
		}
	}
	if !sawOutbox || !sawActivity || !sawRun {
		t.Fatalf("expected one outbox+activity+run, got %v/%v/%v", sawOutbox, sawActivity, sawRun)
	}
	if _, _, err := repo.RecordExecutionResult(context.Background(), monitor, work, success); err != nil {
		t.Fatalf("recovery commit: %v", err)
	}
}

func TestDispatchPendingBucketShardsByShardCount(t *testing.T) {
	work := checkexecution.ExecutionWork{RunID: "RUN_BUCKET", AcceptedAt: time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)}
	bucket, shard := dispatchPendingBucket(work)
	if !strings.HasPrefix(bucket, "2026071912") {
		t.Fatalf("bucket = %q", bucket)
	}
	found := false
	for i := 0; i < dispatchPendingShards; i++ {
		if shard == fmt.Sprintf("%02x", i) {
			found = true
		}
	}
	if !found {
		t.Fatalf("shard = %q not in 0..%d", shard, dispatchPendingShards)
	}
}

func TestRecordExecutionResultWritesWorkRunAndStatusTogether(t *testing.T) {
	client := &fakeDynamoClient{}
	repo := newDynamoRuntimeRepository(client, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeManual, RequestedAt: time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC), Status: checkexecution.ExecutionWorkInProgress}
	statusCode := 200
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, RunID: "RUN_1", Type: "http", Trigger: checkexecution.TriggerTypeManual, StartedAt: time.Date(2026, 5, 22, 12, 0, 1, 0, time.UTC), FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC), DurationMs: 1000, Outcome: checkexecution.OutcomeSuccess, StatusCode: &statusCode}

	if _, _, err := repo.RecordExecutionResult(context.Background(), monitor, work, result); err != nil {
		t.Fatalf("RecordExecutionResult returned error: %v", err)
	}
	if client.transactInput == nil {
		t.Fatal("transact input not captured")
	}
	if len(client.transactInput.TransactItems) != 4 {
		t.Fatalf("transact items = %d, want 4", len(client.transactInput.TransactItems))
	}
	if got := sharedaws.ToString(client.transactInput.TransactItems[0].Update.TableName); got != "table-name" {
		t.Fatalf("table name = %q, want table-name", got)
	}
	if sharedaws.ToString(client.transactInput.TransactItems[0].Update.ConditionExpression) != "#status = :inProgress AND FencingToken = :token" {
		t.Fatalf("completion condition = %v", client.transactInput.TransactItems[0].Update.ConditionExpression)
	}
}

func TestMarkExecutionWorkSkippedFencesAndRemovesRecoveryMarker(t *testing.T) {
	client := &fakeDynamoClient{}
	repo := newDynamoRuntimeRepository(client, "table-name")
	work := checkexecution.ExecutionWork{
		TenantID: defaultTenantID, RunID: "RUN_1", AcceptedAt: time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC),
		Status: checkexecution.ExecutionWorkInProgress, FencingToken: "token",
	}

	if err := repo.MarkExecutionWorkSkipped(context.Background(), work, work.AcceptedAt, "monitor disabled"); err != nil {
		t.Fatalf("MarkExecutionWorkSkipped returned error: %v", err)
	}
	if client.transactInput == nil || len(client.transactInput.TransactItems) != 2 {
		t.Fatalf("transaction = %#v", client.transactInput)
	}
	update := client.transactInput.TransactItems[0].Update
	if sharedaws.ToString(update.ConditionExpression) != "#status = :inProgress AND FencingToken = :token" {
		t.Fatalf("skip condition = %v", update.ConditionExpression)
	}
	if client.transactInput.TransactItems[1].Delete == nil {
		t.Fatal("recovery marker delete missing")
	}
}

func TestIncidentRecordsForResultOpensIncidentOnFirstFailure(t *testing.T) {
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Outcome: checkexecution.OutcomeFailure, Error: "boom", FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	currentStatus := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "UP", ConsecutiveFailures: 0, ConsecutiveSuccesses: 0}
	thresholdConfig := resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}

	records, transition, _, _, err := decideExecutionResult(monitor, result, currentStatus, thresholdConfig, dynamodbrecord.IncidentRecord{}, false, newIncidentID, newAuditID)
	if err != nil {
		t.Fatalf("dynamodbrecord.IncidentRecordsForResult returned error: %v", err)
	}
	if transition != "incident.down" {
		t.Fatalf("transition = %q, want incident.down", transition)
	}
	incident := findIncidentRecord(t, records)
	if incident.Status != incidentStatusOpen {
		t.Fatalf("status = %q, want %q", incident.Status, incidentStatusOpen)
	}
	if incident.Type != "monitoring" {
		t.Fatalf("type = %q, want monitoring", incident.Type)
	}
}

func TestDecideExecutionResultUsesInjectedIdentifiers(t *testing.T) {
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, RunID: "RUN_1", Outcome: checkexecution.OutcomeFailure, FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	status := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "UP"}

	records, transition, incidentID, _, err := decideExecutionResult(monitor, result, status, resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}, dynamodbrecord.IncidentRecord{}, false, func(time.Time) string { return "INC_FIXED" }, func(time.Time) string { return "AUD_FIXED" })
	if err != nil {
		t.Fatalf("decideExecutionResult: %v", err)
	}
	if transition != "incident.down" || incidentID != "INC_FIXED" || findIncidentRecord(t, records).IncidentID != "INC_FIXED" {
		t.Fatalf("transition=%q incidentID=%q records=%+v", transition, incidentID, records)
	}
}

func TestIncidentRecordsForResultUpdatesExistingOpenIncident(t *testing.T) {
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	current := dynamodbrecord.IncidentRecord{IncidentID: "INC_1", ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Summary: "old", Status: incidentStatusOpen, OpenedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:59:00Z", Origin: "system"}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Outcome: checkexecution.OutcomeFailure, Error: "still bad", FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	currentStatus := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "DOWN", ConsecutiveFailures: 1, ConsecutiveSuccesses: 0}
	thresholdConfig := resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}

	records, transition, _, _, err := decideExecutionResult(monitor, result, currentStatus, thresholdConfig, current, true, newIncidentID, newAuditID)
	if err != nil {
		t.Fatalf("dynamodbrecord.IncidentRecordsForResult returned error: %v", err)
	}
	if transition != "" {
		t.Fatalf("transition = %q, want empty transition", transition)
	}
	incident := findIncidentRecord(t, records)
	if incident.IncidentID != "INC_1" {
		t.Fatalf("incidentID = %q, want INC_1", incident.IncidentID)
	}
	if incident.Status != incidentStatusOpen {
		t.Fatalf("status = %q, want %q", incident.Status, incidentStatusOpen)
	}
	if incident.UpdatedAt != "2026-05-22T12:00:02Z" {
		t.Fatalf("updatedAt = %q, want completion time", incident.UpdatedAt)
	}
}

func TestIncidentRecordsForResultResolvesOpenIncidentOnSuccess(t *testing.T) {
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	current := dynamodbrecord.IncidentRecord{IncidentID: "INC_1", ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Summary: "old", Status: incidentStatusOpen, OpenedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:59:00Z", Origin: "system"}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Outcome: checkexecution.OutcomeSuccess, FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	currentStatus := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "RECOVERING", ConsecutiveFailures: 0, ConsecutiveSuccesses: 1}
	thresholdConfig := resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}

	records, transition, _, _, err := decideExecutionResult(monitor, result, currentStatus, thresholdConfig, current, true, newIncidentID, newAuditID)
	if err != nil {
		t.Fatalf("dynamodbrecord.IncidentRecordsForResult returned error: %v", err)
	}
	if transition != "incident.up" {
		t.Fatalf("transition = %q, want incident.up", transition)
	}
	incident := findIncidentRecord(t, records)
	if incident.Status != incidentStatusResolved {
		t.Fatalf("status = %q, want %q", incident.Status, incidentStatusResolved)
	}
	if incident.ResolvedAt != "2026-05-22T12:00:02Z" {
		t.Fatalf("resolvedAt = %q, want completion time", incident.ResolvedAt)
	}
}

func findIncidentRecord(t *testing.T, records []any) dynamodbrecord.IncidentRecord {
	t.Helper()
	for _, record := range records {
		incidentItem, ok := record.(dynamodbrecord.IncidentItemRecord)
		if ok && incidentItem.SK == "META" {
			return incidentItem.ToIncident()
		}
	}
	t.Fatal("incident meta record not found")
	return dynamodbrecord.IncidentRecord{}
}

func TestGetOpenIncidentReadsPersistedOpenIncident(t *testing.T) {
	item, err := sharedaws.MarshalMap(dynamodbrecord.NewIncidentMonitorItemRecord(dynamodbrecord.IncidentRecord{IncidentID: "INC_1", ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Summary: "down", Status: incidentStatusOpen, OpenedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T12:00:00Z", Origin: "system"}))
	if err != nil {
		t.Fatalf("MarshalMap returned error: %v", err)
	}
	repo := newDynamoRuntimeRepository(&fakeDynamoClient{queryOutput: &sharedaws.DynamoDBQueryOutput{Items: []map[string]sharedaws.AttributeValue{item}}}, "table-name")
	incident, found, err := repo.getOpenIncident(context.Background(), defaultTenantID, "auth", "public-http")
	if err != nil {
		t.Fatalf("getOpenIncident returned error: %v", err)
	}
	if !found || incident.IncidentID != "INC_1" {
		t.Fatalf("incident = %+v, found = %v, want INC_1", incident, found)
	}
}
