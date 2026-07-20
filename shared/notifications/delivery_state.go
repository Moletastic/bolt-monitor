package notifications

import (
	"errors"
	"fmt"
	"strings"
)

type DeliveryState string

const (
	DeliveryPending        DeliveryState = "pending"
	DeliveryInFlight       DeliveryState = "in_flight"
	DeliveryRetryable      DeliveryState = "retryable_failed"
	DeliveryAmbiguous      DeliveryState = "ambiguous"
	DeliveryDelivered      DeliveryState = "delivered"
	DeliveryTerminalFailed DeliveryState = "terminal_failed"
)

func (s DeliveryState) Validate() error {
	switch s {
	case DeliveryPending, DeliveryInFlight, DeliveryRetryable, DeliveryAmbiguous, DeliveryDelivered, DeliveryTerminalFailed:
		return nil
	}
	return fmt.Errorf("unknown delivery state %q", s)
}

func (s DeliveryState) IsTerminal() bool {
	return s == DeliveryDelivered || s == DeliveryTerminalFailed
}

type DeliveryOutcomeClass string

const (
	OutcomeAccepted       DeliveryOutcomeClass = "accepted"
	OutcomeTimeout        DeliveryOutcomeClass = "timeout"
	OutcomeTransport      DeliveryOutcomeClass = "transport"
	OutcomeThrottled      DeliveryOutcomeClass = "throttled"
	OutcomeProvider5xx    DeliveryOutcomeClass = "provider_5xx"
	OutcomeProvider4xx    DeliveryOutcomeClass = "provider_4xx"
	OutcomeInvalidConfig  DeliveryOutcomeClass = "invalid_config"
	OutcomeUnsupported    DeliveryOutcomeClass = "unsupported_channel"
	OutcomeRetryExhausted DeliveryOutcomeClass = "retry_exhausted"
	OutcomeAbandonedLease DeliveryOutcomeClass = "abandoned_lease"
)

type ProviderMetadata struct {
	ProviderStatusClass string `json:"providerStatusClass,omitempty"`
	ProviderRequestID   string `json:"providerRequestId,omitempty"`
	RetryAfterSeconds   int    `json:"retryAfterSeconds,omitempty"`
}

type DeliveryRecord struct {
	TenantID         string               `json:"tenantId"`
	IncidentID       string               `json:"incidentId"`
	TransitionID     string               `json:"transitionId"`
	DeliveryID       string               `json:"deliveryId"`
	ChannelID        string               `json:"channelId"`
	ChannelType      string               `json:"channelType"`
	StepNumber       int                  `json:"stepNumber"`
	State            DeliveryState        `json:"state"`
	AttemptCount     int                  `json:"attemptCount"`
	LastAttemptAt    string               `json:"lastAttemptAt,omitempty"`
	LeaseUntil       string               `json:"leaseUntil,omitempty"`
	NextAttemptAt    string               `json:"nextAttemptAt,omitempty"`
	FencingToken     string               `json:"fencingToken,omitempty"`
	LastOutcomeClass DeliveryOutcomeClass `json:"lastOutcomeClass,omitempty"`
	ProviderMetadata ProviderMetadata     `json:"providerMetadata,omitempty"`
	CreatedAt        string               `json:"createdAt"`
	UpdatedAt        string               `json:"updatedAt"`
}

func (d DeliveryRecord) Validate() error {
	if strings.TrimSpace(d.TenantID) == "" {
		return errors.New("delivery tenantId is required")
	}
	if strings.TrimSpace(d.IncidentID) == "" {
		return errors.New("delivery incidentId is required")
	}
	if strings.TrimSpace(d.TransitionID) == "" {
		return errors.New("delivery transitionId is required")
	}
	if strings.TrimSpace(d.DeliveryID) == "" {
		return errors.New("delivery deliveryId is required")
	}
	if strings.TrimSpace(d.ChannelID) == "" {
		return errors.New("delivery channelId is required")
	}
	if d.StepNumber <= 0 {
		return errors.New("delivery stepNumber must be positive")
	}
	return d.State.Validate()
}

type EscalationPlan struct {
	TenantID     string   `json:"tenantId"`
	IncidentID   string   `json:"incidentId"`
	TransitionID string   `json:"transitionId"`
	PolicyID     string   `json:"policyId"`
	SelectedPath string   `json:"selectedPath"`
	StepNumbers  []int    `json:"stepNumbers"`
	StepChannels []string `json:"stepChannels"`
	CreatedAt    string   `json:"createdAt"`
}

func (p EscalationPlan) Validate() error {
	if strings.TrimSpace(p.TenantID) == "" {
		return errors.New("plan tenantId is required")
	}
	if strings.TrimSpace(p.IncidentID) == "" {
		return errors.New("plan incidentId is required")
	}
	if strings.TrimSpace(p.TransitionID) == "" {
		return errors.New("plan transitionId is required")
	}
	if len(p.StepNumbers) != len(p.StepChannels) {
		return errors.New("plan step numbers and channels must align")
	}
	return nil
}

type ReplayCommand struct {
	TenantID       string `json:"tenantId"`
	IncidentID     string `json:"incidentId"`
	TransitionID   string `json:"transitionId"`
	DeliveryID     string `json:"deliveryId"`
	IdempotencyKey string `json:"idempotencyKey"`
	RequestedAt    string `json:"requestedAt"`
	RequestedBy    string `json:"requestedBy,omitempty"`
}

type ReplayIdempotencyRecord struct {
	TenantID           string `json:"tenantId"`
	IncidentID         string `json:"incidentId"`
	DeliveryID         string `json:"deliveryId"`
	Operation          string `json:"operation"`
	IdempotencyKey     string `json:"idempotencyKey"`
	RequestFingerprint string `json:"requestFingerprint"`
	ResultDeliveryID   string `json:"resultDeliveryId"`
	CreatedAt          string `json:"createdAt"`
	ExpiresAt          int64  `json:"expiresAt"`
}
