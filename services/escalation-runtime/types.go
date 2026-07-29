package main

import (
	"context"
	"time"
)

type scheduleClient interface {
	ScheduleNextStep(ctx context.Context, event scheduledInvocationEvent, when time.Time) error
}

type scheduledInvocationEvent struct {
	TenantID   string `json:"tenantId"`
	IncidentID string `json:"incidentId"`
	Step       int    `json:"step"`
}
