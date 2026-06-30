package dynamodbrecord

import (
	"strings"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/monitorconfig"
)

type ExecutionWorkItemRecord struct {
	PK              string `dynamodbav:"PK"`
	SK              string `dynamodbav:"SK"`
	EntityType      string `dynamodbav:"EntityType"`
	TenantID        string `dynamodbav:"TenantID"`
	ServiceID       string `dynamodbav:"ServiceID"`
	MonitorID       string `dynamodbav:"MonitorID"`
	RunID           string `dynamodbav:"RunID"`
	ProbeLocationID string `dynamodbav:"ProbeLocationID"`
	Trigger         string `dynamodbav:"Trigger"`
	AcceptedAt      string `dynamodbav:"AcceptedAt"`
	Status          string `dynamodbav:"Status"`
	StartedAt       string `dynamodbav:"StartedAt,omitempty"`
	CompletedAt     string `dynamodbav:"CompletedAt,omitempty"`
	LastError       string `dynamodbav:"LastError,omitempty"`
}

func NewExecutionWorkItemRecord(tenantID, serviceID, monitorID, runID, probeLocationID string, trigger checkexecution.TriggerType, acceptedAt string, status checkexecution.ExecutionWorkStatus, startedAt, completedAt *time.Time, lastError string) ExecutionWorkItemRecord {
	item := dynamodbschema.ExecutionWorkItem(tenantID, acceptedAt, runID, probeLocationID)
	record := ExecutionWorkItemRecord{
		PK:              item.PK,
		SK:              item.SK,
		EntityType:      dynamodbschema.EntityExecutionWork,
		TenantID:        dynamodbschema.NormalizeToken(tenantID),
		ServiceID:       dynamodbschema.NormalizeField(serviceID),
		MonitorID:       dynamodbschema.NormalizeField(monitorID),
		RunID:           dynamodbschema.NormalizeToken(runID),
		ProbeLocationID: dynamodbschema.NormalizeToken(probeLocationID),
		Trigger:         string(trigger),
		AcceptedAt:      acceptedAt,
		Status:          string(status),
		LastError:       lastError,
	}
	if startedAt != nil {
		record.StartedAt = startedAt.UTC().Format(time.RFC3339)
	}
	if completedAt != nil {
		record.CompletedAt = completedAt.UTC().Format(time.RFC3339)
	}
	return record
}

func NewExecutionWorkItemRecords(monitor monitorconfig.Monitor, trigger checkexecution.TriggerType, runID, acceptedAt string) []ExecutionWorkItemRecord {
	works := make([]ExecutionWorkItemRecord, 0, len(monitor.ProbeLocations))
	for _, probeLocationID := range monitor.ProbeLocations {
		item := dynamodbschema.ExecutionWorkItem(monitor.TenantID, acceptedAt, runID, probeLocationID)
		works = append(works, ExecutionWorkItemRecord{
			PK:              item.PK,
			SK:              item.SK,
			EntityType:      dynamodbschema.EntityExecutionWork,
			TenantID:        dynamodbschema.NormalizeToken(monitor.TenantID),
			ServiceID:       dynamodbschema.NormalizeField(monitor.ServiceID),
			MonitorID:       dynamodbschema.NormalizeField(monitor.MonitorID),
			RunID:           dynamodbschema.NormalizeToken(runID),
			ProbeLocationID: dynamodbschema.NormalizeToken(probeLocationID),
			Trigger:         string(trigger),
			AcceptedAt:      acceptedAt,
			Status:          string(checkexecution.ExecutionWorkPending),
		})
	}
	return works
}

func ExecutionWorkItemRecordFromWork(work checkexecution.ExecutionWork) ExecutionWorkItemRecord {
	return NewExecutionWorkItemRecord(work.TenantID, work.ServiceID, work.MonitorID, work.RunID, work.ProbeLocationID, work.Trigger, work.RequestedAt.UTC().Format(time.RFC3339), work.Status, work.StartedAt, work.CompletedAt, work.LastError)
}

func (r ExecutionWorkItemRecord) ToWork() (checkexecution.ExecutionWork, error) {
	requestedAt, err := time.Parse(time.RFC3339, r.AcceptedAt)
	if err != nil {
		return checkexecution.ExecutionWork{}, err
	}
	work := checkexecution.ExecutionWork{
		TenantID:        r.TenantID,
		ServiceID:       r.ServiceID,
		MonitorID:       r.MonitorID,
		RunID:           r.RunID,
		ProbeLocationID: r.ProbeLocationID,
		Trigger:         checkexecution.TriggerType(strings.ToLower(r.Trigger)),
		RequestedAt:     requestedAt,
		Status:          checkexecution.ExecutionWorkStatus(strings.ToLower(r.Status)),
		LastError:       r.LastError,
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
