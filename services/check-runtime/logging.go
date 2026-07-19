package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"
)

type runtimeOutcome string

const (
	outcomeCreated        runtimeOutcome = "created"
	outcomeExisting       runtimeOutcome = "existing"
	outcomePublished      runtimeOutcome = "published"
	outcomeRecovered      runtimeOutcome = "recovered"
	outcomeClaimed        runtimeOutcome = "claimed"
	outcomeReclaimed      runtimeOutcome = "reclaimed"
	outcomeSkipped        runtimeOutcome = "skipped"
	outcomeDuplicate      runtimeOutcome = "duplicate"
	outcomeStale          runtimeOutcome = "stale"
	outcomeCompleted      runtimeOutcome = "completed"
	outcomeMarkerCleaned  runtimeOutcome = "marker_cleaned"
	outcomePublicationErr runtimeOutcome = "publication_failed"
	outcomeDispatchPend   runtimeOutcome = "dispatch_pending"
	outcomeLeaseLost      runtimeOutcome = "lease_lost"
	outcomeMalformed      runtimeOutcome = "malformed_envelope"
)

type runtimeLog struct {
	Outcome   runtimeOutcome `json:"outcome"`
	Operation string         `json:"operation"`
	TenantID  string         `json:"tenantId,omitempty"`
	RunID     string         `json:"runId,omitempty"`
	Monitor   string         `json:"monitor,omitempty"`
	Reason    string         `json:"reason,omitempty"`
	Time      time.Time      `json:"time"`
}

var runtimeLogWriter io.Writer = os.Stdout

func runtimeOutcomeLog(outcome runtimeOutcome, operation, tenantID, runID, monitor, reason string) {
	if runtimeLogWriter == nil {
		return
	}
	entry := runtimeLog{Outcome: outcome, Operation: operation, TenantID: redact(tenantID), RunID: redact(runID), Monitor: redact(monitor), Reason: redact(reason), Time: time.Now().UTC()}
	encoded, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = io.WriteString(runtimeLogWriter, string(encoded)+"\n")
}

func redact(value string) string {
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, "\n\r\t") {
		return ""
	}
	return value
}
