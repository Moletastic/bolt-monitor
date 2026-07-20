package notifications

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	CanonicalEnvelopeVersion = "1"

	CanonicalKindTransition   = "transition"
	CanonicalKindScheduled    = "scheduled_step"
	CanonicalKindReplay       = "delivery_replay"
	CanonicalKindTestSend     = "channel_test"
	CanonicalSourceTransition = "transition"
	CanonicalSourceSchedule   = "scheduler_target"
	CanonicalSourceReplay     = "delivery_replay"
	CanonicalSourceTest       = "channel_test"
)

type CanonicalEnvelope struct {
	Version      string `json:"version"`
	Kind         string `json:"kind"`
	SourceKind   string `json:"sourceKind"`
	TenantID     string `json:"tenantId"`
	ServiceID    string `json:"serviceId,omitempty"`
	MonitorID    string `json:"monitorID,omitempty"`
	IncidentID   string `json:"incidentId,omitempty"`
	TransitionID string `json:"transitionId"`
	RunID        string `json:"runId,omitempty"`
	StepNumber   int    `json:"stepNumber,omitempty"`
	DeliveryID   string `json:"deliveryId,omitempty"`
	ScheduledAt  string `json:"scheduledAt,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

func (e CanonicalEnvelope) Validate() error {
	if strings.TrimSpace(e.Version) == "" {
		return errors.New("canonical envelope version is required")
	}
	if e.Version != CanonicalEnvelopeVersion {
		return fmt.Errorf("unsupported canonical envelope version %q", e.Version)
	}
	if strings.TrimSpace(e.TenantID) == "" {
		return errors.New("canonical envelope tenantId is required")
	}
	if strings.TrimSpace(e.TransitionID) == "" {
		return errors.New("canonical envelope transitionId is required")
	}
	switch e.Kind {
	case CanonicalKindTransition, CanonicalKindScheduled, CanonicalKindReplay, CanonicalKindTestSend:
	default:
		return fmt.Errorf("unsupported canonical envelope kind %q", e.Kind)
	}
	return nil
}

func CanonicalTransitionEnvelope(tenantID, eventID, sourceKind, createdAt string) CanonicalEnvelope {
	return CanonicalEnvelope{
		Version:      CanonicalEnvelopeVersion,
		Kind:         CanonicalKindTransition,
		SourceKind:   firstCanonicalSource(sourceKind, CanonicalSourceTransition),
		TenantID:     tenantID,
		TransitionID: eventID,
		RunID:        eventID,
		CreatedAt:    createdAt,
	}
}

func CanonicalScheduledStepEnvelope(tenantID, transitionID string, stepNumber int, sourceKind, createdAt string) CanonicalEnvelope {
	return CanonicalEnvelope{
		Version:      CanonicalEnvelopeVersion,
		Kind:         CanonicalKindScheduled,
		SourceKind:   firstCanonicalSource(sourceKind, CanonicalSourceSchedule),
		TenantID:     tenantID,
		TransitionID: transitionID,
		StepNumber:   stepNumber,
		CreatedAt:    createdAt,
	}
}

func CanonicalReplayEnvelope(tenantID, transitionID, deliveryID string, createdAt string) CanonicalEnvelope {
	return CanonicalEnvelope{
		Version:      CanonicalEnvelopeVersion,
		Kind:         CanonicalKindReplay,
		SourceKind:   CanonicalSourceReplay,
		TenantID:     tenantID,
		TransitionID: transitionID,
		DeliveryID:   deliveryID,
		CreatedAt:    createdAt,
	}
}

func CanonicalTestSendEnvelope(tenantID string, createdAt string) CanonicalEnvelope {
	return CanonicalEnvelope{
		Version:      CanonicalEnvelopeVersion,
		Kind:         CanonicalKindTestSend,
		SourceKind:   CanonicalSourceTest,
		TenantID:     tenantID,
		TransitionID: "test",
		CreatedAt:    createdAt,
	}
}

func firstCanonicalSource(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func DeliveryIdentity(tenantID, transitionID string, stepNumber int, channelID string) string {
	digest := sha256.Sum256([]byte(strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(tenantID)),
		strings.ToUpper(strings.TrimSpace(transitionID)),
		fmt.Sprintf("%d", stepNumber),
		strings.ToUpper(strings.TrimSpace(channelID)),
	}, "\n")))
	return "dlv_" + hex.EncodeToString(digest[:16])
}

func ReplayKeyFingerprint(tenantID, incidentID, deliveryID, idempotencyKey string) string {
	digest := sha256.Sum256([]byte(strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(tenantID)),
		strings.ToUpper(strings.TrimSpace(incidentID)),
		strings.ToUpper(strings.TrimSpace(deliveryID)),
		strings.TrimSpace(idempotencyKey),
	}, "\n")))
	return hex.EncodeToString(digest[:])
}
