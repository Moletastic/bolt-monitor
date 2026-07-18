package domainvalues

import (
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

func TestNewTenantIDRejectsBlank(t *testing.T) {
	cases := []string{"", " ", "\t\n"}
	for _, value := range cases {
		t.Run("blank/"+value, func(t *testing.T) {
			id, err := NewTenantID(value)
			if err == nil {
				t.Fatalf("expected error for %q", value)
			}
			typed, ok := sharederrors.As(err)
			if !ok || typed.Code != sharederrors.CodeValidationFailed {
				t.Fatalf("expected VALIDATION_FAILED, got %v", err)
			}
			if id != "" {
				t.Fatalf("expected zero value, got %q", id)
			}
		})
	}
}

func TestNewTenantIDCanonicalizesUpperCase(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"default", "DEFAULT"},
		{"DeFaUlT", "DEFAULT"},
		{"  trimmed  ", "TRIMMED"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			id, err := NewTenantID(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id.String() != tc.expected {
				t.Fatalf("got %q, want %q", id, tc.expected)
			}
		})
	}
}

func TestNewServiceIDCanonicalizesLowerCase(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Auth", "auth"},
		{"auth-api", "auth-api"},
		{"  Auth API  ", "auth api"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			id, err := NewServiceID(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id.String() != tc.expected {
				t.Fatalf("got %q, want %q", id, tc.expected)
			}
		})
	}
}

func TestNewServiceIDRejectsBlank(t *testing.T) {
	if _, err := NewServiceID("   "); err == nil {
		t.Fatal("expected validation error for blank service id")
	}
}

func TestNewMonitorIDCanonicalizesLowerCase(t *testing.T) {
	id, err := NewMonitorID("Public-HTTP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "public-http" {
		t.Fatalf("got %q, want public-http", id)
	}
}

func TestNewMonitorIDRejectsBlank(t *testing.T) {
	if _, err := NewMonitorID(""); err == nil {
		t.Fatal("expected validation error for blank monitor id")
	}
}

func TestNewIncidentIDCanonicalizesUpperCase(t *testing.T) {
	id, err := NewIncidentID("inc_42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "INC_42" {
		t.Fatalf("got %q, want INC_42", id)
	}
}

func TestNewIncidentIDRejectsBlank(t *testing.T) {
	if _, err := NewIncidentID("   "); err == nil {
		t.Fatal("expected validation error for blank incident id")
	}
}

func TestNewRunIDCanonicalizesUpperCase(t *testing.T) {
	id, err := NewRunID("run_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "RUN_ABC" {
		t.Fatalf("got %q, want RUN_ABC", id)
	}
}

func TestNewRunIDRejectsBlank(t *testing.T) {
	if _, err := NewRunID(""); err == nil {
		t.Fatal("expected validation error for blank run id")
	}
}

func TestNewPolicyIDCanonicalizesUpperCase(t *testing.T) {
	id, err := NewPolicyID("pol_main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "POL_MAIN" {
		t.Fatalf("got %q, want POL_MAIN", id)
	}
}

func TestNewPolicyIDRejectsBlank(t *testing.T) {
	if _, err := NewPolicyID("   "); err == nil {
		t.Fatal("expected validation error for blank policy id")
	}
}

func TestNewChannelIDCanonicalizesUpperCase(t *testing.T) {
	id, err := NewChannelID("ch_email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "CH_EMAIL" {
		t.Fatalf("got %q, want CH_EMAIL", id)
	}
}

func TestNewChannelIDRejectsBlank(t *testing.T) {
	if _, err := NewChannelID(""); err == nil {
		t.Fatal("expected validation error for blank channel id")
	}
}

func TestIdentifierErrorIncludesField(t *testing.T) {
	_, err := NewMonitorID("   ")
	if err == nil {
		t.Fatal("expected error")
	}
	typed, ok := sharederrors.As(err)
	if !ok {
		t.Fatalf("expected TypedError, got %T", err)
	}
	field, _ := typed.Details["field"].(string)
	if field != "monitorId" {
		t.Fatalf("field = %q, want monitorId", field)
	}
}
