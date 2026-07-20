package notifications

import (
	"strings"
	"testing"
)

func TestCanonicalEnvelopeValidate(t *testing.T) {
	cases := []struct {
		name     string
		envelope CanonicalEnvelope
		wantErr  string
	}{
		{
			name:     "valid transition",
			envelope: CanonicalTransitionEnvelope("DEFAULT", "EV_1", CanonicalSourceTransition, "2026-01-01T00:00:00Z"),
		},
		{
			name: "missing version",
			envelope: CanonicalEnvelope{
				Kind: CanonicalKindTransition, SourceKind: CanonicalSourceTransition,
				TenantID: "DEFAULT", TransitionID: "EV_1", CreatedAt: "now",
			},
			wantErr: "version is required",
		},
		{
			name: "unsupported version",
			envelope: CanonicalEnvelope{
				Version: "99", Kind: CanonicalKindTransition, SourceKind: CanonicalSourceTransition,
				TenantID: "DEFAULT", TransitionID: "EV_1", CreatedAt: "now",
			},
			wantErr: "unsupported canonical envelope version",
		},
		{
			name: "unknown kind",
			envelope: CanonicalEnvelope{
				Version: CanonicalEnvelopeVersion, Kind: "unknown", SourceKind: CanonicalSourceTransition,
				TenantID: "DEFAULT", TransitionID: "EV_1", CreatedAt: "now",
			},
			wantErr: "unsupported canonical envelope kind",
		},
		{
			name: "missing tenant",
			envelope: CanonicalEnvelope{
				Version: CanonicalEnvelopeVersion, Kind: CanonicalKindTransition, SourceKind: CanonicalSourceTransition,
				TransitionID: "EV_1", CreatedAt: "now",
			},
			wantErr: "tenantId is required",
		},
		{
			name: "missing transitionId",
			envelope: CanonicalEnvelope{
				Version: CanonicalEnvelopeVersion, Kind: CanonicalKindTransition, SourceKind: CanonicalSourceTransition,
				TenantID: "DEFAULT", CreatedAt: "now",
			},
			wantErr: "transitionId is required",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.envelope.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDeliveryIdentityDeterministic(t *testing.T) {
	a := DeliveryIdentity("DEFAULT", "TRN_1", 2, "CH_1")
	b := DeliveryIdentity("default", "trn_1", 2, "ch_1")
	if a != b {
		t.Fatalf("expected deterministic id, got %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "dlv_") {
		t.Fatalf("delivery id should be prefixed, got %q", a)
	}
	c := DeliveryIdentity("DEFAULT", "TRN_1", 3, "CH_1")
	if a == c {
		t.Fatalf("step number must change identity")
	}
	d := DeliveryIdentity("DEFAULT", "TRN_2", 2, "CH_1")
	if a == d {
		t.Fatalf("transition id must change identity")
	}
	e := DeliveryIdentity("DEFAULT", "TRN_1", 2, "CH_2")
	if a == e {
		t.Fatalf("channel id must change identity")
	}
}

func TestDeliveryStateRoundTrip(t *testing.T) {
	for _, state := range []DeliveryState{
		DeliveryPending, DeliveryInFlight, DeliveryRetryable,
		DeliveryAmbiguous, DeliveryDelivered, DeliveryTerminalFailed,
	} {
		parsed, err := ParseDeliveryState(string(state))
		if err != nil {
			t.Fatalf("state %s: %v", state, err)
		}
		if parsed != state {
			t.Fatalf("round-trip mismatch: %s -> %s", state, parsed)
		}
		if err := state.Validate(); err != nil {
			t.Fatalf("state %s should validate: %v", state, err)
		}
	}
	if _, err := ParseDeliveryState("invalid"); err == nil {
		t.Fatalf("invalid state should fail parse")
	}
}

func TestDeliveryStateTerminal(t *testing.T) {
	terminal := []DeliveryState{DeliveryDelivered, DeliveryTerminalFailed}
	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Fatalf("%s should be terminal", s)
		}
	}
	active := []DeliveryState{DeliveryPending, DeliveryInFlight, DeliveryRetryable, DeliveryAmbiguous}
	for _, s := range active {
		if s.IsTerminal() {
			t.Fatalf("%s should not be terminal", s)
		}
	}
}

func TestRetryableClassification(t *testing.T) {
	retryable := []DeliveryOutcomeClass{OutcomeTimeout, OutcomeTransport, OutcomeThrottled, OutcomeProvider5xx}
	for _, c := range retryable {
		if !IsRetryableClass(c) {
			t.Fatalf("%s should be retryable", c)
		}
	}
	terminal := []DeliveryOutcomeClass{OutcomeInvalidConfig, OutcomeProvider4xx, OutcomeUnsupported}
	for _, c := range terminal {
		if IsRetryableClass(c) {
			t.Fatalf("%s should not be retryable", c)
		}
	}
}

func TestSendOutcomeValidate(t *testing.T) {
	if err := (SendOutcome{Class: OutcomeAccepted, Retryable: false}).Validate(); err != nil {
		t.Fatalf("accepted outcome should validate: %v", err)
	}
	if err := (SendOutcome{Class: ""}).Validate(); err == nil {
		t.Fatalf("empty class should fail")
	}
	if err := (SendOutcome{Class: "weird"}).Validate(); err == nil {
		t.Fatalf("unknown class should fail")
	}
}

func TestReplayKeyFingerprint(t *testing.T) {
	a := ReplayKeyFingerprint("DEFAULT", "INC_1", "DLV_1", "key-abc")
	b := ReplayKeyFingerprint("DEFAULT", "INC_1", "DLV_1", "key-abc")
	if a != b {
		t.Fatalf("fingerprint should be deterministic")
	}
	c := ReplayKeyFingerprint("DEFAULT", "INC_1", "DLV_1", "key-other")
	if a == c {
		t.Fatalf("different keys must produce different fingerprints")
	}
	d := ReplayKeyFingerprint("DEFAULT", "INC_2", "DLV_1", "key-abc")
	if a == d {
		t.Fatalf("different incident must produce different fingerprint")
	}
}
