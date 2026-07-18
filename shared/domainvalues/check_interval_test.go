package domainvalues

import (
	"testing"
	"time"

	sharederrors "bolt-monitor/shared/errors"
)

func TestIsAllowedIntervalSecondsMatchesCanonicalSet(t *testing.T) {
	cases := []struct {
		seconds int
		allowed bool
	}{
		{60, true},
		{120, true},
		{180, true},
		{300, true},
		{600, true},
		{900, true},
		{1800, true},
		{3600, true},

		{0, false},
		{1, false},
		{45, false},
		{7200, false},
	}
	for _, tc := range cases {
		got := IsAllowedIntervalSeconds(tc.seconds)
		if got != tc.allowed {
			t.Fatalf("IsAllowedIntervalSeconds(%d) = %v, want %v", tc.seconds, got, tc.allowed)
		}
	}
}

func TestNewCheckIntervalRejectsUnsupported(t *testing.T) {
	if _, err := NewCheckInterval(45); err == nil {
		t.Fatal("expected validation error for unsupported cadence")
	} else if typed, ok := sharederrors.As(err); !ok || typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("expected VALIDATION_FAILED, got %v", err)
	}
}

func TestNewCheckIntervalAcceptsAllowed(t *testing.T) {
	for _, value := range AllowedIntervalSeconds() {
		got, err := NewCheckInterval(value)
		if err != nil {
			t.Fatalf("NewCheckInterval(%d) error: %v", value, err)
		}
		if got.Seconds() != value {
			t.Fatalf("Seconds() = %d, want %d", got.Seconds(), value)
		}
	}
}

func TestCheckIntervalSecondsAndDurationRoundTrip(t *testing.T) {
	interval, err := NewCheckInterval(300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interval.Seconds() != 300 {
		t.Fatalf("Seconds() = %d, want 300", interval.Seconds())
	}
	if interval.Duration() != 5*time.Minute {
		t.Fatalf("Duration() = %v, want 5m", interval.Duration())
	}
}

func TestMustCheckIntervalReturnsValue(t *testing.T) {
	got := MustCheckInterval(120)
	if got.Seconds() != 120 {
		t.Fatalf("Seconds() = %d, want 120", got.Seconds())
	}
	if got.Duration() != 2*time.Minute {
		t.Fatalf("Duration() = %v, want 2m", got.Duration())
	}
}

func TestCheckIntervalStringIncludesSeconds(t *testing.T) {
	interval, err := NewCheckInterval(3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interval.String() != "1h0m0s" {
		t.Fatalf("String() = %q, want 1h0m0s", interval.String())
	}
}

func TestAllowedIntervalSecondsSortedAscending(t *testing.T) {
	values := AllowedIntervalSeconds()
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			t.Fatalf("values not sorted: %v", values)
		}
	}
}
