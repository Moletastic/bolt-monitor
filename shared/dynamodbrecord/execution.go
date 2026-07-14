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
	Status      string `dynamodbav:"Status"`
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
	return NewExecutionWorkItemRecord(work.TenantID, work.ServiceID, work.MonitorID, work.RunID, work.Trigger, work.RequestedAt.UTC().Format(time.RFC3339), work.Status, work.StartedAt, work.CompletedAt, work.LastError)
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
		Status:      checkexecution.ExecutionWorkStatus(strings.ToLower(r.Status)),
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
	return work, nil
}
