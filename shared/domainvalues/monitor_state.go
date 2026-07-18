package domainvalues

import (
	"strings"

	sharederrors "bolt-monitor/shared/errors"
)

// MonitorState is the domain representation of current monitor state. Storage
// and API serialization live on the adapter boundaries.
type MonitorState string

// Canonical monitor states. Persisted items use the UPPERCASE form. API
// adapters return the lower-case form to keep the public response envelope
// stable.
const (
	MonitorStateUp          MonitorState = "UP"
	MonitorStateDegraded    MonitorState = "DEGRADED"
	MonitorStateDown        MonitorState = "DOWN"
	MonitorStateRecovering  MonitorState = "RECOVERING"
	MonitorStateMaintenance MonitorState = "MAINTENANCE"
	MonitorStateUnknown     MonitorState = "UNKNOWN"
)

// allowedStates is the closed set of canonical monitor states.
var allowedStates = map[MonitorState]struct{}{
	MonitorStateUp:          {},
	MonitorStateDegraded:    {},
	MonitorStateDown:        {},
	MonitorStateRecovering:  {},
	MonitorStateMaintenance: {},
	MonitorStateUnknown:     {},
}

// IsMonitorState reports whether the input string maps to a canonical monitor
// state. The check normalizes the casing to match the persisted form before
// testing membership.
func IsMonitorState(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	_, ok := allowedStates[MonitorState(normalizeToken(value))]
	return ok
}

// NewMonitorState validates that the input maps to a canonical monitor state
// and returns the canonical UPPERCASE value.
func NewMonitorState(value string) (MonitorState, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "monitorState", "reason": "required"})
	}
	upper := MonitorState(normalizeToken(trimmed))
	if _, ok := allowedStates[upper]; !ok {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "monitorState", "reason": "unsupported monitor state"})
	}
	return upper, nil
}

// MonitorStateFromStored normalizes a stored value into a canonical
// MonitorState without validation. The adapter accepts legacy-cased items
// already in the table.
func MonitorStateFromStored(stored string) MonitorState {
	return MonitorState(normalizeToken(stored))
}

// MonitorStateFromAPI normalizes an API-supplied state into a canonical
// MonitorState without validation. Use NewMonitorState when validation is
// required.
func MonitorStateFromAPI(api string) MonitorState {
	return MonitorState(normalizeToken(api))
}

// Stored returns the UPPERCASE persisted form.
func (s MonitorState) Stored() string { return string(s) }

// API returns the lower-case serialized form for HTTP responses.
func (s MonitorState) API() string { return strings.ToLower(string(s)) }

// String returns the canonical stored form so log lines and error messages
// are unambiguous.
func (s MonitorState) String() string { return string(s) }

// AllowedMonitorStates returns the sorted list of canonical monitor states
// for introspection and dashboard selectors.
func AllowedMonitorStates() []MonitorState {
	out := make([]MonitorState, 0, len(allowedStates))
	for state := range allowedStates {
		out = append(out, state)
	}
	// Stable order: U, D, D, R, M, U
	sortMonitorStates(out)
	return out
}

func sortMonitorStates(states []MonitorState) {
	for i := 1; i < len(states); i++ {
		for j := i; j > 0 && states[j-1] > states[j]; j-- {
			states[j-1], states[j] = states[j], states[j-1]
		}
	}
}
