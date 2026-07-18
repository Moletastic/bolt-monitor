package dynamodbschema

import "testing"

// TestNormalizationForms locks down the established canonical forms for every
// shared identifier family before any value-object migration changes the
// surrounding call sites. Each row pins the input shape and the expected
// canonical string so accidental case drift cannot slip in.
func TestNormalizationForms(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		normal   func(string) string
		expected string
	}{
		// Tenant identifiers are upper-cased and trimmed.
		{"tenant uppercase", "default", NormalizeToken, "DEFAULT"},
		{"tenant trimmed", "  default  ", NormalizeToken, "DEFAULT"},
		{"tenant mixed case", "DeFaUlT", NormalizeToken, "DEFAULT"},
		{"tenant lowercase stays", "default", NormalizeToken, "DEFAULT"},

		// Service and monitor slugs are lower-cased and trimmed.
		{"service lowercase", "Auth", NormalizeField, "auth"},
		{"service trimmed", "  Auth-API  ", NormalizeField, "auth-api"},
		{"monitor lowercase", "Public-HTTP", NormalizeField, "public-http"},
		{"monitor trimmed", "  PUBLIC_HTTP  ", NormalizeField, "public_http"},

		// Run, incident, audit, policy and channel identifiers are tokens (upper-case).
		{"run token", "run_abc", NormalizeToken, "RUN_ABC"},
		{"incident token", "inc_42", NormalizeToken, "INC_42"},
		{"audit token", "aud_xyz", NormalizeToken, "AUD_XYZ"},
		{"policy token", "pol_main", NormalizeToken, "POL_MAIN"},
		{"channel token", "ch_email", NormalizeToken, "CH_EMAIL"},
		{"audit change field token", "service.name", NormalizeToken, "SERVICE.NAME"},

		// Search prefixes are lower-cased and trimmed.
		{"search prefix lower", "Auth API", normalizeSearchPrefix, "auth api"},
		{"search prefix trimmed", "  AUTH  ", normalizeSearchPrefix, "auth"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.normal(tc.input)
			if got != tc.expected {
				t.Fatalf("normalize(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// TestPrimaryKeyConstructorsApplyNormalization pins the exact PK shapes for the
// canonical partition keys so that any future value-object migration that
// delegates to these constructors cannot accidentally drop the established
// normalization.
func TestPrimaryKeyConstructorsApplyNormalization(t *testing.T) {
	cases := []struct {
		name     string
		build    func() string
		expected string
	}{
		{"tenant", func() string { return TenantPK("default") }, "TENANT#DEFAULT"},
		{"service", func() string { return ServicePK("default", "Auth") }, "SERVICE#DEFAULT#AUTH"},
		{"monitor", func() string { return MonitorPK("default", "Auth", "Public-HTTP") }, "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP"},
		{"service ref", func() string { return ServiceRefSK("Auth") }, "SERVICE#AUTH"},
		{"monitor ref", func() string { return ServiceMonitorRefSK("Public-HTTP") }, "MONITOR#PUBLIC-HTTP"},
		{"incident", func() string { return IncidentPK("inc_42") }, "INCIDENT#INC_42"},
		{"audit", func() string { return AuditPK("aud_xyz") }, "AUDIT#AUD_XYZ"},
		{"search prefix", func() string { return SearchIndexSKPrefix("Auth API") }, "SEARCH#auth api"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.build()
			if got != tc.expected {
				t.Fatalf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestEscalationItemKeysApplyNormalization pins the policy, channel, and
// escalation-state key shapes so that later record moves into
// shared/dynamodbrecord cannot drop the established prefixes.
func TestEscalationItemKeysApplyNormalization(t *testing.T) {
	policy := EscalationPolicyItem("default", "pol_main")
	if policy.PK != "TENANT#DEFAULT" {
		t.Fatalf("policy PK = %q, want TENANT#DEFAULT", policy.PK)
	}
	if policy.SK != "ESCALATION_POLICY#POL_MAIN" {
		t.Fatalf("policy SK = %q, want ESCALATION_POLICY#POL_MAIN", policy.SK)
	}
	if policy.EntityType != EntityEscalationPolicy {
		t.Fatalf("policy EntityType = %q, want %q", policy.EntityType, EntityEscalationPolicy)
	}

	channel := NotificationChannelItem("default", "ch_email")
	if channel.PK != "TENANT#DEFAULT" {
		t.Fatalf("channel PK = %q, want TENANT#DEFAULT", channel.PK)
	}
	if channel.SK != "NOTIFICATION_CHANNEL#CH_EMAIL" {
		t.Fatalf("channel SK = %q, want NOTIFICATION_CHANNEL#CH_EMAIL", channel.SK)
	}
	if channel.EntityType != EntityNotificationChannel {
		t.Fatalf("channel EntityType = %q, want %q", channel.EntityType, EntityNotificationChannel)
	}

	state := EscalationStateItem("default", "inc_42")
	if state.PK != "INCIDENT#INC_42" {
		t.Fatalf("state PK = %q, want INCIDENT#INC_42", state.PK)
	}
	if state.SK != "ESCALATION_STATE" {
		t.Fatalf("state SK = %q, want ESCALATION_STATE", state.SK)
	}
	if state.EntityType != EntityEscalationState {
		t.Fatalf("state EntityType = %q, want %q", state.EntityType, EntityEscalationState)
	}
}
