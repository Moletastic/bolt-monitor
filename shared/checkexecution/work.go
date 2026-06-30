package checkexecution

import "time"

type ExecutionWorkStatus string

const (
	ExecutionWorkPending    ExecutionWorkStatus = "pending"
	ExecutionWorkInProgress ExecutionWorkStatus = "in_progress"
	ExecutionWorkCompleted  ExecutionWorkStatus = "completed"
	ExecutionWorkSkipped    ExecutionWorkStatus = "skipped"
)

type ExecutionWork struct {
	TenantID        string              `json:"tenantId"`
	ServiceID       string              `json:"serviceId"`
	MonitorID       string              `json:"monitorId"`
	RunID           string              `json:"runId"`
	ProbeLocationID string              `json:"probeLocationId"`
	Trigger         TriggerType         `json:"trigger"`
	RequestedAt     time.Time           `json:"requestedAt"`
	Status          ExecutionWorkStatus `json:"status"`
	StartedAt       *time.Time          `json:"startedAt,omitempty"`
	CompletedAt     *time.Time          `json:"completedAt,omitempty"`
	LastError       string              `json:"lastError,omitempty"`
}
