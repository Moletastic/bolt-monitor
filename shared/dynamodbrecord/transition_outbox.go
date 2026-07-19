package dynamodbrecord

import "bolt-monitor/shared/dynamodbschema"

const DispatchPending = "pending"

type TransitionOutboxRecord struct {
	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType"`
	TenantID string `dynamodbav:"TenantID"`
	EventID string `dynamodbav:"EventID"`
	TransitionID string `dynamodbav:"TransitionID"`
	ActivityID string `dynamodbav:"ActivityID"`
	RunID string `dynamodbav:"RunID"`
	IncidentID string `dynamodbav:"IncidentID"`
	TransitionType string `dynamodbav:"TransitionType"`
	ScheduleDefinitionVersion string `dynamodbav:"ScheduleDefinitionVersion"`
	ScheduledFor string `dynamodbav:"ScheduledFor"`
	DispatchStatus string `dynamodbav:"DispatchStatus"`
	CreatedAt string `dynamodbav:"CreatedAt"`
}

func NewTransitionOutboxRecord(tenantID, transitionID, runID, incidentID, transitionType, scheduleDefinitionVersion, scheduledFor, createdAt string) TransitionOutboxRecord {
	item := dynamodbschema.TransitionOutboxItem(tenantID, transitionID)
	return TransitionOutboxRecord{PK: item.PK, SK: item.SK, EntityType: item.EntityType, TenantID: item.TenantID, EventID: transitionID, TransitionID: transitionID, ActivityID: transitionID, RunID: runID, IncidentID: incidentID, TransitionType: transitionType, ScheduleDefinitionVersion: scheduleDefinitionVersion, ScheduledFor: scheduledFor, DispatchStatus: DispatchPending, CreatedAt: createdAt}
}
