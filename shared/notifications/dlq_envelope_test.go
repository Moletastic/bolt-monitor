package notifications

import (
	"strings"
	"testing"
	"time"
)

func TestDLQEnvelopeValidate(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	envelope, err := StreamExhaustionEnvelope("DEFAULT", "TRN_1", "stream exhausted", now)
	if err != nil {
		t.Fatalf("valid envelope rejected: %v", err)
	}
	if envelope.SourceKind != DLQSourceStream {
		t.Fatalf("source = %s, want dynamodb_stream", envelope.SourceKind)
	}
	if envelope.FailureReason != "stream exhausted" {
		t.Fatalf("reason = %q", envelope.FailureReason)
	}
	if _, err := StreamExhaustionEnvelope("", "TRN_1", "", now); err == nil {
		t.Fatalf("empty tenant should fail")
	}
	if _, err := StreamExhaustionEnvelope("DEFAULT", "TRN_1", strings.Repeat("x", 500), now); err != nil {
		t.Fatalf("safe reason should not fail validation: %v", err)
	} else if len(strings.TrimSpace(envelope.FailureReason)) > 243 {
		t.Fatalf("reason was not truncated")
	}
}

func TestDLQEnvelopeRoundTrip(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	original, err := SchedulerExhaustionEnvelope("DEFAULT", "TRN_2", 3, "retries exhausted", now)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	body, err := original.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	parsed, err := ParseDLQEnvelope(body)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parsed.StepNumber != 3 || parsed.SourceKind != DLQSourceSchedule {
		t.Fatalf("round-trip mismatch: %+v", parsed)
	}
}

func TestDLQEnvelopeRejectsUnknownKind(t *testing.T) {
	if _, err := ParseDLQEnvelope(`{"sourceKind":"unknown","tenantId":"DEFAULT","observedAt":"2026-01-01T00:00:00Z"}`); err == nil {
		t.Fatalf("unknown source kind should fail")
	}
	if _, err := ParseDLQEnvelope(`not-json`); err == nil {
		t.Fatalf("malformed envelope should fail")
	}
}

func TestSQSDeliveryDLQEnvelope(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	envelope, err := SQSEnvelope("DEFAULT", "TRN_1", "DLV_1", "retryable_failed", now)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if envelope.DeliveryID != "DLV_1" || envelope.SourceKind != DLQSourceSQS {
		t.Fatalf("envelope mismatch: %+v", envelope)
	}
}
