package resultstatus

import (
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
)

// TestMonitorStateCasing pins the stored CurrentStatus values for every
// declared MonitorState. Persisted items must keep the UPPERCASE form so that
// legacy items decode to the canonical strings; API adapters lower-case at the
// boundary.
func TestMonitorStateCasing(t *testing.T) {
	cases := []struct {
		name  string
		state MonitorState
		want  string
	}{
		{"up", MonitorStateUp, "UP"},
		{"degraded", MonitorStateDegraded, "DEGRADED"},
		{"down", MonitorStateDown, "DOWN"},
		{"recovering", MonitorStateRecovering, "RECOVERING"},
		{"maintenance", MonitorStateMaintenance, "MAINTENANCE"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.state) != tc.want {
				t.Fatalf("state = %q, want %q", string(tc.state), tc.want)
			}
		})
	}
}

// TestNewMonitorStatusProducesUpperCaseState pins that the stored status is
// UPPERCASE regardless of the outcome.
func TestNewMonitorStatusProducesUpperCaseState(t *testing.T) {
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	result := checkexecution.ExecutionResult{
		ServiceID:  "auth",
		MonitorID:  "public-http",
		TenantID:   "DEFAULT",
		Type:       "http",
		Trigger:    checkexecution.TriggerTypeManual,
		StartedAt:  now,
		FinishedAt: now.Add(time.Second),
		DurationMs: 1000,
		Outcome:    checkexecution.OutcomeSuccess,
	}

	status := NewMonitorStatus(result)
	if status.CurrentStatus != string(MonitorStateUp) {
		t.Fatalf("CurrentStatus = %q, want %q (uppercase stored form)", status.CurrentStatus, MonitorStateUp)
	}

	failed := checkexecution.ExecutionResult{
		ServiceID: result.ServiceID, MonitorID: result.MonitorID, TenantID: result.TenantID,
		Type: result.Type, Trigger: result.Trigger,
		StartedAt: now, FinishedAt: now.Add(time.Second), DurationMs: 1000,
		Outcome: checkexecution.OutcomeFailure,
	}
	failedStatus := NewMonitorStatus(failed)
	if failedStatus.CurrentStatus != string(MonitorStateUp) {
		t.Fatalf("failed CurrentStatus = %q, want %q (first observation seeds UP)", failedStatus.CurrentStatus, MonitorStateUp)
	}
}

// TestMonitorStatusToRecordKeepsStatusUpperCase pins that the persisted record
// preserves the UPPERCASE state so existing items continue to decode.
func TestMonitorStatusToRecordKeepsStatusUpperCase(t *testing.T) {
	status := MonitorStatus{
		ServiceID:     "auth",
		MonitorID:     "public-http",
		TenantID:      "DEFAULT",
		CurrentStatus: string(MonitorStateDown),
		LastCheckedAt: time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC),
		LastOutcome:   checkexecution.OutcomeFailure,
	}

	record := status.ToRecord()
	if record.CurrentStatus != "DOWN" {
		t.Fatalf("record.CurrentStatus = %q, want DOWN", record.CurrentStatus)
	}
	if record.EntityType != "MonitorStatus" {
		t.Fatalf("record.EntityType = %q, want MonitorStatus", record.EntityType)
	}
	if record.PK != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("record.PK = %q, want composite monitor key", record.PK)
	}
	if record.SK != "STATUS" {
		t.Fatalf("record.SK = %q, want STATUS", record.SK)
	}
}
