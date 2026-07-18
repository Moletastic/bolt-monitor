package monitorconfig

import (
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

// TestAllowedIntervalCadence pins the supported interval seconds values and
// confirms anything outside the allowed set fails validation.
func TestAllowedIntervalCadence(t *testing.T) {
	cases := []struct {
		name     string
		interval int
		allowed  bool
	}{
		{"one minute", 60, true},
		{"two minutes", 120, true},
		{"three minutes", 180, true},
		{"five minutes", 300, true},
		{"ten minutes", 600, true},
		{"fifteen minutes", 900, true},
		{"thirty minutes", 1800, true},
		{"one hour", 3600, true},

		{"zero", 0, false},
		{"negative", -60, false},
		{"one second unsupported", 1, false},
		{"five seconds unsupported", 5, false},
		{"forty five seconds unsupported", 45, false},
		{"two hours unsupported", 7200, false},
		{"one day unsupported", 86400, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsAllowedIntervalSeconds(tc.interval)
			if got != tc.allowed {
				t.Fatalf("IsAllowedIntervalSeconds(%d) = %v, want %v", tc.interval, got, tc.allowed)
			}
		})
	}
}

// TestAllowedIntervalSecondsSorted ensures the introspection helper returns a
// sorted slice so that value-object adapters and dashboard selectors see the
// canonical order.
func TestAllowedIntervalSecondsSorted(t *testing.T) {
	values := AllowedIntervalSeconds()
	if len(values) == 0 {
		t.Fatal("AllowedIntervalSeconds returned empty")
	}
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			t.Fatalf("values not sorted: %v", values)
		}
	}
	if values[0] != 60 {
		t.Fatalf("first interval = %d, want 60", values[0])
	}
	if values[len(values)-1] != 3600 {
		t.Fatalf("last interval = %d, want 3600", values[len(values)-1])
	}
}

// TestIntervalCharacterizationSurfacesValidationField pins that constructing a
// Monitor with an unsupported interval surfaces the typed validation error.
// This locks down the response contract: API responses surface the
// VALIDATION_FAILED code via the shared error registry.
func TestIntervalCharacterizationSurfacesValidationField(t *testing.T) {
	monitor := Monitor{
		TenantID:          "DEFAULT",
		ServiceID:         "auth",
		MonitorID:         "public-http",
		Name:              "Homepage",
		Type:              MonitorTypeHTTP,
		IntervalSeconds:   45,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
	}

	err := monitor.Validate()
	if err == nil {
		t.Fatal("expected validation error for unsupported interval")
	}
	typed, ok := sharederrors.As(err)
	if !ok {
		t.Fatalf("expected *TypedError, got %T", err)
	}
	if typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("error code = %q, want %q", typed.Code, sharederrors.CodeValidationFailed)
	}
}

// TestIntervalCharacterizationAcceptsAllAllowedValues pins that every canonical
// allowed interval validates cleanly.
func TestIntervalCharacterizationAcceptsAllAllowedValues(t *testing.T) {
	for _, interval := range AllowedIntervalSeconds() {
		monitor := Monitor{
			TenantID:          "DEFAULT",
			ServiceID:         "auth",
			MonitorID:         "public-http",
			Name:              "Homepage",
			Type:              MonitorTypeHTTP,
			IntervalSeconds:   interval,
			Enabled:           true,
			FailureThreshold:  1,
			RecoveryThreshold: 1,
			HTTP: &HTTPConfiguration{
				Target:    "https://example.com",
				Method:    "GET",
				TimeoutMs: 5000,
			},
		}
		if err := monitor.Validate(); err != nil {
			t.Fatalf("interval %d: unexpected error %v", interval, err)
		}
	}
}
