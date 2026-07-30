package main

import (
	"context"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
)

func TestExecuteIncidentCommandTransactsReservationWithIncidentRecords(t *testing.T) {
	incident := dynamodbrecord.IncidentRecord{TenantID: defaultTenantID, IncidentID: "INC_1", ServiceID: "SVC_1", MonitorID: "MON_1", Status: incidentStatusOpen, OpenedAt: "2026-07-22T09:00:00Z"}
	meta, err := sharedaws.MarshalMap(dynamodbrecord.NewIncidentMetaItemRecord(incident))
	if err != nil {
		t.Fatalf("marshal incident: %v", err)
	}
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{
		dynamodbschema.IncidentPK(incident.IncidentID) + "|META": meta,
	}}
	repo := newDynamoMonitorRepository(client, "table-name")
	now := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	record := newCommandIdempotencyRecord(defaultTenantID, "incident-acknowledge", incident.IncidentID, "incident-ack-123", "fingerprint", now, 24*time.Hour)

	stored, found, err := repo.ExecuteIncidentCommand(context.Background(), record, now)
	if err != nil || !found || stored.State != commandIdempotencyCompleted || stored.Response == "" {
		t.Fatalf("stored = %+v, found = %v, err = %v", stored, found, err)
	}
	if client.transactInput == nil || len(client.transactInput.TransactItems) != 7 {
		t.Fatalf("transaction items = %d, want 7", len(client.transactInput.TransactItems))
	}
	last := client.transactInput.TransactItems[len(client.transactInput.TransactItems)-1].Put
	if last == nil || last.ConditionExpression == nil || *last.ConditionExpression != "attribute_not_exists(PK) OR TTL <= :now" {
		t.Fatalf("idempotency transaction item = %+v", last)
	}
}
