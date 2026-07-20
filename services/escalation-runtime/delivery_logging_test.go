package main

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"bolt-monitor/shared/notifications"
)

func captureDeliveryLogs(t *testing.T, fn func()) []deliveryLog {
	t.Helper()
	original := deliveryLogWriter
	pr, pw := io.Pipe()
	deliveryLogWriter = pw
	done := make(chan []deliveryLog, 1)
	go func() {
		var lines []deliveryLog
		decoder := json.NewDecoder(pr)
		for {
			var entry deliveryLog
			if err := decoder.Decode(&entry); err != nil {
				break
			}
			lines = append(lines, entry)
		}
		done <- lines
	}()
	fn()
	_ = pw.Close()
	deliveryLogWriter = original
	lines := <-done
	return lines
}

func TestLogDeliveryDispatchedRedactsSecrets(t *testing.T) {
	lines := captureDeliveryLogs(t, func() {
		logDeliveryDispatched("DEFAULT\nBAD", "TRN\nBAD", 1)
	})
	if len(lines) != 1 {
		t.Fatalf("expected one log line, got %d", len(lines))
	}
	if lines[0].TenantID != "" || lines[0].TransitionID != "" {
		t.Fatalf("secret leakage: %+v", lines[0])
	}
}

func TestLogDeliveryResultClassifiesOutcome(t *testing.T) {
	cases := []struct {
		name     string
		class    notifications.DeliveryOutcomeClass
		expected string
	}{
		{"accepted", notifications.OutcomeAccepted, "delivered"},
		{"retryable timeout", notifications.OutcomeTimeout, "retryable_failed"},
		{"retryable transport", notifications.OutcomeTransport, "retryable_failed"},
		{"retryable throttled", notifications.OutcomeThrottled, "retryable_failed"},
		{"retryable 5xx", notifications.OutcomeProvider5xx, "retryable_failed"},
		{"retry exhausted", notifications.OutcomeRetryExhausted, "terminal_failed"},
		{"terminal invalid config", notifications.OutcomeInvalidConfig, "terminal_failed"},
		{"terminal 4xx", notifications.OutcomeProvider4xx, "terminal_failed"},
		{"terminal unsupported", notifications.OutcomeUnsupported, "terminal_failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := captureDeliveryLogs(t, func() {
				logDeliveryResult("DEFAULT", "TRN_1", "DLV_1", 2, "telegram", 1, tc.class)
			})
			if len(lines) != 1 {
				t.Fatalf("expected one line, got %d", len(lines))
			}
			if lines[0].Outcome != tc.expected {
				t.Fatalf("outcome = %q, want %q", lines[0].Outcome, tc.expected)
			}
		})
	}
}

func TestLogScheduleDispatchedAndReconciliation(t *testing.T) {
	lines := captureDeliveryLogs(t, func() {
		logScheduleDispatched("DEFAULT", "TRN_2", 3)
		logReconciled("DEFAULT", "TRN_3", "20260719")
		logStreamExhausted("DEFAULT", "TRN_4", "retries exhausted")
	})
	if len(lines) != 3 {
		t.Fatalf("expected three lines, got %d", len(lines))
	}
	outcomes := make([]string, 0, 3)
	for _, l := range lines {
		outcomes = append(outcomes, l.Outcome)
	}
	if !strings.Contains(strings.Join(outcomes, ","), "scheduled") {
		t.Fatalf("missing scheduled outcome: %v", outcomes)
	}
	if !strings.Contains(strings.Join(outcomes, ","), "reconciled") {
		t.Fatalf("missing reconciled outcome: %v", outcomes)
	}
	if !strings.Contains(strings.Join(outcomes, ","), "stream_exhausted") {
		t.Fatalf("missing stream exhausted outcome: %v", outcomes)
	}
}

func TestRedactSecretStripsQuotes(t *testing.T) {
	if redactSecret("DEFAULT\"hack") != "" {
		t.Fatalf("expected quote-bearing secret to be redacted")
	}
	if redactSecret("DEFAULT") != "DEFAULT" {
		t.Fatalf("plain identifier should pass through redact")
	}
}
