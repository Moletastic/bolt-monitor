package dynamodbrecord

import (
	"strings"
	"testing"

	"bolt-monitor/shared/notifications"
)

func TestNewDeliveryItemRecordRoundTrip(t *testing.T) {
	delivery := notifications.DeliveryRecord{
		TenantID:     "DEFAULT",
		IncidentID:   "INC_1",
		TransitionID: "TRN_1",
		DeliveryID:   "dlv_abc",
		ChannelID:    "CH_1",
		ChannelType:  "telegram",
		StepNumber:   2,
		State:        notifications.DeliveryPending,
		AttemptCount: 0,
		ProviderMetadata: notifications.ProviderMetadata{
			ProviderStatusClass: "2xx",
			ProviderRequestID:   "req-1",
			RetryAfterSeconds:   30,
		},
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-01T00:00:01Z",
	}
	record := NewDeliveryItemRecord(delivery)
	if !strings.HasPrefix(record.SK, "DELIVERY#2026-01-01T00:00:00Z#DLV_ABC") {
		t.Fatalf("unexpected SK %q", record.SK)
	}
	if record.EntityType != "Delivery" {
		t.Fatalf("unexpected entity %q", record.EntityType)
	}
	if record.ProviderStatusClass != "2xx" {
		t.Fatalf("provider status class lost: %q", record.ProviderStatusClass)
	}
	if !record.BelongsToTenant("default", "inc_1") {
		t.Fatalf("ownership should match case-insensitively")
	}
	if record.BelongsToTenant("OTHER", "INC_1") {
		t.Fatalf("ownership should reject other tenant")
	}
	back := record.ToDelivery()
	if back.ProviderMetadata.ProviderStatusClass != "2xx" || back.ProviderMetadata.RetryAfterSeconds != 30 {
		t.Fatalf("round-trip metadata mismatch: %+v", back.ProviderMetadata)
	}
}

func TestEscalationPlanRecordRoundTrip(t *testing.T) {
	plan := notifications.EscalationPlan{
		TenantID:     "DEFAULT",
		IncidentID:   "INC_1",
		TransitionID: "TRN_1",
		PolicyID:     "POL_1",
		SelectedPath: "business-hours",
		StepNumbers:  []int{1, 2},
		StepChannels: []string{"CH_1", "CH_2"},
		CreatedAt:    "2026-01-01T00:00:00Z",
	}
	record := NewEscalationPlanItemRecord(plan)
	if record.SK != "ESCALATION_PLAN#TRN_1" {
		t.Fatalf("unexpected SK %q", record.SK)
	}
	back := record.ToEscalationPlan()
	if back.PolicyID != "POL_1" || len(back.StepNumbers) != 2 {
		t.Fatalf("round-trip mismatch: %+v", back)
	}
	if !record.BelongsToTenant("DEFAULT", "INC_1") {
		t.Fatalf("plan ownership failed")
	}
}

func TestReplayIdempotencyRoundTrip(t *testing.T) {
	rec := notifications.ReplayIdempotencyRecord{
		TenantID:           "DEFAULT",
		IncidentID:         "INC_1",
		DeliveryID:         "DLV_1",
		Operation:          "delivery_replay",
		IdempotencyKey:     "key-abc",
		RequestFingerprint: "fingerprint",
		ResultDeliveryID:   "DLV_1",
		CreatedAt:          "2026-01-01T00:00:00Z",
		ExpiresAt:          1000,
	}
	item := NewReplayIdempotencyItemRecord(rec)
	if !strings.HasPrefix(item.SK, "REPLAY_IDEMPOTENCY#") {
		t.Fatalf("unexpected SK %q", item.SK)
	}
	if !item.MatchesFingerprint("fingerprint") {
		t.Fatalf("fingerprint match failed")
	}
	if item.MatchesFingerprint("other") {
		t.Fatalf("fingerprint should not match")
	}
	back := item.ToRecord()
	if back.RequestFingerprint != "fingerprint" || back.ResultDeliveryID != "DLV_1" {
		t.Fatalf("round-trip mismatch: %+v", back)
	}
}

func TestValidateReplayIdempotencyKey(t *testing.T) {
	if err := ValidateReplayIdempotencyKey(""); err == nil {
		t.Fatalf("empty key should fail")
	}
	if err := ValidateReplayIdempotencyKey("ok"); err != nil {
		t.Fatalf("non-empty key should pass: %v", err)
	}
	if err := ValidateReplayIdempotencyKey(strings.Repeat("x", 201)); err == nil {
		t.Fatalf("oversized key should fail")
	}
}
