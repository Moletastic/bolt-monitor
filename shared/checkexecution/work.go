package checkexecution

import "time"

type ExecutionWorkStatus string

type PublicationState string

const (
	PublicationPending      PublicationState = "pending"
	PublicationAcknowledged PublicationState = "acknowledged"
)

const (
	ExecutionWorkPending    ExecutionWorkStatus = "pending"
	ExecutionWorkInProgress ExecutionWorkStatus = "in_progress"
	ExecutionWorkCompleted  ExecutionWorkStatus = "completed"
	ExecutionWorkSkipped    ExecutionWorkStatus = "skipped"

	DefaultExecutionWorkRetentionDays = 7
)

type ExecutionWork struct {
	TenantID    string              `json:"tenantId"`
	ServiceID   string              `json:"serviceId"`
	MonitorID   string              `json:"monitorId"`
	RunID       string              `json:"runId"`
	Trigger     TriggerType         `json:"trigger"`
	AcceptedAt  time.Time           `json:"acceptedAt"`
	ScheduleDefinitionVersion string `json:"scheduleDefinitionVersion,omitempty"`
	ScheduledFor              *time.Time `json:"scheduledFor,omitempty"`
	RequestedAt time.Time           `json:"requestedAt"`
	Status      ExecutionWorkStatus `json:"status"`
	PublicationState PublicationState `json:"publicationState"`
	FencingToken     string           `json:"fencingToken,omitempty"`
	LeaseUntil       *time.Time       `json:"leaseUntil,omitempty"`
	AttemptCount     int              `json:"attemptCount"`
	TerminalReason   string           `json:"terminalReason,omitempty"`
	TransitionID     string           `json:"transitionId,omitempty"`
	StartedAt   *time.Time          `json:"startedAt,omitempty"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
	LastError   string              `json:"lastError,omitempty"`
}
