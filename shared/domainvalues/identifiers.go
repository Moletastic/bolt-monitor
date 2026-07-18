package domainvalues

import (
	"strings"

	sharederrors "bolt-monitor/shared/errors"
)

// normalizeToken upper-cases and trims identifiers that are stored as
// tenant or token scope (tenant IDs, run, incident, audit, policy, channel).
func normalizeToken(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

// normalizeField lower-cases and trims identifiers that are stored as field
// scope (service and monitor slugs).
func normalizeField(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// TenantID is the canonical upper-cased tenant identifier. It mirrors the
// auth.TenantID type to converge the two surfaces.
type TenantID string

// DefaultTenantID is the canonical single-tenant identifier deployed today.
const DefaultTenantID TenantID = "DEFAULT"

// NewTenantID validates blank input and returns the canonical upper-cased
// tenant identifier. Empty or whitespace-only input returns a typed
// VALIDATION_FAILED error.
func NewTenantID(value string) (TenantID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "tenantId", "reason": "required"})
	}
	return TenantID(normalizeToken(trimmed)), nil
}

// String returns the serialized tenant identifier.
func (t TenantID) String() string { return string(t) }

// ServiceID is the canonical lower-cased service slug.
type ServiceID string

// NewServiceID validates blank input and returns the canonical lower-cased
// service identifier.
func NewServiceID(value string) (ServiceID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "serviceId", "reason": "required"})
	}
	return ServiceID(normalizeField(trimmed)), nil
}

// String returns the serialized service identifier.
func (s ServiceID) String() string { return string(s) }

// MonitorID is the canonical lower-cased monitor slug.
type MonitorID string

// NewMonitorID validates blank input and returns the canonical lower-cased
// monitor identifier.
func NewMonitorID(value string) (MonitorID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "monitorId", "reason": "required"})
	}
	return MonitorID(normalizeField(trimmed)), nil
}

// String returns the serialized monitor identifier.
func (m MonitorID) String() string { return string(m) }

// IncidentID is the canonical upper-cased incident identifier.
type IncidentID string

// NewIncidentID validates blank input and returns the canonical upper-cased
// incident identifier.
func NewIncidentID(value string) (IncidentID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "incidentId", "reason": "required"})
	}
	return IncidentID(normalizeToken(trimmed)), nil
}

// String returns the serialized incident identifier.
func (i IncidentID) String() string { return string(i) }

// RunID is the canonical upper-cased execution run identifier.
type RunID string

// NewRunID validates blank input and returns the canonical upper-cased run
// identifier.
func NewRunID(value string) (RunID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "runId", "reason": "required"})
	}
	return RunID(normalizeToken(trimmed)), nil
}

// String returns the serialized run identifier.
func (r RunID) String() string { return string(r) }

// PolicyID is the canonical upper-cased escalation policy identifier.
type PolicyID string

// NewPolicyID validates blank input and returns the canonical upper-cased
// policy identifier.
func NewPolicyID(value string) (PolicyID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "policyId", "reason": "required"})
	}
	return PolicyID(normalizeToken(trimmed)), nil
}

// String returns the serialized policy identifier.
func (p PolicyID) String() string { return string(p) }

// ChannelID is the canonical upper-cased notification channel identifier.
type ChannelID string

// NewChannelID validates blank input and returns the canonical upper-cased
// channel identifier.
func NewChannelID(value string) (ChannelID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "channelId", "reason": "required"})
	}
	return ChannelID(normalizeToken(trimmed)), nil
}

// String returns the serialized channel identifier.
func (c ChannelID) String() string { return string(c) }
