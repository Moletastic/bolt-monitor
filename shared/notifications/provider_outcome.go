package notifications

import (
	"errors"
	"strconv"
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

// ParseRetryAfterSeconds accepts the raw Retry-After header value (seconds or
// HTTP-date) and returns a bounded positive integer. Returns 0 when absent.
func ParseRetryAfterSeconds(value string, maxSeconds int) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	if n, err := strconv.Atoi(trimmed); err == nil {
		if n < 0 {
			return 0
		}
		if maxSeconds > 0 && n > maxSeconds {
			return maxSeconds
		}
		return n
	}
	return 0
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

// ClassifyHTTPStatus maps an HTTP status code to a normalized outcome class
// using the documented retry policy: 2xx = accepted, 429 + 5xx = retryable,
// other 4xx = terminal, 1xx/3xx = unsupported.
func ClassifyHTTPStatus(statusCode int) DeliveryOutcomeClass {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return OutcomeAccepted
	case statusCode == 429:
		return OutcomeThrottled
	case statusCode >= 500 && statusCode < 600:
		return OutcomeProvider5xx
	case statusCode >= 400 && statusCode < 500:
		return OutcomeProvider4xx
	default:
		return OutcomeUnsupported
	}
}

func SafeProviderMetadata(statusCode int, requestID string, retryAfterSeconds int, maxRetryAfterSeconds int) ProviderMetadata {
	class := ""
	switch {
	case statusCode >= 200 && statusCode < 300:
		class = "2xx"
	case statusCode == 429:
		class = "429"
	case statusCode >= 500 && statusCode < 600:
		class = "5xx"
	case statusCode >= 400 && statusCode < 500:
		class = "4xx"
	}
	retry := retryAfterSeconds
	if retry < 0 {
		retry = 0
	}
	if maxRetryAfterSeconds > 0 && retry > maxRetryAfterSeconds {
		retry = maxRetryAfterSeconds
	}
	return ProviderMetadata{
		ProviderStatusClass: class,
		ProviderRequestID:   strings.TrimSpace(requestID),
		RetryAfterSeconds:   retry,
	}
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
