package main

import (
	"context"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type fakeDynamoClient struct {
	transactInput *sharedaws.DynamoDBTransactWriteItemsInput
	queryOutput   *sharedaws.DynamoDBQueryOutput
	getItemOutput *sharedaws.DynamoDBGetItemOutput
}

func (f *fakeDynamoClient) GetItem(context.Context, *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	if f.getItemOutput != nil {
		return f.getItemOutput, nil
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
	if err := repo.EnqueueExecutionRequests(context.Background(), []checkexecution.ExecutionRequest{request}, acceptedAt); err != nil {
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

func (f *fakeDynamoClient) UpdateItem(context.Context, *sharedaws.DynamoDBUpdateItemInput) (*sharedaws.DynamoDBUpdateItemOutput, error) {
	return &sharedaws.DynamoDBUpdateItemOutput{}, nil
}

func (f *fakeDynamoClient) DeleteItem(context.Context, *sharedaws.DynamoDBDeleteItemInput) (*sharedaws.DynamoDBDeleteItemOutput, error) {
	return &sharedaws.DynamoDBDeleteItemOutput{}, nil
}

func (f *fakeDynamoClient) Scan(context.Context, *sharedaws.DynamoDBScanInput) (*sharedaws.DynamoDBScanOutput, error) {
	return &sharedaws.DynamoDBScanOutput{}, nil
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
	repo := newDynamoRuntimeRepository(&fakeDynamoClient{}, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Outcome: checkexecution.OutcomeFailure, Error: "boom", FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	currentStatus := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "UP", ConsecutiveFailures: 0, ConsecutiveSuccesses: 0}
	thresholdConfig := resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}

	records, transition, _, _, err := repo.incidentRecordsForResult(monitor, result, currentStatus, thresholdConfig, dynamodbrecord.IncidentRecord{}, false)
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

func TestIncidentRecordsForResultUpdatesExistingOpenIncident(t *testing.T) {
	repo := newDynamoRuntimeRepository(&fakeDynamoClient{}, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	current := dynamodbrecord.IncidentRecord{IncidentID: "INC_1", ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Summary: "old", Status: incidentStatusOpen, OpenedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:59:00Z", Origin: "system"}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Outcome: checkexecution.OutcomeFailure, Error: "still bad", FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	currentStatus := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "DOWN", ConsecutiveFailures: 1, ConsecutiveSuccesses: 0}
	thresholdConfig := resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}

	records, transition, _, _, err := repo.incidentRecordsForResult(monitor, result, currentStatus, thresholdConfig, current, true)
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
	repo := newDynamoRuntimeRepository(&fakeDynamoClient{}, "table-name")
	monitor := monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", FailureThreshold: 1, RecoveryThreshold: 1}
	current := dynamodbrecord.IncidentRecord{IncidentID: "INC_1", ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Summary: "old", Status: incidentStatusOpen, OpenedAt: "2026-05-22T11:59:00Z", UpdatedAt: "2026-05-22T11:59:00Z", Origin: "system"}
	result := checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Outcome: checkexecution.OutcomeSuccess, FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)}
	currentStatus := resultstatus.MonitorStatus{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, CurrentStatus: "RECOVERING", ConsecutiveFailures: 0, ConsecutiveSuccesses: 1}
	thresholdConfig := resultstatus.ThresholdConfig{FailureThreshold: 1, RecoveryThreshold: 1}

	records, transition, _, _, err := repo.incidentRecordsForResult(monitor, result, currentStatus, thresholdConfig, current, true)
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
