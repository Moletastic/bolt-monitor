package main

import (
	"context"
	"time"
)

type scheduleClient interface {
	ScheduleNextStep(ctx context.Context, event scheduledInvocationEvent, when time.Time) error
}

type scheduledInvocationEvent struct {
	IncidentID string `json:"incidentId"`
	Step       int    `json:"step"`
}
