package dynamodbrecord

import (
	"time"

	"bolt-monitor/shared/dynamodbschema"
)

type AuditEventRecord struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType"`
	TenantID   string `dynamodbav:"TenantID"`
	ServiceID  string `dynamodbav:"ServiceID,omitempty"`
	MonitorID  string `dynamodbav:"MonitorID,omitempty"`
	AuditID    string `dynamodbav:"AuditID"`
	Action     string `dynamodbav:"Action"`
	ResourceID string `dynamodbav:"ResourceID"`
	Timestamp  string `dynamodbav:"Timestamp"`
	Actor      string `dynamodbav:"Actor,omitempty"`
	Origin     string `dynamodbav:"Origin,omitempty"`
	GSI3PK     string `dynamodbav:"GSI3PK,omitempty"`
	GSI3SK     string `dynamodbav:"GSI3SK,omitempty"`
}

func NewAuditEventRecord(now time.Time, auditID, tenantID, action, serviceID, monitorID string) AuditEventRecord {
	item := dynamodbschema.AuditEventItem(tenantID, auditID, now.UTC().Format(time.RFC3339))
	resourceItem := dynamodbschema.AuditResourceItem(tenantID, serviceID, monitorID, auditID, now.UTC().Format(time.RFC3339))
	return AuditEventRecord{
		PK:         item.PK,
		SK:         item.SK,
		EntityType: item.EntityType,
		TenantID:   item.TenantID,
		ServiceID:  dynamodbschema.NormalizeField(serviceID),
		MonitorID:  dynamodbschema.NormalizeField(monitorID),
		AuditID:    item.AuditID,
		Action:     action,
		ResourceID: monitorAuditResourceID(serviceID, monitorID),
		Timestamp:  now.UTC().Format(time.RFC3339),
		Origin:     "system",
		GSI3PK:     resourceItem.GSI3PK,
		GSI3SK:     resourceItem.GSI3SK,
	}
}

type AuditChangeRecord struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType"`
	AuditID    string `dynamodbav:"AuditID"`
	FieldPath  string `dynamodbav:"FieldPath"`
	OldValue   string `dynamodbav:"OldValue"`
	NewValue   string `dynamodbav:"NewValue"`
}

func NewAuditChangeRecord(auditID, fieldPath, oldValue, newValue string) AuditChangeRecord {
	item := dynamodbschema.AuditChangeItem(auditID, fieldPath)
	return AuditChangeRecord{
		PK:         item.PK,
		SK:         item.SK,
		EntityType: item.EntityType,
		AuditID:    item.AuditID,
		FieldPath:  dynamodbschema.NormalizeToken(fieldPath),
		OldValue:   oldValue,
		NewValue:   newValue,
	}
}

func monitorAuditResourceID(serviceID, monitorID string) string {
	return dynamodbschema.NormalizeToken(serviceID) + "/" + dynamodbschema.NormalizeToken(monitorID)
}

func MonitorAuditResourceID(serviceID, monitorID string) string {
	return monitorAuditResourceID(serviceID, monitorID)
}
