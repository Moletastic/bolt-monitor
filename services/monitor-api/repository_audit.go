package main

import (
	"context"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
)

// AuditStore is the narrow interface used by audit-event handlers. These
// methods issue GSI queries (named methods retained per §4.6); they expose
// pagination so dashboards can paginate without truncation.
type AuditStore interface {
	ListMonitorAuditEvents(context.Context, string, string, string) ([]auditEventView, error)
	ListMonitorAuditEventsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error)
	ListServiceAuditEvents(context.Context, string, string) ([]auditEventView, error)
	ListServiceAuditEventsPage(context.Context, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error)
}

func (r *dynamoMonitorRepository) ListMonitorAuditEvents(ctx context.Context, tenantID, serviceID, monitorID string) ([]auditEventView, error) {
	page, err := r.ListMonitorAuditEventsPage(ctx, tenantID, serviceID, monitorID, 20, nil)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *dynamoMonitorRepository) ListMonitorAuditEventsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	if err := r.requireTableName(); err != nil {
		return historyPage[auditEventView]{}, err
	}
	resource := dynamodbschema.AuditResourceItem(tenantID, serviceID, monitorID, "cursor", "cursor")
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		IndexName:              sharedaws.String(dynamodbschema.GSIAuditByResource),
		KeyConditionExpression: sharedaws.String("GSI3PK = :pk AND begins_with(GSI3SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: resource.GSI3PK},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "AUDIT#"},
		},
		ScanIndexForward:  sharedaws.Bool(false),
		Limit:             sharedaws.Int32(limit),
		ExclusiveStartKey: startKey,
	})
	if err != nil {
		return historyPage[auditEventView]{}, err
	}
	eventsList := make([]auditEventView, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.AuditEventRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return historyPage[auditEventView]{}, err
		}
		eventsList = append(eventsList, auditEventView{
			AuditID:    record.AuditID,
			ServiceID:  record.ServiceID,
			MonitorID:  record.MonitorID,
			EventType:  record.Action,
			OccurredAt: record.Timestamp,
			Actor:      record.Actor,
			Origin:     record.Origin,
		})
	}
	return historyPage[auditEventView]{Items: eventsList, NextKey: out.LastEvaluatedKey}, nil
}

func (r *dynamoMonitorRepository) ListServiceAuditEvents(ctx context.Context, tenantID, serviceID string) ([]auditEventView, error) {
	page, err := r.ListServiceAuditEventsPage(ctx, tenantID, serviceID, 20, nil)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *dynamoMonitorRepository) ListServiceAuditEventsPage(ctx context.Context, tenantID, serviceID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	if err := r.requireTableName(); err != nil {
		return historyPage[auditEventView]{}, err
	}
	resource := dynamodbschema.AuditResourceItem(tenantID, serviceID, "", "cursor", "cursor")
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		IndexName:              sharedaws.String(dynamodbschema.GSIAuditByResource),
		KeyConditionExpression: sharedaws.String("GSI3PK = :pk AND begins_with(GSI3SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: resource.GSI3PK},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "AUDIT#"},
		},
		ScanIndexForward:  sharedaws.Bool(false),
		Limit:             sharedaws.Int32(limit),
		ExclusiveStartKey: startKey,
	})
	if err != nil {
		return historyPage[auditEventView]{}, err
	}
	eventsList := make([]auditEventView, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.AuditEventRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return historyPage[auditEventView]{}, err
		}
		eventsList = append(eventsList, auditEventView{
			AuditID:    record.AuditID,
			ServiceID:  record.ServiceID,
			MonitorID:  record.MonitorID,
			EventType:  record.Action,
			OccurredAt: record.Timestamp,
			Actor:      record.Actor,
			Origin:     record.Origin,
		})
	}
	return historyPage[auditEventView]{Items: eventsList, NextKey: out.LastEvaluatedKey}, nil
}
