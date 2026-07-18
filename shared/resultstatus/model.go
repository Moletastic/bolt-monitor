package resultstatus

import (
	"fmt"
	"strings"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
)

const DefaultCheckRunRetentionDays = 30

type MonitorState string

const (
	MonitorStateUp          MonitorState = "UP"
	MonitorStateDegraded    MonitorState = "DEGRADED"
	MonitorStateDown        MonitorState = "DOWN"
	MonitorStateRecovering  MonitorState = "RECOVERING"
	MonitorStateMaintenance MonitorState = "MAINTENANCE"
)

type CheckRun struct {
	ServiceID   string                     `json:"serviceId"`
	MonitorID   string                     `json:"monitorId"`
	TenantID    string                     `json:"tenantId"`
	RunID       string                     `json:"runId"`
	Type        string                     `json:"type"`
	Trigger     checkexecution.TriggerType `json:"trigger"`
	StartedAt   time.Time                  `json:"startedAt"`
	FinishedAt  time.Time                  `json:"finishedAt"`
	DurationMs  int64                      `json:"durationMs"`
	Outcome     checkexecution.Outcome     `json:"outcome"`
	StatusCode  *int                       `json:"statusCode,omitempty"`
	Error       string                     `json:"error,omitempty"`
	FailureCode string                     `json:"failureCode,omitempty"`
	TTL         int64                      `json:"ttl"`
}

type MonitorStatus struct {
	ServiceID            string                 `json:"serviceId"`
	MonitorID            string                 `json:"monitorId"`
	TenantID             string                 `json:"tenantId"`
	CurrentStatus        string                 `json:"currentStatus"`
	ConsecutiveFailures  int                    `json:"consecutiveFailures"`
	ConsecutiveSuccesses int                    `json:"consecutiveSuccesses"`
	LastCheckedAt        time.Time              `json:"lastCheckedAt"`
	LastDurationMs       int64                  `json:"lastDurationMs"`
	LastError            string                 `json:"lastError,omitempty"`
	LastFailureCode      string                 `json:"lastFailureCode,omitempty"`
	LastOutcome          checkexecution.Outcome `json:"lastOutcome"`
}

type CheckRunRecord struct {
	PK          string `json:"pk"`
	SK          string `json:"sk"`
	EntityType  string `json:"entityType"`
	TenantID    string `json:"tenantId"`
	ServiceID   string `json:"serviceId"`
	MonitorID   string `json:"monitorId"`
	RunID       string `json:"runId"`
	Type        string `json:"type"`
	Trigger     string `json:"trigger"`
	StartedAt   string `json:"startedAt"`
	FinishedAt  string `json:"finishedAt"`
	DurationMs  int64  `json:"durationMs"`
	Outcome     string `json:"outcome"`
	StatusCode  *int   `json:"statusCode,omitempty"`
	Error       string `json:"error,omitempty"`
	FailureCode string `json:"failureCode,omitempty"`
	TTL         int64  `json:"ttl"`
}

type MonitorStatusRecord struct {
	PK                   string `json:"pk" dynamodbav:"PK"`
	SK                   string `json:"sk" dynamodbav:"SK"`
	EntityType           string `json:"entityType" dynamodbav:"EntityType"`
	TenantID             string `json:"tenantId" dynamodbav:"TenantID"`
	ServiceID            string `json:"serviceId" dynamodbav:"ServiceID"`
	MonitorID            string `json:"monitorId" dynamodbav:"MonitorID"`
	CurrentStatus        string `json:"currentStatus" dynamodbav:"CurrentStatus"`
	ConsecutiveFailures  int    `json:"consecutiveFailures" dynamodbav:"ConsecutiveFailures"`
	ConsecutiveSuccesses int    `json:"consecutiveSuccesses" dynamodbav:"ConsecutiveSuccesses"`
	LastCheckedAt        string `json:"lastCheckedAt" dynamodbav:"LastCheckedAt"`
	UpdatedAt            string `json:"updatedAt" dynamodbav:"UpdatedAt,omitempty"`
	LastDurationMs       int64  `json:"lastDurationMs" dynamodbav:"LastDurationMs"`
	LastError            string `json:"lastError,omitempty" dynamodbav:"LastError,omitempty"`
	LastFailureCode      string `json:"lastFailureCode,omitempty" dynamodbav:"LastFailureCode,omitempty"`
	LastOutcome          string `json:"lastOutcome" dynamodbav:"LastOutcome"`
	GSI2PK               string `json:"gsi2pk,omitempty" dynamodbav:"GSI2PK,omitempty"`
	GSI2SK               string `json:"gsi2sk,omitempty" dynamodbav:"GSI2SK,omitempty"`
}

func NewCheckRun(result checkexecution.ExecutionResult, now time.Time) CheckRun {
	runID := strings.TrimSpace(result.RunID)
	if runID == "" {
		runID = fmt.Sprintf("RUN_%s", strings.TrimPrefix(result.StartedAt.UTC().Format("20060102T150405.000000000"), ""))
	}
	return CheckRun{
		ServiceID:   strings.ToLower(result.ServiceID),
		MonitorID:   strings.ToLower(result.MonitorID),
		TenantID:    strings.ToUpper(result.TenantID),
		RunID:       strings.ToUpper(runID),
		Type:        result.Type,
		Trigger:     result.Trigger,
		StartedAt:   result.StartedAt.UTC(),
		FinishedAt:  result.FinishedAt.UTC(),
		DurationMs:  result.DurationMs,
		Outcome:     result.Outcome,
		StatusCode:  result.StatusCode,
		Error:       result.Error,
		FailureCode: result.FailureCode,
		TTL:         now.UTC().Add(DefaultCheckRunRetentionDays * 24 * time.Hour).Unix(),
	}
}

func NewMonitorStatus(result checkexecution.ExecutionResult) MonitorStatus {
	status := "DOWN"
	if result.Outcome == checkexecution.OutcomeSuccess {
		status = "UP"
	}
	return MonitorStatus{
		ServiceID:            strings.ToLower(result.ServiceID),
		MonitorID:            strings.ToLower(result.MonitorID),
		TenantID:             strings.ToUpper(result.TenantID),
		CurrentStatus:        status,
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 0,
		LastCheckedAt:        result.FinishedAt.UTC(),
		LastDurationMs:       result.DurationMs,
		LastError:            result.Error,
		LastFailureCode:      result.FailureCode,
		LastOutcome:          result.Outcome,
	}
}

type ThresholdConfig struct {
	FailureThreshold  int
	RecoveryThreshold int
}

func NewMonitorStatusWithConfig(result checkexecution.ExecutionResult, config ThresholdConfig) MonitorStatus {
	failureThreshold := config.FailureThreshold
	if failureThreshold < 1 {
		failureThreshold = 1
	}

	var status string
	var consecutiveFailures int
	var consecutiveSuccesses int

	if result.Outcome == checkexecution.OutcomeSuccess {
		status = string(MonitorStateUp)
		consecutiveSuccesses = 1
	} else {
		consecutiveFailures = 1
		if consecutiveFailures >= failureThreshold {
			status = string(MonitorStateDown)
		} else {
			status = string(MonitorStateDegraded)
		}
	}

	return MonitorStatus{
		ServiceID:            strings.ToLower(result.ServiceID),
		MonitorID:            strings.ToLower(result.MonitorID),
		TenantID:             strings.ToUpper(result.TenantID),
		CurrentStatus:        status,
		ConsecutiveFailures:  consecutiveFailures,
		ConsecutiveSuccesses: consecutiveSuccesses,
		LastCheckedAt:        result.FinishedAt.UTC(),
		LastDurationMs:       result.DurationMs,
		LastError:            result.Error,
		LastFailureCode:      result.FailureCode,
		LastOutcome:          result.Outcome,
	}
}

func (r CheckRun) ToRecord() CheckRunRecord {
	item := dynamodbschema.CheckRunItem(r.TenantID, r.ServiceID, r.MonitorID, r.StartedAt.UTC().Format(time.RFC3339), r.RunID, r.TTL)
	return CheckRunRecord{
		PK:          item.PK,
		SK:          item.SK,
		EntityType:  item.EntityType,
		TenantID:    r.TenantID,
		ServiceID:   r.ServiceID,
		MonitorID:   r.MonitorID,
		RunID:       r.RunID,
		Type:        r.Type,
		Trigger:     string(r.Trigger),
		StartedAt:   r.StartedAt.UTC().Format(time.RFC3339),
		FinishedAt:  r.FinishedAt.UTC().Format(time.RFC3339),
		DurationMs:  r.DurationMs,
		Outcome:     string(r.Outcome),
		StatusCode:  r.StatusCode,
		Error:       r.Error,
		FailureCode: r.FailureCode,
		TTL:         r.TTL,
	}
}

func (s MonitorStatus) ToRecord() MonitorStatusRecord {
	item := dynamodbschema.MonitorStatusItem(s.TenantID, s.ServiceID, s.MonitorID, s.CurrentStatus, s.LastCheckedAt.UTC().Format(time.RFC3339))
	return MonitorStatusRecord{
		PK:                   item.PK,
		SK:                   item.SK,
		EntityType:           item.EntityType,
		TenantID:             s.TenantID,
		ServiceID:            s.ServiceID,
		MonitorID:            s.MonitorID,
		CurrentStatus:        s.CurrentStatus,
		ConsecutiveFailures:  s.ConsecutiveFailures,
		ConsecutiveSuccesses: s.ConsecutiveSuccesses,
		LastCheckedAt:        s.LastCheckedAt.UTC().Format(time.RFC3339),
		UpdatedAt:            s.LastCheckedAt.UTC().Format(time.RFC3339),
		LastDurationMs:       s.LastDurationMs,
		LastError:            s.LastError,
		LastFailureCode:      s.LastFailureCode,
		LastOutcome:          string(s.LastOutcome),
	}
}
