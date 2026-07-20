package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"bolt-monitor/shared/notifications"
)

type deliveryLog struct {
	Outcome      string                             `json:"outcome"`
	Operation    string                             `json:"operation"`
	SourceKind   string                             `json:"sourceKind,omitempty"`
	TenantID     string                             `json:"tenantId,omitempty"`
	TransitionID string                             `json:"transitionId,omitempty"`
	DeliveryID   string                             `json:"deliveryId,omitempty"`
	StepNumber   int                                `json:"stepNumber,omitempty"`
	ChannelType  string                             `json:"channelType,omitempty"`
	Attempt      int                                `json:"attempt,omitempty"`
	OutcomeClass notifications.DeliveryOutcomeClass `json:"outcomeClass,omitempty"`
	Reason       string                             `json:"reason,omitempty"`
	Time         time.Time                          `json:"time"`
}

type deliveryOutcome string

const (
	deliveryOutcomeDispatched deliveryOutcome = "dispatched"
	deliveryOutcomeRetryable  deliveryOutcome = "retryable_failed"
	deliveryOutcomeTerminal   deliveryOutcome = "terminal_failed"
	deliveryOutcomeAccepted   deliveryOutcome = "delivered"
	deliveryOutcomeAmbiguous  deliveryOutcome = "ambiguous"
	deliveryOutcomeReplayed   deliveryOutcome = "replayed"
	deliveryOutcomeScheduled  deliveryOutcome = "scheduled"
	deliveryOutcomeSuppressed deliveryOutcome = "suppressed"
	deliveryOutcomeExhausted  deliveryOutcome = "stream_exhausted"
	deliveryOutcomeReconciled deliveryOutcome = "reconciled"
)

var deliveryLogWriter io.Writer = os.Stdout

func deliveryOutcomeLog(entry deliveryLog) {
	if deliveryLogWriter == nil {
		return
	}
	entry.TenantID = redactSecret(entry.TenantID)
	entry.TransitionID = redactSecret(entry.TransitionID)
	entry.DeliveryID = redactSecret(entry.DeliveryID)
	entry.ChannelType = redactSecret(entry.ChannelType)
	entry.Reason = redactSecret(entry.Reason)
	encoded, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = io.WriteString(deliveryLogWriter, string(encoded)+"\n")
}

func redactSecret(value string) string {
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, "\n\r\t") || strings.ContainsAny(value, "\"'`<>") {
		return ""
	}
	return value
}

func logDeliveryDispatched(tenantID, transitionID string, attempt int) {
	deliveryOutcomeLog(deliveryLog{
		Outcome:      string(deliveryOutcomeDispatched),
		Operation:    "stream-dispatch",
		SourceKind:   "dynamodb_stream",
		TenantID:     tenantID,
		TransitionID: transitionID,
		Attempt:      attempt,
		Time:         time.Now().UTC(),
	})
}

func logDeliveryResult(tenantID, transitionID, deliveryID string, stepNumber int, channelType string, attempt int, class notifications.DeliveryOutcomeClass) {
	var outcome string
	switch class {
	case notifications.OutcomeAccepted:
		outcome = string(deliveryOutcomeAccepted)
	case notifications.OutcomeRetryExhausted:
		outcome = string(deliveryOutcomeTerminal)
	default:
		if notifications.IsTerminalClass(class) {
			outcome = string(deliveryOutcomeTerminal)
		} else {
			outcome = string(deliveryOutcomeRetryable)
		}
	}
	deliveryOutcomeLog(deliveryLog{
		Outcome:      outcome,
		Operation:    "attempt-result",
		TenantID:     tenantID,
		TransitionID: transitionID,
		DeliveryID:   deliveryID,
		StepNumber:   stepNumber,
		ChannelType:  channelType,
		Attempt:      attempt,
		OutcomeClass: class,
		Time:         time.Now().UTC(),
	})
}

var _ = logDeliveryReplay

func logDeliveryReplay(tenantID, transitionID, deliveryID string) {
	deliveryOutcomeLog(deliveryLog{
		Outcome:      string(deliveryOutcomeReplayed),
		Operation:    "replay",
		TenantID:     tenantID,
		TransitionID: transitionID,
		DeliveryID:   deliveryID,
		Time:         time.Now().UTC(),
	})
}

func logScheduleDispatched(tenantID, transitionID string, stepNumber int) {
	deliveryOutcomeLog(deliveryLog{
		Outcome:      string(deliveryOutcomeScheduled),
		Operation:    "scheduler-target",
		SourceKind:   "scheduler_target",
		TenantID:     tenantID,
		TransitionID: transitionID,
		StepNumber:   stepNumber,
		Time:         time.Now().UTC(),
	})
}

var _ = logSuppression

func logSuppression(tenantID, incidentID, reason string) {
	deliveryOutcomeLog(deliveryLog{
		Outcome:    string(deliveryOutcomeSuppressed),
		Operation:  "suppress",
		TenantID:   tenantID,
		DeliveryID: incidentID,
		Reason:     reason,
		Time:       time.Now().UTC(),
	})
}

func logStreamExhausted(tenantID, transitionID, reason string) {
	deliveryOutcomeLog(deliveryLog{
		Outcome:      string(deliveryOutcomeExhausted),
		Operation:    "stream-dispatch",
		SourceKind:   "dynamodb_stream",
		TenantID:     tenantID,
		TransitionID: transitionID,
		Reason:       reason,
		Time:         time.Now().UTC(),
	})
}

func logReconciled(tenantID, transitionID string, bucket string) {
	deliveryOutcomeLog(deliveryLog{
		Outcome:      string(deliveryOutcomeReconciled),
		Operation:    "reconcile-recent",
		SourceKind:   "dynamodb_stream",
		TenantID:     tenantID,
		TransitionID: transitionID,
		Reason:       bucket,
		Time:         time.Now().UTC(),
	})
}
