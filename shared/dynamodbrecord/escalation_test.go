package dynamodbrecord

import (
	"encoding/json"
	"testing"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/escalation"
)

// TestEscalationPolicyRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the escalation policy item family after normalization.
func TestEscalationPolicyRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	policy := escalation.EscalationPolicy{
		TenantID:    "default",
		PolicyID:    "pol_main",
		Name:        "Primary",
		Description: "Primary escalation policy",
		BusinessHoursPath: escalation.EscalationPath{
			Steps: []escalation.EscalationStep{
				{
					DelayMinutes: 0,
					Channels: []escalation.ChannelConfig{
						{Type: escalation.ChannelTypeEmail, Target: "ops@example.com", Config: json.RawMessage(`{"apiKey":"key","fromEmail":"from@example.com"}`)},
					},
				},
			},
		},
		OffHoursPath: escalation.EscalationPath{
			Steps: []escalation.EscalationStep{{ChannelID: "CH_1", DelayMinutes: 5}},
		},
		CreatedAt: "2026-05-17T22:00:00Z",
		UpdatedAt: "2026-05-17T22:00:00Z",
	}

	record := NewEscalationPolicyItemRecord(policy)
	roundTripAssertKeyAttributes(t, record,
		"TENANT#DEFAULT",
		"ESCALATION_POLICY#POL_MAIN",
		dynamodbschema.EntityEscalationPolicy,
		map[string]string{
			"TenantID": "DEFAULT",
			"PolicyID": "POL_MAIN",
		},
	)

	var decoded EscalationPolicyItemRecord
	wire, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if err := sharedaws.UnmarshalMap(wire, &decoded); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}
	if decoded.PolicyID != "POL_MAIN" {
		t.Fatalf("decoded PolicyID = %q", decoded.PolicyID)
	}
}

// TestEscalationPolicyRecordDefensiveCopyIsolatesChannels pins that the
// BusinessHoursPath and OffHoursPath fields are defensively copied so callers
// can mutate the source path without leaking changes into the record.
func TestEscalationPolicyRecordDefensiveCopyIsolatesChannels(t *testing.T) {
	original := escalation.EscalationPath{
		Steps: []escalation.EscalationStep{
			{
				DelayMinutes: 0,
				Channels: []escalation.ChannelConfig{
					{Type: escalation.ChannelTypeEmail, Target: "ops@example.com", Config: json.RawMessage(`{"apiKey":"key"}`)},
				},
			},
		},
	}
	policy := escalation.EscalationPolicy{
		TenantID: "DEFAULT", PolicyID: "POL_MAIN", Name: "Primary",
		BusinessHoursPath: original,
		OffHoursPath:      original,
	}

	record := NewEscalationPolicyItemRecord(policy)
	// Mutate the record copy and confirm the source path is unchanged.
	record.BusinessHoursPath.Steps[0].Channels[0].Target = "mutated"
	record.BusinessHoursPath.Steps[0].Channels[0].Config[0] = 'X'

	if original.Steps[0].Channels[0].Target != "ops@example.com" {
		t.Fatalf("source target leaked mutation: %s", original.Steps[0].Channels[0].Target)
	}
	if string(original.Steps[0].Channels[0].Config) != `{"apiKey":"key"}` {
		t.Fatalf("source config leaked mutation: %s", original.Steps[0].Channels[0].Config)
	}
}

// TestNotificationChannelRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the notification channel item family.
func TestNotificationChannelRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	channel := escalation.NotificationChannel{
		TenantID:  "default",
		ChannelID: "ch_email",
		Name:      "Ops Email",
		Type:      escalation.ChannelTypeEmail,
		Target:    "ops@example.com",
		Config:    json.RawMessage(`{"apiKey":"key","fromEmail":"from@example.com"}`),
	}

	record := NewNotificationChannelItemRecord(channel)
	roundTripAssertKeyAttributes(t, record,
		"TENANT#DEFAULT",
		"NOTIFICATION_CHANNEL#CH_EMAIL",
		dynamodbschema.EntityNotificationChannel,
		map[string]string{
			"TenantID":  "DEFAULT",
			"ChannelID": "CH_EMAIL",
		},
	)

	var decoded NotificationChannelItemRecord
	wire, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if err := sharedaws.UnmarshalMap(wire, &decoded); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}
	if decoded.ChannelID != "CH_EMAIL" {
		t.Fatalf("decoded ChannelID = %q", decoded.ChannelID)
	}
}

// TestNotificationChannelRecordDefensiveCopyIsolatesConfig pins that the
// Config payload is defensively copied so callers can mutate the source
// without leaking changes into the record.
func TestNotificationChannelRecordDefensiveCopyIsolatesConfig(t *testing.T) {
	channel := escalation.NotificationChannel{
		TenantID:  "DEFAULT",
		ChannelID: "CH_EMAIL",
		Name:      "Ops Email",
		Type:      escalation.ChannelTypeEmail,
		Target:    "ops@example.com",
		Config:    json.RawMessage(`{"apiKey":"key"}`),
	}

	record := NewNotificationChannelItemRecord(channel)
	if record.Config == nil {
		t.Fatal("record.Config is nil")
	}
	record.Config[0] = 'X'

	if string(channel.Config) != `{"apiKey":"key"}` {
		t.Fatalf("source config leaked mutation: %s", channel.Config)
	}

	// Round-trip back; the original record's config must survive intact.
	decoded := record.ToNotificationChannel()
	if string(decoded.Config) != string(record.Config) {
		t.Fatalf("decoded config differs: got %s want %s", decoded.Config, record.Config)
	}
}

// TestEscalationStateRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the per-incident escalation state item family.
func TestEscalationStateRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	state := escalation.EscalationState{
		TenantID:     "default",
		IncidentID:   "inc_42",
		PolicyID:     "pol_main",
		ServiceID:    "auth",
		MonitorID:    "public-http",
		CurrentStep:  2,
		StepsFired:   []int{1},
		SelectedPath: "business-hours",
		ScheduledFor: "2026-05-17T22:15:00Z",
		Status:       escalation.EscalationStatusActive,
		CreatedAt:    "2026-05-17T22:00:00Z",
		UpdatedAt:    "2026-05-17T22:05:00Z",
	}

	record := NewEscalationStateItemRecord(state)
	roundTripAssertKeyAttributes(t, record,
		"INCIDENT#INC_42",
		"ESCALATION_STATE",
		dynamodbschema.EntityEscalationState,
		map[string]string{
			"TenantID":   "DEFAULT",
			"IncidentID": "INC_42",
		},
	)

	var decoded EscalationStateItemRecord
	wire, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if err := sharedaws.UnmarshalMap(wire, &decoded); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}
	if decoded.IncidentID != "INC_42" {
		t.Fatalf("decoded IncidentID = %q", decoded.IncidentID)
	}
	if len(decoded.StepsFired) != 1 || decoded.StepsFired[0] != 1 {
		t.Fatalf("decoded StepsFired = %v", decoded.StepsFired)
	}
}

// TestEscalationStateRecordDefensiveCopyIsolatesStepsFired pins that the
// StepsFired slice is defensively copied so callers can append without
// mutating the record.
func TestEscalationStateRecordDefensiveCopyIsolatesStepsFired(t *testing.T) {
	state := escalation.EscalationState{
		TenantID:    "DEFAULT",
		IncidentID:  "INC_42",
		PolicyID:    "POL_MAIN",
		ServiceID:   "auth",
		MonitorID:   "public-http",
		CurrentStep: 2,
		StepsFired:  []int{1, 2},
		Status:      escalation.EscalationStatusActive,
	}

	record := NewEscalationStateItemRecord(state)
	record.StepsFired[0] = 999

	if state.StepsFired[0] != 1 {
		t.Fatalf("source StepsFired leaked mutation: %v", state.StepsFired)
	}
}

// TestEscalationStateRecordAcceptsLegacyCasedIdentifiers pins that the
// constructor normalizes legacy-cased identifiers into the canonical token
// form for tenant, incident, and policy and the canonical field form for
// service and monitor.
func TestEscalationStateRecordAcceptsLegacyCasedIdentifiers(t *testing.T) {
	state := escalation.EscalationState{
		TenantID:   "Default",
		IncidentID: "Inc_42",
		PolicyID:   "Pol_main",
		ServiceID:  "Auth",
		MonitorID:  "Public-HTTP",
		Status:     escalation.EscalationStatusActive,
	}

	record := NewEscalationStateItemRecord(state)
	if record.TenantID != "DEFAULT" {
		t.Fatalf("TenantID = %q, want DEFAULT", record.TenantID)
	}
	if record.IncidentID != "INC_42" {
		t.Fatalf("IncidentID = %q, want INC_42", record.IncidentID)
	}
	if record.PolicyID != "POL_MAIN" {
		t.Fatalf("PolicyID = %q, want POL_MAIN", record.PolicyID)
	}
	if record.ServiceID != "auth" {
		t.Fatalf("ServiceID = %q, want auth", record.ServiceID)
	}
	if record.MonitorID != "public-http" {
		t.Fatalf("MonitorID = %q, want public-http", record.MonitorID)
	}
	if record.PK != "INCIDENT#INC_42" {
		t.Fatalf("PK = %q, want INCIDENT#INC_42", record.PK)
	}
	if record.SK != "ESCALATION_STATE" {
		t.Fatalf("SK = %q, want ESCALATION_STATE", record.SK)
	}
}

// TestEscalationPolicyRecordAcceptsLegacyCasedIdentifiers pins the canonical
// wire shape for legacy-cased policy identifiers.
func TestEscalationPolicyRecordAcceptsLegacyCasedIdentifiers(t *testing.T) {
	policy := escalation.EscalationPolicy{
		TenantID: "Default",
		PolicyID: "Pol_main",
		Name:     "Primary",
	}
	record := NewEscalationPolicyItemRecord(policy)
	if record.PK != "TENANT#DEFAULT" {
		t.Fatalf("PK = %q, want TENANT#DEFAULT", record.PK)
	}
	if record.SK != "ESCALATION_POLICY#POL_MAIN" {
		t.Fatalf("SK = %q, want ESCALATION_POLICY#POL_MAIN", record.SK)
	}
	if record.PolicyID != "POL_MAIN" {
		t.Fatalf("PolicyID = %q, want POL_MAIN", record.PolicyID)
	}
}

// TestNotificationChannelRecordAcceptsLegacyCasedIdentifiers pins the
// canonical wire shape for legacy-cased channel identifiers.
func TestNotificationChannelRecordAcceptsLegacyCasedIdentifiers(t *testing.T) {
	channel := escalation.NotificationChannel{
		TenantID:  "Default",
		ChannelID: "Ch_email",
		Name:      "Ops Email",
		Type:      escalation.ChannelTypeEmail,
		Target:    "ops@example.com",
	}
	record := NewNotificationChannelItemRecord(channel)
	if record.PK != "TENANT#DEFAULT" {
		t.Fatalf("PK = %q", record.PK)
	}
	if record.SK != "NOTIFICATION_CHANNEL#CH_EMAIL" {
		t.Fatalf("SK = %q", record.SK)
	}
	if record.ChannelID != "CH_EMAIL" {
		t.Fatalf("ChannelID = %q", record.ChannelID)
	}
}
