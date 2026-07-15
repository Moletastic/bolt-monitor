package errors

import "net/http"

type Code string

const (
	CodeNotFound               Code = "NOT_FOUND"
	CodeInvalidJSON            Code = "INVALID_JSON"
	CodeValidationFailed       Code = "VALIDATION_FAILED"
	CodeImmutableField         Code = "IMMUTABLE_FIELD"
	CodeInlineChannelConfig    Code = "INLINE_CHANNEL_CONFIG"
	CodeServiceNotFound        Code = "SERVICE_NOT_FOUND"
	CodeServiceAlreadyExists   Code = "SERVICE_ALREADY_EXISTS"
	CodeServiceActive          Code = "SERVICE_ACTIVE"
	CodeServiceNotArchived     Code = "SERVICE_NOT_ARCHIVED"
	CodeServiceHasNoPolicy     Code = "SERVICE_HAS_NO_POLICY"
	CodeMonitorNotFound        Code = "MONITOR_NOT_FOUND"
	CodeMonitorAlreadyExists   Code = "MONITOR_ALREADY_EXISTS"
	CodeMonitorDisabled        Code = "MONITOR_DISABLED"
	CodeMonitorStatusNotFound  Code = "MONITOR_STATUS_NOT_FOUND"
	CodeLastMonitor            Code = "LAST_MONITOR"
	CodeIncidentNotFound       Code = "INCIDENT_NOT_FOUND"
	CodeIncidentNotActionable  Code = "INCIDENT_NOT_ACTIONABLE"
	CodePolicyNotFound         Code = "POLICY_NOT_FOUND"
	CodePolicyReferenced       Code = "POLICY_REFERENCED"
	CodeChannelNotFound        Code = "CHANNEL_NOT_FOUND"
	CodeNotificationDelivery   Code = "NOTIFICATION_DELIVERY_FAILED"
	CodeAuthenticationRequired Code = "AUTHENTICATION_REQUIRED"
	CodeAuthorizationDenied    Code = "AUTHORIZATION_DENIED"
	CodeInternal               Code = "INTERNAL"
)

type codeSpec struct {
	status int
}

var registry = map[Code]codeSpec{
	CodeNotFound:               {status: http.StatusNotFound},
	CodeInvalidJSON:            {status: http.StatusBadRequest},
	CodeValidationFailed:       {status: http.StatusBadRequest},
	CodeImmutableField:         {status: http.StatusBadRequest},
	CodeInlineChannelConfig:    {status: http.StatusBadRequest},
	CodeServiceNotFound:        {status: http.StatusNotFound},
	CodeServiceAlreadyExists:   {status: http.StatusConflict},
	CodeServiceActive:          {status: http.StatusConflict},
	CodeServiceNotArchived:     {status: http.StatusConflict},
	CodeServiceHasNoPolicy:     {status: http.StatusNotFound},
	CodeMonitorNotFound:        {status: http.StatusNotFound},
	CodeMonitorAlreadyExists:   {status: http.StatusConflict},
	CodeMonitorDisabled:        {status: http.StatusConflict},
	CodeMonitorStatusNotFound:  {status: http.StatusNotFound},
	CodeLastMonitor:            {status: http.StatusConflict},
	CodeIncidentNotFound:       {status: http.StatusNotFound},
	CodeIncidentNotActionable:  {status: http.StatusConflict},
	CodePolicyNotFound:         {status: http.StatusNotFound},
	CodePolicyReferenced:       {status: http.StatusConflict},
	CodeChannelNotFound:        {status: http.StatusNotFound},
	CodeNotificationDelivery:   {status: http.StatusBadGateway},
	CodeAuthenticationRequired: {status: http.StatusUnauthorized},
	CodeAuthorizationDenied:    {status: http.StatusForbidden},
	CodeInternal:               {status: http.StatusInternalServerError},
}

func StatusOf(code Code) int {
	spec, ok := registry[code]
	if !ok {
		panic("errors: unknown code " + string(code))
	}
	return spec.status
}
