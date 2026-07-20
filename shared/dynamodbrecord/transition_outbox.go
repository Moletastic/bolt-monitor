package dynamodbrecord

import "bolt-monitor/shared/dynamodbschema"

const DispatchPending = "pending"

type TransitionOutboxRecord struct {
	PK                        string `dynamodbav:"PK"`
	SK                        string `dynamodbav:"SK"`
	EntityType                string `dynamodbav:"EntityType"`
	TenantID                  string `dynamodbav:"TenantID"`
	EventID                   string `dynamodbav:"EventID"`
	TransitionID              string `dynamodbav:"TransitionID"`
	ActivityID                string `dynamodbav:"ActivityID"`
	RunID                     string `dynamodbav:"RunID"`
	ServiceID                 string `dynamodbav:"ServiceID"`
	MonitorID                 string `dynamodbav:"MonitorID"`
	IncidentID                string `dynamodbav:"IncidentID"`
	TransitionType            string `dynamodbav:"TransitionType"`
	ScheduleDefinitionVersion string `dynamodbav:"ScheduleDefinitionVersion"`
	ScheduledFor              string `dynamodbav:"ScheduledFor"`
	DispatchStatus            string `dynamodbav:"DispatchStatus"`
	Version                   string `dynamodbav:"Version,omitempty"`
	Kind                      string `dynamodbav:"Kind,omitempty"`
	SourceKind                string `dynamodbav:"SourceKind,omitempty"`
	DeliveryID                string `dynamodbav:"DeliveryID,omitempty"`
	StepNumber                int    `dynamodbav:"StepNumber,omitempty"`
	CreatedAt                 string `dynamodbav:"CreatedAt"`
}

func NewTransitionOutboxRecord(tenantID, serviceID, monitorID, transitionID, runID, incidentID, transitionType, scheduleDefinitionVersion, scheduledFor, createdAt string) TransitionOutboxRecord {
	item := dynamodbschema.TransitionOutboxItem(tenantID, transitionID)
	return TransitionOutboxRecord{PK: item.PK, SK: item.SK, EntityType: item.EntityType, TenantID: item.TenantID, EventID: transitionID, TransitionID: transitionID, ActivityID: transitionID, RunID: runID, ServiceID: serviceID, MonitorID: monitorID, IncidentID: incidentID, TransitionType: transitionType, ScheduleDefinitionVersion: scheduleDefinitionVersion, ScheduledFor: scheduledFor, DispatchStatus: DispatchPending, CreatedAt: createdAt}
}
