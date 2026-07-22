package main

import (
	"context"
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type fakeExecutionResultStore struct {
	loaded      bool
	committed   bool
	monitor     monitorconfig.Monitor
	work        checkexecution.ExecutionWork
	result      checkexecution.ExecutionResult
	records     []any
	status      resultstatus.MonitorStatus
	projection  bool
	publication executionResultPublication
}

func (f *fakeExecutionResultStore) LoadExecutionResultState(context.Context, checkexecution.ExecutionResult) (executionResultState, error) {
	f.loaded = true
	return executionResultState{status: resultstatus.MonitorStatus{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", CurrentStatus: "UP"}, statusFound: true}, nil
}

func (f *fakeExecutionResultStore) CommitExecutionResult(_ context.Context, monitor monitorconfig.Monitor, work checkexecution.ExecutionWork, result checkexecution.ExecutionResult, records []any, status resultstatus.MonitorStatus, projection bool, publication executionResultPublication) error {
	f.committed = true
	f.monitor = monitor
	f.work = work
	f.result = result
	f.records = records
	f.status = status
	f.projection = projection
	f.publication = publication
	return nil
}

type fixedExecutionResultClock struct{ at time.Time }

func (c fixedExecutionResultClock) Now() time.Time { return c.at }

type fixedExecutionResultIDs struct{}

func (fixedExecutionResultIDs) NewIncidentID(time.Time) string { return "INC_FIXED" }
func (fixedExecutionResultIDs) NewAuditID(time.Time) string    { return "AUD_FIXED" }

func TestExecutionResultCommandPersistsMatchingResult(t *testing.T) {
	store := &fakeExecutionResultStore{}
	command := newExecutionResultCommand(store, fixedExecutionResultClock{}, fixedExecutionResultIDs{})
	work := checkexecution.ExecutionWork{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeRecurring}
	result := checkexecution.ExecutionResult{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeRecurring}

	transition, incidentID, err := command.execute(context.Background(), monitorconfig.Monitor{FailureThreshold: 1, RecoveryThreshold: 1}, work, result)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !store.loaded || !store.committed || transition != "incident.down" || incidentID != "INC_FIXED" {
		t.Fatalf("loaded=%v committed=%v transition=%q incident=%q", store.loaded, store.committed, transition, incidentID)
	}
}

func TestExecutionResultCommandCommitsTransitionOutboxIntent(t *testing.T) {
	store := &fakeExecutionResultStore{}
	command := newExecutionResultCommand(store, fixedExecutionResultClock{}, fixedExecutionResultIDs{})
	finishedAt := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	work := checkexecution.ExecutionWork{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeRecurring, ScheduledFor: &finishedAt}
	result := checkexecution.ExecutionResult{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeRecurring, ScheduledFor: &finishedAt, Outcome: checkexecution.OutcomeFailure, FinishedAt: finishedAt}

	transition, incidentID, err := command.execute(context.Background(), monitorconfig.Monitor{Name: "Home", FailureThreshold: 1, RecoveryThreshold: 1}, work, result)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if transition != "incident.down" || incidentID != "INC_FIXED" {
		t.Fatalf("transition=%q incidentID=%q", transition, incidentID)
	}
	if !store.projection || store.publication.transition != "incident.down" || store.publication.incidentID != "INC_FIXED" {
		t.Fatalf("projection=%v publication=%+v", store.projection, store.publication)
	}
	if len(store.records) == 0 {
		t.Fatal("command did not provide incident persistence records")
	}
}

func TestExecutionResultCommandUsesInjectedClockForMissingCompletionTime(t *testing.T) {
	store := &fakeExecutionResultStore{}
	at := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	command := newExecutionResultCommand(store, fixedExecutionResultClock{at: at}, fixedExecutionResultIDs{})
	work := checkexecution.ExecutionWork{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeManual}
	result := checkexecution.ExecutionResult{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeManual}

	if _, _, err := command.execute(context.Background(), monitorconfig.Monitor{}, work, result); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !store.result.FinishedAt.Equal(at) {
		t.Fatalf("finishedAt=%s want %s", store.result.FinishedAt, at)
	}
}

func TestExecutionResultCommandRejectsMismatchedResult(t *testing.T) {
	store := &fakeExecutionResultStore{}
	command := newExecutionResultCommand(store, fixedExecutionResultClock{}, fixedExecutionResultIDs{})
	work := checkexecution.ExecutionWork{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "home", RunID: "run-1", Trigger: checkexecution.TriggerTypeRecurring}
	result := checkexecution.ExecutionResult{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "other", RunID: "run-1", Trigger: checkexecution.TriggerTypeRecurring}

	if _, _, err := command.execute(context.Background(), monitorconfig.Monitor{}, work, result); err == nil {
		t.Fatal("execute error = nil")
	}
	if store.loaded || store.committed {
		t.Fatal("persistence called for mismatched result")
	}
}
