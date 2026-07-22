package main

import (
	"context"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type recordingUpdateMonitorStore struct {
	monitor monitorconfig.Monitor
	updated monitorconfig.Monitor
}

func (s *recordingUpdateMonitorStore) GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error) {
	return s.monitor, true, nil
}

func (s *recordingUpdateMonitorStore) UpdateMonitor(_ context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	s.updated = monitor
	return monitor, nil
}

func TestUpdateMonitorCommandOwnsPatchAndPersistence(t *testing.T) {
	store := &recordingUpdateMonitorStore{monitor: monitorconfig.Monitor{
		TenantID: defaultTenantID, ServiceID: "payments", MonitorID: "api", Name: "API", Type: monitorconfig.MonitorTypeHTTP,
		Enabled: true, IntervalSeconds: 60, FailureThreshold: 2, RecoveryThreshold: 2,
		HTTP: &monitorconfig.HTTPConfiguration{Target: "https://api.example.com", Method: "GET", TimeoutMs: 5000},
	}}
	name := "Billing API"
	interval := 120
	updated, err := (updateMonitorCommand{store: store}).Execute(context.Background(), updateMonitorInput{
		TenantID: defaultTenantID, ServiceID: "payments", MonitorID: "api", Request: updateMonitorRequest{Name: &name, IntervalSeconds: &interval},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if updated.Name != "Billing API" || updated.IntervalSeconds != 120 || store.updated != updated {
		t.Fatalf("updated = %+v, persisted = %+v", updated, store.updated)
	}
}

type recordingMonitorStatusStore struct{ status resultstatus.MonitorStatus }

func (s recordingMonitorStatusStore) GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error) {
	return s.status, true, nil
}

func TestMonitorStatusQueryUsesOnlyStatusPort(t *testing.T) {
	status := resultstatus.MonitorStatus{MonitorID: "api", CurrentStatus: "up"}
	got, found, err := (monitorStatusQuery{store: recordingMonitorStatusStore{status: status}}).Execute(context.Background(), defaultTenantID, "payments", "api")
	if err != nil || !found || got != status {
		t.Fatalf("status = %+v, found = %v, err = %v", got, found, err)
	}
}

type recordingMonitorRunsStore struct {
	page historyPage[resultstatus.CheckRun]
}

type recordingMonitorAuditStore struct {
	found      bool
	auditCalls int
}

func (s *recordingMonitorAuditStore) GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error) {
	return monitorconfig.Monitor{}, s.found, nil
}

func (s *recordingMonitorAuditStore) ListMonitorAuditEventsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	s.auditCalls++
	return historyPage[auditEventView]{Items: []auditEventView{{AuditID: "evt-1"}}}, nil
}

func TestMonitorAuditQueryChecksMonitorBeforeReadingAudit(t *testing.T) {
	store := &recordingMonitorAuditStore{}
	page, found, err := (monitorAuditQuery{monitors: store, store: store}).Execute(context.Background(), defaultTenantID, "payments", "api", historyPageSize, nil)
	if err != nil || found || store.auditCalls != 0 || len(page.Items) != 0 {
		t.Fatalf("page = %+v, found = %v, calls = %d, err = %v", page, found, store.auditCalls, err)
	}

	store.found = true
	page, found, err = (monitorAuditQuery{monitors: store, store: store}).Execute(context.Background(), defaultTenantID, "payments", "api", historyPageSize, nil)
	if err != nil || !found || store.auditCalls != 1 || len(page.Items) != 1 {
		t.Fatalf("page = %+v, found = %v, calls = %d, err = %v", page, found, store.auditCalls, err)
	}
}

func (s recordingMonitorRunsStore) ListMonitorRunsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error) {
	return s.page, nil
}

func TestMonitorRunsQueryPassesThroughPage(t *testing.T) {
	want := historyPage[resultstatus.CheckRun]{Items: []resultstatus.CheckRun{{RunID: "run-1"}}}
	got, err := (monitorRunsQuery{store: recordingMonitorRunsStore{page: want}}).Execute(context.Background(), defaultTenantID, "payments", "api", historyPageSize, nil)
	if err != nil || len(got.Items) != 1 || got.Items[0].RunID != "run-1" {
		t.Fatalf("page = %+v, err = %v", got, err)
	}
}

type recordingManualRunStore struct {
	monitor  monitorconfig.Monitor
	reserved manualIdempotencyRecord
	recorded bool
}

func (s *recordingManualRunStore) GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error) {
	return s.monitor, true, nil
}

func (s *recordingManualRunStore) ReserveManualIdempotency(_ context.Context, record manualIdempotencyRecord) (manualIdempotencyRecord, error) {
	s.reserved = record
	return record, nil
}

func (s *recordingManualRunStore) RecordExecutionResult(context.Context, monitorconfig.Monitor, string, checkexecution.ExecutionResult) error {
	s.recorded = true
	return nil
}

func TestManualRunCommandReservesThenRecordsExecution(t *testing.T) {
	store := &recordingManualRunStore{monitor: monitorconfig.Monitor{
		TenantID: defaultTenantID, ServiceID: "payments", MonitorID: "api", Name: "API", Type: monitorconfig.MonitorTypeHTTP,
		Enabled: true, IntervalSeconds: 60, FailureThreshold: 1, RecoveryThreshold: 1,
		HTTP: &monitorconfig.HTTPConfiguration{Target: "https://api.example.com", Method: "GET", TimeoutMs: 5000},
	}}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	operations := newMonitorOperations(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, store, func() time.Time { return now }, identifierGenerator{newRunID: func(time.Time) string { return "RUN_TEST" }}, &recordingMonitorExecutor{}, nil)
	result, err := operations.manualRun.Execute(context.Background(), defaultTenantID, "payments", "api", "key-1")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if store.reserved.RunID == "" || !store.recorded || result.Execution == nil || result.Record.RunID != store.reserved.RunID {
		t.Fatalf("result = %+v, reserved = %+v, recorded = %v", result, store.reserved, store.recorded)
	}
}
