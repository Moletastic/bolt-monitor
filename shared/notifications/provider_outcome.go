package notifications

import (
	"errors"
	"strings"
)

type SendOutcome struct {
	Class        DeliveryOutcomeClass `json:"class"`
	Retryable    bool                 `json:"retryable"`
	Metadata     ProviderMetadata     `json:"metadata"`
	ProviderName string               `json:"providerName,omitempty"`
}

func (o SendOutcome) Validate() error {
	switch o.Class {
	case OutcomeAccepted, OutcomeTimeout, OutcomeTransport, OutcomeThrottled,
		OutcomeProvider5xx, OutcomeProvider4xx, OutcomeInvalidConfig, OutcomeUnsupported:
		return nil
	}
	return errors.New("send outcome class is required")
}

func IsRetryableClass(class DeliveryOutcomeClass) bool {
	switch class {
	case OutcomeTimeout, OutcomeTransport, OutcomeThrottled, OutcomeProvider5xx:
		return true
	}
	return false
}

func IsTerminalClass(class DeliveryOutcomeClass) bool {
	switch class {
	case OutcomeInvalidConfig, OutcomeProvider4xx, OutcomeUnsupported, OutcomeRetryExhausted:
		return true
	}
	return false
}

type Provider interface {
	Send(ctx ProviderContext, deliveryID string) (SendOutcome, error)
	ChannelType() string
}

type ProviderContext struct {
	TenantID     string
	IncidentID   string
	TransitionID string
	DeliveryID   string
	Config       []byte
	Subject      string
	Message      string
}

func ParseDeliveryState(value string) (DeliveryState, error) {
	state := DeliveryState(strings.TrimSpace(value))
	if err := state.Validate(); err != nil {
		return "", err
	}
	return state, nil
}
