package checkexecution

import (
	"testing"
	"time"
)

func TestRecurringRunIDNormalizesIdentity(t *testing.T) {
	scheduledFor := time.Date(2026, 7, 18, 10, 5, 0, 0, time.FixedZone("offset", -3*60*60))
	first := RecurringRunID(" default ", "AUTH", "Public-HTTP", "v1", scheduledFor)
	second := RecurringRunID("DEFAULT", "auth", "public-http", "v1", scheduledFor.UTC())
	if first != second {
		t.Fatalf("run IDs differ: %q != %q", first, second)
	}
	if first == RecurringRunID("DEFAULT", "auth", "public-http", "v2", scheduledFor.UTC()) {
		t.Fatal("schedule definition version did not affect run ID")
	}
}

func TestRuntimeFailureCarriesSafeIdentity(t *testing.T) {
	err := Publication("publish-work", "RUN_1")
	if !err.Retryable || err.Code != FailurePublication || err.RunID != "RUN_1" {
		t.Fatalf("failure = %#v", err)
	}
}

func TestTransitionIDIsStableForRunID(t *testing.T) {
	if TransitionID("run_1") != TransitionID(" RUN_1 ") || TransitionID("RUN_1") == TransitionID("RUN_2") {
		t.Fatal("transition ID is not stable and distinct")
	}
}

func TestScheduleIdentityUsesCurrentDefinitionAndBoundary(t *testing.T) {
	monitor := validExecutionMonitor()
	first := ScheduleDefinitionVersion(monitor)
	monitor.IntervalSeconds = 300
	if first == ScheduleDefinitionVersion(monitor) {
		t.Fatal("schedule version did not change with interval")
	}
	actual := ScheduledFor(time.Date(2026, 7, 18, 10, 5, 58, 0, time.UTC), 60)
	want := time.Date(2026, 7, 18, 10, 5, 0, 0, time.UTC)
	if !actual.Equal(want) {
		t.Fatalf("scheduledFor = %s, want %s", actual, want)
	}
}
