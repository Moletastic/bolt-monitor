package domainvalues

import (
	"sort"
	"time"

	sharederrors "bolt-monitor/shared/errors"
)

// allowedIntervalSeconds is the closed set of supported monitor cadence
// values. Keeping it in this dependency-light package means domainworkflow
// code can validate intervals without importing monitorconfig.
var allowedIntervalSeconds = map[int]struct{}{
	60:   {},
	120:  {},
	180:  {},
	300:  {},
	600:  {},
	900:  {},
	1800: {},
	3600: {},
}

// IsAllowedIntervalSeconds reports whether the given interval is one of the
// supported cadence values.
func IsAllowedIntervalSeconds(seconds int) bool {
	_, ok := allowedIntervalSeconds[seconds]
	return ok
}

// AllowedIntervalSeconds returns the supported cadence values in ascending
// order. The order is stable so dashboard selectors see a canonical list.
func AllowedIntervalSeconds() []int {
	values := make([]int, 0, len(allowedIntervalSeconds))
	for value := range allowedIntervalSeconds {
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}

// CheckInterval is the validated monitor cadence value. Construction is
// restricted to the supported set so scheduler and worker code cannot
// receive an unsupported cadence.
type CheckInterval struct {
	seconds int
}

// NewCheckInterval validates that the input is a supported cadence value and
// returns the validated CheckInterval. Inputs outside the allowed set surface
// the existing VALIDATION_FAILED error contract.
func NewCheckInterval(seconds int) (CheckInterval, error) {
	if !IsAllowedIntervalSeconds(seconds) {
		return CheckInterval{}, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{
			"field":  "intervalSeconds",
			"reason": "must be one of the supported cadence values",
		})
	}
	return CheckInterval{seconds: seconds}, nil
}

// MustCheckInterval returns a CheckInterval without validation. Callers must
// have already verified that the input is in the supported set.
func MustCheckInterval(seconds int) CheckInterval {
	return CheckInterval{seconds: seconds}
}

// Seconds returns the original integer value for adapters and persistence.
func (c CheckInterval) Seconds() int { return c.seconds }

// Duration returns the equivalent time.Duration for scheduler calculations.
func (c CheckInterval) Duration() time.Duration {
	return time.Duration(c.seconds) * time.Second
}

// String returns the canonical duration representation for logs and
// serialization.
func (c CheckInterval) String() string {
	return (time.Duration(c.seconds) * time.Second).String()
}
