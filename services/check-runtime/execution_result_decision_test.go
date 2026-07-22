package main

import (
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

func TestDecideExecutionResultCharacterization(t *testing.T) {
	finishedAt := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	monitor := monitorconfig.Monitor{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", Name: "Homepage"}
	openIncident := dynamodbrecord.IncidentRecord{IncidentID: "INC_EXISTING", TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", Status: incidentStatusOpen, Summary: "old", OpenedAt: "2026-07-22T11:00:00Z"}
	tests := []struct {
		name                string
		status              resultstatus.MonitorStatus
		outcome             checkexecution.Outcome
		trigger             checkexecution.TriggerType
		failureThreshold    int
		recoveryThreshold   int
		current             dynamodbrecord.IncidentRecord
		incidentFound       bool
		wantStatus          string
		wantFailures        int
		wantSuccesses       int
		wantTransition      string
		wantIncidentID      string
		wantIncidentStatus  string
		wantAuditAction     string
		wantIncidentRecords int
	}{
		{name: "manual result refreshes details without transition", status: resultstatus.MonitorStatus{CurrentStatus: "UP", ConsecutiveFailures: 3}, outcome: checkexecution.OutcomeFailure, trigger: checkexecution.TriggerTypeManual, wantStatus: "UP", wantFailures: 3},
		{name: "first failure below threshold degrades", status: resultstatus.MonitorStatus{CurrentStatus: "UP"}, outcome: checkexecution.OutcomeFailure, failureThreshold: 2, wantStatus: "DEGRADED", wantFailures: 1},
		{name: "threshold failure opens incident with audit and transition", status: resultstatus.MonitorStatus{CurrentStatus: "DEGRADED", ConsecutiveFailures: 1}, outcome: checkexecution.OutcomeFailure, failureThreshold: 2, wantStatus: "DOWN", wantFailures: 2, wantTransition: "incident.down", wantIncidentID: "INC_FIXED", wantIncidentStatus: incidentStatusOpen, wantAuditAction: "INCIDENT_OPENED", wantIncidentRecords: 6},
		{name: "ongoing failure updates open incident without notification transition", status: resultstatus.MonitorStatus{CurrentStatus: "DOWN", ConsecutiveFailures: 2}, outcome: checkexecution.OutcomeFailure, current: openIncident, incidentFound: true, wantStatus: "DOWN", wantFailures: 3, wantIncidentStatus: incidentStatusOpen, wantAuditAction: "INCIDENT_UPDATED", wantIncidentRecords: 6},
		{name: "first recovery success enters recovering", status: resultstatus.MonitorStatus{CurrentStatus: "DOWN", ConsecutiveFailures: 2}, outcome: checkexecution.OutcomeSuccess, recoveryThreshold: 2, wantStatus: "RECOVERING", wantSuccesses: 1},
		{name: "recovery threshold resolves incident with audit and transition", status: resultstatus.MonitorStatus{CurrentStatus: "RECOVERING", ConsecutiveSuccesses: 1}, outcome: checkexecution.OutcomeSuccess, recoveryThreshold: 2, current: openIncident, incidentFound: true, wantStatus: "UP", wantTransition: "incident.up", wantIncidentID: "INC_EXISTING", wantIncidentStatus: incidentStatusResolved, wantAuditAction: "INCIDENT_RESOLVED", wantIncidentRecords: 6},
		{name: "maintenance failure remains in maintenance", status: resultstatus.MonitorStatus{CurrentStatus: "MAINTENANCE"}, outcome: checkexecution.OutcomeFailure, wantStatus: "MAINTENANCE", wantFailures: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkexecution.ExecutionResult{TenantID: monitor.TenantID, ServiceID: monitor.ServiceID, MonitorID: monitor.MonitorID, RunID: "RUN_1", Trigger: tt.trigger, Outcome: tt.outcome, Error: "boom", FinishedAt: finishedAt}
			records, transition, incidentID, status, err := decideExecutionResult(monitor, result, tt.status, resultstatus.ThresholdConfig{FailureThreshold: tt.failureThreshold, RecoveryThreshold: tt.recoveryThreshold}, tt.current, tt.incidentFound, func(time.Time) string { return "INC_FIXED" }, func(time.Time) string { return "AUD_FIXED" })
			if err != nil {
				t.Fatalf("decideExecutionResult: %v", err)
			}
			if status.CurrentStatus != tt.wantStatus || status.ConsecutiveFailures != tt.wantFailures || status.ConsecutiveSuccesses != tt.wantSuccesses {
				t.Fatalf("status=%+v, want state=%s failures=%d successes=%d", status, tt.wantStatus, tt.wantFailures, tt.wantSuccesses)
			}
			if transition != tt.wantTransition || incidentID != tt.wantIncidentID || len(records) != tt.wantIncidentRecords {
				t.Fatalf("transition=%q incidentID=%q records=%d", transition, incidentID, len(records))
			}
			if tt.wantIncidentRecords == 0 {
				return
			}
			incident := findIncidentRecord(t, records)
			if incident.Status != tt.wantIncidentStatus {
				t.Fatalf("incident status=%q, want %q", incident.Status, tt.wantIncidentStatus)
			}
			for _, record := range records {
				audit, ok := record.(dynamodbrecord.AuditEventRecord)
				if !ok {
					continue
				}
				if audit.Action != tt.wantAuditAction {
					t.Fatalf("audit action=%q, want %q", audit.Action, tt.wantAuditAction)
				}
				return
			}
			t.Fatal("audit event record not found")
		})
	}
}
