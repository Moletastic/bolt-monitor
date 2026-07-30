package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
)

func (r *dynamoMonitorRepository) ListIncidents(ctx context.Context, tenantID, status string) ([]dynamodbrecord.IncidentRecord, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "INCIDENT#")
	if err != nil {
		return nil, err
	}
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.IncidentItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != entityIncidentRef {
			continue
		}
		incident := record.ToIncident()
		if matchesIncidentFilter(incident.Status, status) {
			incidents = append(incidents, incident)
		}
	}
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].UpdatedAt > incidents[j].UpdatedAt })
	return incidents, nil
}

func (r *dynamoMonitorRepository) GetIncident(ctx context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	if len(out.Item) == 0 {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	var record dynamodbrecord.IncidentItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return dynamodbrecord.IncidentRecord{}, false, err
	}
	incident := record.ToIncident()
	if !strings.EqualFold(incident.TenantID, tenantID) {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	return incident, true, nil
}

func (r *dynamoMonitorRepository) ListIncidentActivities(ctx context.Context, tenantID, incidentID string) ([]dynamodbrecord.IncidentActivityRecord, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.IncidentPK(incidentID), "ACTIVITY#")
	if err != nil {
		return nil, err
	}
	activities := make([]dynamodbrecord.IncidentActivityRecord, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.IncidentActivityRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityIncidentActivity || !strings.EqualFold(record.TenantID, tenantID) {
			continue
		}
		activities = append(activities, record)
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].Timestamp < activities[j].Timestamp })
	return activities, nil
}

func (r *dynamoMonitorRepository) ExecuteIncidentCommand(ctx context.Context, record commandIdempotencyRecord, now time.Time) (commandIdempotencyRecord, bool, error) {
	incident, found, err := r.GetIncident(ctx, record.TenantID, record.ResourceID)
	if err != nil || !found {
		return commandIdempotencyRecord{}, found, err
	}
	action, changeValue := "", ""
	switch record.Operation {
	case "incident-acknowledge":
		if incident.Status != incidentStatusOpen {
			return commandIdempotencyRecord{}, true, errIncidentNotActionable
		}
		incident.Status = incidentStatusAcknowledged
		incident.AcknowledgedAt = now.UTC().Format(time.RFC3339)
		incident.UpdatedAt = incident.AcknowledgedAt
		action, changeValue = "INCIDENT_ACKNOWLEDGED", incident.AcknowledgedAt
	case "incident-resolve":
		if incident.Status == incidentStatusResolved {
			return commandIdempotencyRecord{}, true, errIncidentNotActionable
		}
		incident.Status = incidentStatusResolved
		incident.ResolvedAt = now.UTC().Format(time.RFC3339)
		incident.UpdatedAt = incident.ResolvedAt
		action, changeValue = "INCIDENT_RESOLVED", incident.ResolvedAt
	default:
		return commandIdempotencyRecord{}, true, fmt.Errorf("unsupported incident command operation %q", record.Operation)
	}
	encoded, err := json.Marshal(toIncidentResponse(incident))
	if err != nil {
		return commandIdempotencyRecord{}, true, err
	}
	record.State = commandIdempotencyCompleted
	record.Response = string(encoded)
	items, err := r.incidentWriteItems(incident, action, now, changeValue)
	if err != nil {
		return commandIdempotencyRecord{}, true, err
	}
	key := sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(record.TenantID), commandIdempotencyAddress(record.TenantID, record.Operation, record.ResourceID, record.Key))
	idempotencyItem := key.AttributeMap()
	for name, value := range map[string]string{"Operation": record.Operation, "ResourceID": record.ResourceID, "Key": record.Key, "Fingerprint": record.Fingerprint, "ReservationToken": record.ReservationToken, "Response": record.Response, "State": string(record.State), "CreatedAt": record.CreatedAt.Format(time.RFC3339)} {
		idempotencyItem[name] = &sharedaws.AttributeValueMemberS{Value: value}
	}
	idempotencyItem["TTL"] = &sharedaws.AttributeValueMemberN{Value: fmt.Sprintf("%d", record.TTL)}
	items = append(items, sharedaws.TransactWriteItem{Put: &sharedaws.Put{
		TableName:                 sharedaws.String(r.tableName),
		Item:                      idempotencyItem,
		ConditionExpression:       sharedaws.String("attribute_not_exists(PK) OR TTL <= :now"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{":now": &sharedaws.AttributeValueMemberN{Value: fmt.Sprintf("%d", record.CreatedAt.Unix())}},
	}})
	if err := r.writeTransaction(ctx, items); err == nil {
		return record, true, nil
	} else if existing, exists, loadErr := r.loadCommandIdempotency(ctx, record); loadErr != nil {
		return commandIdempotencyRecord{}, true, loadErr
	} else if exists {
		return existing, true, nil
	} else {
		return commandIdempotencyRecord{}, true, err
	}
}

// writeIncident is shared with the monitor vertical slice. It persists an
// incident state change along with its audit event, audit change row, and
// incident activity record.
func (r *dynamoMonitorRepository) writeIncident(ctx context.Context, incident dynamodbrecord.IncidentRecord, action string, now time.Time, changeValue string) error {
	items, err := r.incidentWriteItems(incident, action, now, changeValue)
	if err != nil {
		return err
	}
	return r.writeTransaction(ctx, items)
}

func (r *dynamoMonitorRepository) incidentWriteItems(incident dynamodbrecord.IncidentRecord, action string, now time.Time, changeValue string) ([]sharedaws.TransactWriteItem, error) {
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, incident.TenantID, action, incident.ServiceID, incident.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "incident", "", changeValue)
	activity := dynamodbrecord.NewIncidentActivityRecord(incident.TenantID, incident.IncidentID, newActivityID(now), action, now)
	return marshalPutItems(
		r.tableName,
		dynamodbrecord.NewIncidentMonitorItemRecord(incident),
		dynamodbrecord.NewIncidentRefItemRecord(incident),
		dynamodbrecord.NewIncidentMetaItemRecord(incident),
		activity,
		auditEvent,
		change,
	)
}
