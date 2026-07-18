package domainvalues

import (
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

func TestNewMonitorStateRejectsBlank(t *testing.T) {
	if _, err := NewMonitorState(""); err == nil {
		t.Fatal("expected validation error for blank state")
	}
}

func TestNewMonitorStateRejectsUnsupported(t *testing.T) {
	if _, err := NewMonitorState("disabled"); err == nil {
		t.Fatal("expected validation error for unsupported state")
	} else if typed, ok := sharederrors.As(err); !ok || typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("expected VALIDATION_FAILED, got %v", err)
	}
}

func TestNewMonitorStateAcceptsCanonicalForms(t *testing.T) {
	for _, expected := range []MonitorState{
		MonitorStateUp,
		MonitorStateDegraded,
		MonitorStateDown,
		MonitorStateRecovering,
		MonitorStateMaintenance,
		MonitorStateUnknown,
	} {
		got, err := NewMonitorState(string(expected))
		if err != nil {
			t.Fatalf("state %q: unexpected error %v", expected, err)
		}
		if got != expected {
			t.Fatalf("got %q, want %q", got, expected)
		}
	}
}

func TestNewMonitorStateCanonicalizesMixedCase(t *testing.T) {
	got, err := NewMonitorState("up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != MonitorStateUp {
		t.Fatalf("got %q, want UP", got)
	}
}

func TestMonitorStateStoredAPIConversion(t *testing.T) {
	got := MonitorStateFromStored("down")
	if got.Stored() != "DOWN" {
		t.Fatalf("Stored() = %q, want DOWN", got.Stored())
	}
	if got.API() != "down" {
		t.Fatalf("API() = %q, want down", got.API())
	}
}

func TestMonitorStateFromAPINormalizes(t *testing.T) {
	got := MonitorStateFromAPI("UP")
	if got != MonitorStateUp {
		t.Fatalf("got %q, want UP", got)
	}
}

func TestIsMonitorStateAcceptsLegacyCasing(t *testing.T) {
	if !IsMonitorState("up") {
		t.Fatal("IsMonitorState should accept lower-case input")
	}
	if !IsMonitorState("DOWN") {
		t.Fatal("IsMonitorState should accept upper-case input")
	}
	if IsMonitorState("disabled") {
		t.Fatal("IsMonitorState should reject unsupported input")
	}
}

func TestAllowedMonitorStatesSorted(t *testing.T) {
	states := AllowedMonitorStates()
	if len(states) != 6 {
		t.Fatalf("got %d states, want 6", len(states))
	}
	for i := 1; i < len(states); i++ {
		if states[i-1] > states[i] {
			t.Fatalf("states not sorted: %v", states)
		}
	}
}
