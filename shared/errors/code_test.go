package errors

import "testing"

func TestRegistryCoversAllConstants(t *testing.T) {
	allCodes := []Code{
		CodeNotFound,
		CodeInvalidJSON,
		CodeValidationFailed,
		CodeImmutableField,
		CodeInlineChannelConfig,
		CodeServiceNotFound,
		CodeServiceAlreadyExists,
		CodeServiceActive,
		CodeServiceNotArchived,
		CodeServiceHasNoPolicy,
		CodeMonitorNotFound,
		CodeMonitorAlreadyExists,
		CodeMonitorDisabled,
		CodeMonitorStatusNotFound,
		CodeLastMonitor,
		CodeIncidentNotFound,
		CodeIncidentNotActionable,
		CodePolicyNotFound,
		CodePolicyReferenced,
		CodeChannelNotFound,
		CodeNotificationDelivery,
		CodeDeliveryNotFound,
		CodeDeliveryNotReplayable,
		CodeIdempotencyConflict,
		CodeAuthenticationRequired,
		CodeAuthorizationDenied,
		CodeInternal,
	}
	for _, code := range allCodes {
		if _, ok := registry[code]; !ok {
			t.Fatalf("registry missing entry for %s", code)
		}
	}
	if len(allCodes) != len(registry) {
		t.Fatalf("registry has %d entries but %d constants declared", len(registry), len(allCodes))
	}
}

func TestStatusOfUnknownPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("StatusOf(unknown) did not panic")
		}
	}()
	StatusOf(Code("UNKNOWN_CODE"))
}

func TestCodeWireIdentity(t *testing.T) {
	cases := map[Code]string{
		CodeNotFound:               "NOT_FOUND",
		CodeInvalidJSON:            "INVALID_JSON",
		CodeValidationFailed:       "VALIDATION_FAILED",
		CodeImmutableField:         "IMMUTABLE_FIELD",
		CodeInlineChannelConfig:    "INLINE_CHANNEL_CONFIG",
		CodeServiceNotFound:        "SERVICE_NOT_FOUND",
		CodeServiceAlreadyExists:   "SERVICE_ALREADY_EXISTS",
		CodeServiceActive:          "SERVICE_ACTIVE",
		CodeServiceNotArchived:     "SERVICE_NOT_ARCHIVED",
		CodeServiceHasNoPolicy:     "SERVICE_HAS_NO_POLICY",
		CodeMonitorNotFound:        "MONITOR_NOT_FOUND",
		CodeMonitorAlreadyExists:   "MONITOR_ALREADY_EXISTS",
		CodeMonitorDisabled:        "MONITOR_DISABLED",
		CodeMonitorStatusNotFound:  "MONITOR_STATUS_NOT_FOUND",
		CodeLastMonitor:            "LAST_MONITOR",
		CodeIncidentNotFound:       "INCIDENT_NOT_FOUND",
		CodeIncidentNotActionable:  "INCIDENT_NOT_ACTIONABLE",
		CodePolicyNotFound:         "POLICY_NOT_FOUND",
		CodePolicyReferenced:       "POLICY_REFERENCED",
		CodeChannelNotFound:        "CHANNEL_NOT_FOUND",
		CodeNotificationDelivery:   "NOTIFICATION_DELIVERY_FAILED",
		CodeDeliveryNotFound:       "DELIVERY_NOT_FOUND",
		CodeDeliveryNotReplayable:  "DELIVERY_NOT_REPLAYABLE",
		CodeIdempotencyConflict:    "IDEMPOTENCY_CONFLICT",
		CodeAuthenticationRequired: "AUTHENTICATION_REQUIRED",
		CodeAuthorizationDenied:    "AUTHORIZATION_DENIED",
		CodeInternal:               "INTERNAL",
	}
	for code, want := range cases {
		if string(code) != want {
			t.Fatalf("Code %s stringifies to %q, want %q", code, string(code), want)
		}
	}
}
