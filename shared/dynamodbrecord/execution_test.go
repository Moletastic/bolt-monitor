package dynamodbrecord

import (
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/monitorconfig"
)

func TestNewExecutionWorkItemRecordSetsRetentionTTL(t *testing.T) {
	acceptedAt := "2026-05-16T10:00:00Z"
	record := NewExecutionWorkItemRecord("default", "auth", "public-http", "run_123", "iad", checkexecution.TriggerTypeManual, acceptedAt, checkexecution.ExecutionWorkPending, nil, nil, "")
	expected := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC).Unix()
	if record.TTL != expected {
		t.Fatalf("TTL = %d, want %d", record.TTL, expected)
	}
}

func TestExecutionWorkItemRecordFromWorkRecomputesRetentionTTL(t *testing.T) {
	requestedAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	startedAt := time.Date(2026, 5, 16, 10, 0, 5, 0, time.UTC)
	record := ExecutionWorkItemRecordFromWork(checkexecution.ExecutionWork{TenantID: "default", ServiceID: "auth", MonitorID: "public-http", RunID: "run_123", ProbeLocationID: "iad", Trigger: checkexecution.TriggerTypeManual, RequestedAt: requestedAt, Status: checkexecution.ExecutionWorkInProgress, StartedAt: &startedAt})
	expected := requestedAt.Add(checkexecution.DefaultExecutionWorkRetentionDays * 24 * time.Hour).Unix()
	if record.TTL != expected {
		t.Fatalf("TTL = %d, want %d", record.TTL, expected)
	}
}

func TestNewExecutionWorkItemRecordsSetsRetentionTTL(t *testing.T) {
	monitor := monitorconfig.Monitor{TenantID: "default", ServiceID: "auth", MonitorID: "public-http", ProbeLocations: []string{"iad"}}
	records := NewExecutionWorkItemRecords(monitor, checkexecution.TriggerTypeManual, "run_123", "2026-05-16T10:00:00Z")
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	expected := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC).Unix()
	if records[0].TTL != expected {
		t.Fatalf("TTL = %d, want %d", records[0].TTL, expected)
	}
}
