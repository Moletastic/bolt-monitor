package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/escalation"
)

type dynamoAPI = sharedaws.DynamoDBAPI

// incidentRecord aliases the shared incident record so handler interfaces and
// tests can stay typed against the shared domain value.
type incidentRecord = dynamodbrecord.IncidentRecord

type dynamoEscalationRepository struct {
	client    dynamoAPI
	tableName string
}

func newDynamoEscalationRepository(client dynamoAPI, tableName string) *dynamoEscalationRepository {
	return &dynamoEscalationRepository{client: client, tableName: tableName}
}

func (r *dynamoEscalationRepository) GetService(ctx context.Context, tenantID, serviceID string) (*serviceRecord, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.ServicePK(tenantID, serviceID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record serviceRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *dynamoEscalationRepository) GetEscalationPolicy(ctx context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.EscalationPolicySK(policyID)},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.EscalationPolicyItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	policy := record.ToEscalationPolicy()
	return &policy, nil
}

func (r *dynamoEscalationRepository) GetChannel(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.NotificationChannelSK(channelID)},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.NotificationChannelItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	channel := record.ToNotificationChannel()
	return &channel, nil
}

func (r *dynamoEscalationRepository) LoadTransitionOutbox(ctx context.Context, tenantID, eventID string) (*dynamodbrecord.TransitionOutboxRecord, error) {
	item, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key:      sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), "TRANSITION_OUTBOX#"+dynamodbschema.NormalizeToken(eventID)).AttributeMap(),
	})
	if err != nil || len(item.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.TransitionOutboxRecord
	if err := sharedaws.UnmarshalMap(item.Item, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *dynamoEscalationRepository) AcknowledgeDispatch(ctx context.Context, tenantID, eventID string) error {
	_, err := r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
		TableName: sharedaws.String(r.tableName),
		Key:      sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), "TRANSITION_OUTBOX#"+dynamodbschema.NormalizeToken(eventID)).AttributeMap(),
		UpdateExpression: sharedaws.String("SET DispatchStatus = :acknowledged"),
		ConditionExpression: sharedaws.String("DispatchStatus = :pending"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pending":     &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.DispatchPending},
			":acknowledged": &sharedaws.AttributeValueMemberS{Value: "acknowledged"},
		},
	})
	if sharedaws.IsConditionalCheckFailure(err) {
		return nil
	}
	return err
}

func (r *dynamoEscalationRepository) PutEscalationState(ctx context.Context, state escalation.EscalationState) error {
	record := dynamodbrecord.NewEscalationStateItemRecord(state)
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: item})
	return err
}

func (r *dynamoEscalationRepository) GetEscalationState(ctx context.Context, tenantID, incidentID string) (*escalation.EscalationState, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "ESCALATION_STATE"},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.EscalationStateItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	state := record.ToEscalationState()
	if !strings.EqualFold(state.TenantID, tenantID) {
		return nil, nil
	}
	return &state, nil
}

func (r *dynamoEscalationRepository) GetIncident(ctx context.Context, incidentID string) (*incidentRecord, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.IncidentItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	incident := record.ToIncident()
	return &incident, nil
}

func (r *dynamoEscalationRepository) CreateIncident(ctx context.Context, incident incidentRecord) error {
	items := []any{
		dynamodbrecord.NewIncidentMonitorItemRecord(incident),
		dynamodbrecord.NewIncidentRefItemRecord(incident),
		dynamodbrecord.NewIncidentMetaItemRecord(incident),
	}
	transactItems := make([]sharedaws.TransactWriteItem, 0, len(items))
	for _, item := range items {
		av, err := sharedaws.MarshalMap(item)
		if err != nil {
			return err
		}
		transactItems = append(transactItems, sharedaws.TransactWriteItem{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: av}})
	}
	_, err := r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: transactItems})
	return err
}

type serviceRecord struct {
	TenantID           string                          `dynamodbav:"TenantID"`
	ServiceID          string                          `dynamodbav:"ServiceID"`
	EscalationPolicyID string                          `dynamodbav:"EscalationPolicyID,omitempty"`
	BusinessHours      *escalation.BusinessHoursConfig `dynamodbav:"BusinessHours,omitempty"`
}

const (
	incidentStatusOpen         = "open"
	incidentStatusAcknowledged = "acknowledged"
	incidentStatusResolved     = "resolved"
)

func newEscalationExhaustedIncident(original incidentRecord, now time.Time) incidentRecord {
	timestamp := now.UTC().Format(time.RFC3339)
	return incidentRecord{
		TenantID:           original.TenantID,
		ServiceID:          original.ServiceID,
		MonitorID:          original.MonitorID,
		IncidentID:         fmt.Sprintf("INC_ESC_%d", now.UTC().UnixNano()),
		Type:               "escalation.exhausted",
		Summary:            "Escalation exhausted for incident " + original.IncidentID,
		Status:             incidentStatusOpen,
		OpenedAt:           timestamp,
		UpdatedAt:          timestamp,
		Origin:             "system",
		OriginalIncidentID: original.IncidentID,
	}
}
