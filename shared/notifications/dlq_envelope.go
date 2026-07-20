package notifications

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type DLQSourceKind string

const (
	DLQSourceStream   DLQSourceKind = "dynamodb_stream"
	DLQSourceSchedule DLQSourceKind = "scheduler_target"
	DLQSourceSQS      DLQSourceKind = "notification_sqs"
)

type DLQEnvelope struct {
	SourceKind    DLQSourceKind     `json:"sourceKind"`
	TenantID      string            `json:"tenantId"`
	TransitionID  string            `json:"transitionId,omitempty"`
	DeliveryID    string            `json:"deliveryId,omitempty"`
	StepNumber    int               `json:"stepNumber,omitempty"`
	Canonical     CanonicalEnvelope `json:"canonical,omitempty"`
	FailureReason string            `json:"failureReason,omitempty"`
	ObservedAt    string            `json:"observedAt"`
}

func (e DLQEnvelope) Validate() error {
	switch e.SourceKind {
	case DLQSourceStream, DLQSourceSchedule, DLQSourceSQS:
	default:
		return fmt.Errorf("unsupported dlq source kind %q", e.SourceKind)
	}
	if strings.TrimSpace(e.TenantID) == "" {
		return errors.New("dlq envelope tenantId is required")
	}
	if strings.TrimSpace(e.ObservedAt) == "" {
		return errors.New("dlq envelope observedAt is required")
	}
	return nil
}

func StreamExhaustionEnvelope(tenantID, transitionID, reason string, observedAt time.Time) (DLQEnvelope, error) {
	envelope := DLQEnvelope{
		SourceKind:    DLQSourceStream,
		TenantID:      tenantID,
		TransitionID:  transitionID,
		FailureReason: safeReason(reason),
		ObservedAt:    observedAt.UTC().Format(time.RFC3339),
	}
	if err := envelope.Validate(); err != nil {
		return DLQEnvelope{}, err
	}
	return envelope, nil
}

func SchedulerExhaustionEnvelope(tenantID, transitionID string, stepNumber int, reason string, observedAt time.Time) (DLQEnvelope, error) {
	envelope := DLQEnvelope{
		SourceKind:    DLQSourceSchedule,
		TenantID:      tenantID,
		TransitionID:  transitionID,
		StepNumber:    stepNumber,
		FailureReason: safeReason(reason),
		ObservedAt:    observedAt.UTC().Format(time.RFC3339),
	}
	if err := envelope.Validate(); err != nil {
		return DLQEnvelope{}, err
	}
	return envelope, nil
}

func SQSEnvelope(tenantID, transitionID, deliveryID string, reason string, observedAt time.Time) (DLQEnvelope, error) {
	envelope := DLQEnvelope{
		SourceKind:    DLQSourceSQS,
		TenantID:      tenantID,
		TransitionID:  transitionID,
		DeliveryID:    deliveryID,
		FailureReason: safeReason(reason),
		ObservedAt:    observedAt.UTC().Format(time.RFC3339),
	}
	if err := envelope.Validate(); err != nil {
		return DLQEnvelope{}, err
	}
	return envelope, nil
}

func (e DLQEnvelope) Marshal() (string, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func ParseDLQEnvelope(data string) (DLQEnvelope, error) {
	var envelope DLQEnvelope
	if err := json.Unmarshal([]byte(data), &envelope); err != nil {
		return DLQEnvelope{}, fmt.Errorf("parse dlq envelope: %w", err)
	}
	if err := envelope.Validate(); err != nil {
		return DLQEnvelope{}, err
	}
	return envelope, nil
}

func safeReason(reason string) string {
	trimmed := strings.TrimSpace(reason)
	if len(trimmed) > 240 {
		trimmed = trimmed[:240] + "..."
	}
	return trimmed
}
