package dynamodbrecord

import (
	"strings"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
)

type ExecutionWorkItemRecord struct {
	PK          string `dynamodbav:"PK"`
	SK          string `dynamodbav:"SK"`
	EntityType  string `dynamodbav:"EntityType"`
	TTL         int64  `dynamodbav:"TTL"`
	TenantID    string `dynamodbav:"TenantID"`
	ServiceID   string `dynamodbav:"ServiceID"`
	MonitorID   string `dynamodbav:"MonitorID"`
	RunID       string `dynamodbav:"RunID"`
	Trigger     string `dynamodbav:"Trigger"`
	AcceptedAt  string `dynamodbav:"AcceptedAt"`
	ScheduleDefinitionVersion string `dynamodbav:"ScheduleDefinitionVersion,omitempty"`
	ScheduledFor              string `dynamodbav:"ScheduledFor,omitempty"`
	Status      string `dynamodbav:"Status"`
	PublicationState string `dynamodbav:"PublicationState"`
	FencingToken     string `dynamodbav:"FencingToken,omitempty"`
	LeaseUntil       string `dynamodbav:"LeaseUntil,omitempty"`
	AttemptCount     int    `dynamodbav:"AttemptCount"`
	TerminalReason   string `dynamodbav:"TerminalReason,omitempty"`
	TransitionID     string `dynamodbav:"TransitionID,omitempty"`
	StartedAt   string `dynamodbav:"StartedAt,omitempty"`
	CompletedAt string `dynamodbav:"CompletedAt,omitempty"`
	LastError   string `dynamodbav:"LastError,omitempty"`
}

func NewExecutionWorkItemRecord(tenantID, serviceID, monitorID, runID string, trigger checkexecution.TriggerType, acceptedAt string, status checkexecution.ExecutionWorkStatus, startedAt, completedAt *time.Time, lastError string) ExecutionWorkItemRecord {
	ttl := executionWorkTTL(acceptedAt)
	item := dynamodbschema.ExecutionWorkItem(tenantID, acceptedAt, runID, ttl)
	record := ExecutionWorkItemRecord{
		PK:         item.PK,
		SK:         item.SK,
		EntityType: dynamodbschema.EntityExecutionWork,
		TTL:        ttl,
		TenantID:   dynamodbschema.NormalizeToken(tenantID),
		ServiceID:  dynamodbschema.NormalizeField(serviceID),
		MonitorID:  dynamodbschema.NormalizeField(monitorID),
		RunID:      dynamodbschema.NormalizeToken(runID),
		Trigger:    string(trigger),
		AcceptedAt: acceptedAt,
		Status:     string(status),
		PublicationState: string(checkexecution.PublicationPending),
		LastError:  lastError,
	}
	if startedAt != nil {
		record.StartedAt = startedAt.UTC().Format(time.RFC3339)
	}
	if completedAt != nil {
		record.CompletedAt = completedAt.UTC().Format(time.RFC3339)
	}
	return record
}

func ExecutionWorkItemRecordFromWork(work checkexecution.ExecutionWork) ExecutionWorkItemRecord {
	acceptedAt := work.AcceptedAt
	if acceptedAt.IsZero() {
		acceptedAt = work.RequestedAt
	}
	record := NewExecutionWorkItemRecord(work.TenantID, work.ServiceID, work.MonitorID, work.RunID, work.Trigger, acceptedAt.UTC().Format(time.RFC3339), work.Status, work.StartedAt, work.CompletedAt, work.LastError)
	record.ScheduleDefinitionVersion = work.ScheduleDefinitionVersion
	if work.ScheduledFor != nil {
		record.ScheduledFor = work.ScheduledFor.UTC().Format(time.RFC3339)
	}
	record.PublicationState = string(work.PublicationState)
	if record.PublicationState == "" {
		record.PublicationState = string(checkexecution.PublicationPending)
	}
	record.FencingToken = work.FencingToken
	if work.LeaseUntil != nil {
		record.LeaseUntil = work.LeaseUntil.UTC().Format(time.RFC3339)
	}
	record.AttemptCount = work.AttemptCount
	record.TerminalReason = work.TerminalReason
	record.TransitionID = work.TransitionID
	return record
}

func executionWorkTTL(acceptedAt string) int64 {
	parsed, err := time.Parse(time.RFC3339, acceptedAt)
	if err != nil {
		return 0
	}
	return parsed.UTC().Add(checkexecution.DefaultExecutionWorkRetentionDays * 24 * time.Hour).Unix()
}

func (r ExecutionWorkItemRecord) ToWork() (checkexecution.ExecutionWork, error) {
	requestedAt, err := time.Parse(time.RFC3339, r.AcceptedAt)
	if err != nil {
		return checkexecution.ExecutionWork{}, err
	}
	work := checkexecution.ExecutionWork{
		TenantID:    r.TenantID,
		ServiceID:   r.ServiceID,
		MonitorID:   r.MonitorID,
		RunID:       r.RunID,
		Trigger:     checkexecution.TriggerType(strings.ToLower(r.Trigger)),
		RequestedAt: requestedAt,
		AcceptedAt: requestedAt,
		Status:      checkexecution.ExecutionWorkStatus(strings.ToLower(r.Status)),
		PublicationState: checkexecution.PublicationState(strings.ToLower(r.PublicationState)),
		FencingToken: r.FencingToken,
		AttemptCount: r.AttemptCount,
		TerminalReason: r.TerminalReason,
		TransitionID: r.TransitionID,
		LastError:   r.LastError,
	}
	if strings.TrimSpace(r.StartedAt) != "" {
		startedAt, err := time.Parse(time.RFC3339, r.StartedAt)
		if err != nil {
			return checkexecution.ExecutionWork{}, err
		}
		work.StartedAt = &startedAt
	}
	if strings.TrimSpace(r.CompletedAt) != "" {
		completedAt, err := time.Parse(time.RFC3339, r.CompletedAt)
		if err != nil {
			return checkexecution.ExecutionWork{}, err
		}
		work.CompletedAt = &completedAt
	}
	if strings.TrimSpace(r.ScheduledFor) != "" {
		scheduledFor, err := time.Parse(time.RFC3339, r.ScheduledFor)
		if err != nil {
			return checkexecution.ExecutionWork{}, err
		}
		work.ScheduleDefinitionVersion = r.ScheduleDefinitionVersion
		work.ScheduledFor = &scheduledFor
	}
	if strings.TrimSpace(r.LeaseUntil) != "" {
		leaseUntil, err := time.Parse(time.RFC3339, r.LeaseUntil)
		if err != nil {
			return checkexecution.ExecutionWork{}, err
		}
		work.LeaseUntil = &leaseUntil
	}
	return work, nil
}
