package main

import (
	"context"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
)

// IncidentStore is the narrow interface used by incident-lifecycle handlers.
// It deliberately excludes the audit and scheduler methods so handler tests
// stay focused on incident flows.
type IncidentStore interface {
	ListIncidents(context.Context, string, string) ([]dynamodbrecord.IncidentRecord, error)
	GetIncident(context.Context, string, string) (dynamodbrecord.IncidentRecord, bool, error)
	ListIncidentActivities(context.Context, string, string) ([]dynamodbrecord.IncidentActivityRecord, error)
	AcknowledgeIncident(context.Context, string, string, time.Time) (dynamodbrecord.IncidentRecord, bool, error)
	ResolveIncident(context.Context, string, string, time.Time) (dynamodbrecord.IncidentRecord, bool, error)
}

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

func (r *dynamoMonitorRepository) AcknowledgeIncident(ctx context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, found, err := r.GetIncident(ctx, tenantID, incidentID)
	if err != nil || !found {
		return incident, found, err
	}
	if incident.Status != incidentStatusOpen {
		return dynamodbrecord.IncidentRecord{}, true, errIncidentNotActionable
	}
	incident.Status = incidentStatusAcknowledged
	incident.AcknowledgedAt = now.UTC().Format(time.RFC3339)
	incident.UpdatedAt = incident.AcknowledgedAt
	if err := r.writeIncident(ctx, incident, "INCIDENT_ACKNOWLEDGED", now, incident.AcknowledgedAt); err != nil {
		return dynamodbrecord.IncidentRecord{}, true, err
	}
	return incident, true, nil
}

func (r *dynamoMonitorRepository) ResolveIncident(ctx context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, found, err := r.GetIncident(ctx, tenantID, incidentID)
	if err != nil || !found {
		return incident, found, err
	}
	if incident.Status == incidentStatusResolved {
		return dynamodbrecord.IncidentRecord{}, true, errIncidentNotActionable
	}
	incident.Status = incidentStatusResolved
	incident.ResolvedAt = now.UTC().Format(time.RFC3339)
	incident.UpdatedAt = incident.ResolvedAt
	if err := r.writeIncident(ctx, incident, "INCIDENT_RESOLVED", now, incident.ResolvedAt); err != nil {
		return dynamodbrecord.IncidentRecord{}, true, err
	}
	return incident, true, nil
}

// writeIncident is shared with the monitor vertical slice. It persists an
// incident state change along with its audit event, audit change row, and
// incident activity record.
func (r *dynamoMonitorRepository) writeIncident(ctx context.Context, incident dynamodbrecord.IncidentRecord, action string, now time.Time, changeValue string) error {
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, incident.TenantID, action, incident.ServiceID, incident.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "incident", "", changeValue)
	activity := dynamodbrecord.NewIncidentActivityRecord(incident.TenantID, incident.IncidentID, newActivityID(now), action, now)
	items, err := marshalPutItems(
		r.tableName,
		dynamodbrecord.NewIncidentMonitorItemRecord(incident),
		dynamodbrecord.NewIncidentRefItemRecord(incident),
		dynamodbrecord.NewIncidentMetaItemRecord(incident),
		activity,
		auditEvent,
		change,
	)
	if err != nil {
		return err
	}
	return r.writeTransaction(ctx, items)
}
