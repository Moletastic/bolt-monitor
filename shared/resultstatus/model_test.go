package resultstatus

import (
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
)

func sampleResult() checkexecution.ExecutionResult {
	statusCode := 200
	return checkexecution.ExecutionResult{
		ServiceID:  "auth",
		MonitorID:  "public-http",
		TenantID:   "DEFAULT",
		Type:       "http",
		Trigger:    checkexecution.TriggerTypeManual,
		StartedAt:  time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC),
		FinishedAt: time.Date(2026, 5, 17, 22, 0, 1, 0, time.UTC),
		DurationMs: 1000,
		Outcome:    checkexecution.OutcomeSuccess,
		StatusCode: &statusCode,
	}
}

func TestNewCheckRunSetsTTL(t *testing.T) {
	now := time.Date(2026, 5, 17, 22, 0, 1, 0, time.UTC)
	run := NewCheckRun(sampleResult(), now)
	expectedTTL := now.Add(DefaultCheckRunRetentionDays * 24 * time.Hour).Unix()
	if run.TTL != expectedTTL {
		t.Fatalf("TTL = %d, want %d", run.TTL, expectedTTL)
	}
	if run.MonitorID != "public-http" {
		t.Fatalf("MonitorID = %q, want public-http", run.MonitorID)
	}
	if run.ServiceID != "auth" {
		t.Fatalf("ServiceID = %q, want auth", run.ServiceID)
	}
}

func TestMonitorStatusToRecordUsesStatusItemKeys(t *testing.T) {
	status := NewMonitorStatus(sampleResult())
	record := status.ToRecord()
	if record.PK != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("PK = %q, want composite monitor key", record.PK)
	}
	if record.SK != "STATUS" {
		t.Fatalf("SK = %q, want STATUS", record.SK)
	}
}

func TestCheckRunToRecordUsesRunPrefix(t *testing.T) {
	run := NewCheckRun(sampleResult(), time.Date(2026, 5, 17, 22, 0, 1, 0, time.UTC))
	record := run.ToRecord()
	if record.PK != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("PK = %q, want composite monitor key", record.PK)
	}
	if len(record.SK) < 5 || record.SK[:4] != "RUN#" {
		t.Fatalf("SK = %q, want RUN# prefix", record.SK)
	}
}

func TestNewCheckRunPreservesProvidedRunID(t *testing.T) {
	result := sampleResult()
	result.RunID = "RUN_MANUAL_1"
	run := NewCheckRun(result, time.Date(2026, 5, 17, 22, 0, 1, 0, time.UTC))
	if run.RunID != "RUN_MANUAL_1" {
		t.Fatalf("RunID = %q, want RUN_MANUAL_1", run.RunID)
	}
}
