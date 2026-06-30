package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	shareddynamo "bolt-monitor/shared/dynamodb"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/escalation"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoAPI = shareddynamo.DynamoDBAPI

type dynamoEscalationRepository struct {
	client    dynamoAPI
	tableName string
}

func newDynamoEscalationRepository(client dynamoAPI, tableName string) *dynamoEscalationRepository {
	return &dynamoEscalationRepository{client: client, tableName: tableName}
}

func (r *dynamoEscalationRepository) GetService(ctx context.Context, tenantID, serviceID string) (*serviceRecord, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: dynamodbschema.ServicePK(tenantID, serviceID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record serviceRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *dynamoEscalationRepository) GetEscalationPolicy(ctx context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "ESCALATION_POLICY#" + strings.ToUpper(strings.TrimSpace(policyID))},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record escalationPolicyRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	policy := record.toPolicy()
	return &policy, nil
}

func (r *dynamoEscalationRepository) GetChannel(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "NOTIFICATION_CHANNEL#" + strings.ToUpper(strings.TrimSpace(channelID))},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record notificationChannelRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	channel := record.toChannel()
	return &channel, nil
}

func (r *dynamoEscalationRepository) PutEscalationState(ctx context.Context, state escalation.EscalationState) error {
	item, err := attributevalue.MarshalMap(newEscalationStateRecord(state))
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.tableName), Item: item})
	return err
}

func (r *dynamoEscalationRepository) GetEscalationState(ctx context.Context, tenantID, incidentID string) (*escalation.EscalationState, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "ESCALATION_STATE"},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record escalationStateRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	state := record.toState()
	if !strings.EqualFold(state.TenantID, tenantID) {
		return nil, nil
	}
	return &state, nil
}

func (r *dynamoEscalationRepository) GetIncident(ctx context.Context, incidentID string) (*incidentRecord, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record incidentRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *dynamoEscalationRepository) CreateIncident(ctx context.Context, incident incidentRecord) error {
	items := []any{newIncidentMonitorItemRecord(incident), newIncidentRefItemRecord(incident), newIncidentMetaItemRecord(incident)}
	transactItems := make([]ddbtypes.TransactWriteItem, 0, len(items))
	for _, item := range items {
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return err
		}
		transactItems = append(transactItems, ddbtypes.TransactWriteItem{Put: &ddbtypes.Put{TableName: aws.String(r.tableName), Item: av}})
	}
	_, err := r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: transactItems})
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

type incidentRecord struct {
	PK                 string `dynamodbav:"PK,omitempty"`
	SK                 string `dynamodbav:"SK,omitempty"`
	EntityType         string `dynamodbav:"EntityType,omitempty"`
	TenantID           string `dynamodbav:"TenantID"`
	ServiceID          string `dynamodbav:"ServiceID"`
	MonitorID          string `dynamodbav:"MonitorID"`
	IncidentID         string `dynamodbav:"IncidentID"`
	Type               string `dynamodbav:"Type,omitempty"`
	Summary            string `dynamodbav:"Summary"`
	Status             string `dynamodbav:"Status"`
	OpenedAt           string `dynamodbav:"OpenedAt"`
	AcknowledgedAt     string `dynamodbav:"AcknowledgedAt,omitempty"`
	ResolvedAt         string `dynamodbav:"ResolvedAt,omitempty"`
	UpdatedAt          string `dynamodbav:"UpdatedAt"`
	Origin             string `dynamodbav:"Origin,omitempty"`
	OriginalIncidentID string `dynamodbav:"OriginalIncidentID,omitempty"`
	GSI1PK             string `dynamodbav:"GSI1PK,omitempty"`
	GSI1SK             string `dynamodbav:"GSI1SK,omitempty"`
}

type escalationPolicyRecord struct {
	TenantID          string                    `dynamodbav:"TenantID"`
	PolicyID          string                    `dynamodbav:"PolicyID"`
	Name              string                    `dynamodbav:"Name"`
	Description       string                    `dynamodbav:"Description,omitempty"`
	BusinessHoursPath escalation.EscalationPath `dynamodbav:"BusinessHoursPath"`
	OffHoursPath      escalation.EscalationPath `dynamodbav:"OffHoursPath"`
	CreatedAt         string                    `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt         string                    `dynamodbav:"UpdatedAt,omitempty"`
}

type notificationChannelRecord struct {
	TenantID  string                 `dynamodbav:"TenantID"`
	ChannelID string                 `dynamodbav:"ChannelID"`
	Name      string                 `dynamodbav:"Name"`
	Type      escalation.ChannelType `dynamodbav:"Type"`
	Target    string                 `dynamodbav:"Target"`
	Config    []byte                 `dynamodbav:"Config,omitempty"`
	CreatedAt string                 `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt string                 `dynamodbav:"UpdatedAt,omitempty"`
}

func (r notificationChannelRecord) toChannel() escalation.NotificationChannel {
	return escalation.NotificationChannel{TenantID: r.TenantID, ChannelID: r.ChannelID, Name: r.Name, Type: r.Type, Target: r.Target, Config: append([]byte(nil), r.Config...), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (r escalationPolicyRecord) toPolicy() escalation.EscalationPolicy {
	return escalation.EscalationPolicy{TenantID: r.TenantID, PolicyID: r.PolicyID, Name: r.Name, Description: r.Description, BusinessHoursPath: r.BusinessHoursPath, OffHoursPath: r.OffHoursPath, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

type escalationStateRecord struct {
	PK           string                      `dynamodbav:"PK"`
	SK           string                      `dynamodbav:"SK"`
	EntityType   string                      `dynamodbav:"EntityType"`
	TenantID     string                      `dynamodbav:"TenantID"`
	IncidentID   string                      `dynamodbav:"IncidentID"`
	PolicyID     string                      `dynamodbav:"PolicyID"`
	ServiceID    string                      `dynamodbav:"ServiceID"`
	MonitorID    string                      `dynamodbav:"MonitorID"`
	CurrentStep  int                         `dynamodbav:"CurrentStep"`
	StepsFired   []int                       `dynamodbav:"StepsFired,omitempty"`
	SelectedPath string                      `dynamodbav:"SelectedPath,omitempty"`
	ScheduledFor string                      `dynamodbav:"ScheduledFor,omitempty"`
	Status       escalation.EscalationStatus `dynamodbav:"Status"`
	CreatedAt    string                      `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt    string                      `dynamodbav:"UpdatedAt,omitempty"`
}

func newEscalationStateRecord(state escalation.EscalationState) escalationStateRecord {
	item := dynamodbschema.EscalationStateItem(state.TenantID, state.IncidentID)
	return escalationStateRecord{PK: item.PK, SK: item.SK, EntityType: item.EntityType, TenantID: strings.ToUpper(state.TenantID), IncidentID: strings.ToUpper(state.IncidentID), PolicyID: strings.ToUpper(strings.TrimSpace(state.PolicyID)), ServiceID: strings.ToLower(state.ServiceID), MonitorID: strings.ToLower(state.MonitorID), CurrentStep: state.CurrentStep, StepsFired: append([]int(nil), state.StepsFired...), SelectedPath: state.SelectedPath, ScheduledFor: state.ScheduledFor, Status: state.Status, CreatedAt: state.CreatedAt, UpdatedAt: state.UpdatedAt}
}

func (r escalationStateRecord) toState() escalation.EscalationState {
	return escalation.EscalationState{TenantID: r.TenantID, IncidentID: r.IncidentID, PolicyID: r.PolicyID, ServiceID: r.ServiceID, MonitorID: r.MonitorID, CurrentStep: r.CurrentStep, StepsFired: append([]int(nil), r.StepsFired...), SelectedPath: r.SelectedPath, ScheduledFor: r.ScheduledFor, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func newIncidentMonitorItemRecord(incident incidentRecord) incidentRecord {
	item := dynamodbschema.IncidentItem(incident.TenantID, incident.ServiceID, incident.MonitorID, incident.IncidentID, incident.OpenedAt, strings.ToUpper(incident.Status))
	incident.PK = item.PK
	incident.SK = item.SK
	incident.EntityType = dynamodbschema.EntityIncident
	incident.GSI1PK = item.GSI1PK
	incident.GSI1SK = item.GSI1SK
	incident.TenantID = strings.ToUpper(incident.TenantID)
	incident.ServiceID = strings.ToLower(incident.ServiceID)
	incident.MonitorID = strings.ToLower(incident.MonitorID)
	incident.IncidentID = strings.ToUpper(incident.IncidentID)
	incident.OriginalIncidentID = strings.ToUpper(strings.TrimSpace(incident.OriginalIncidentID))
	return incident
}

func newIncidentRefItemRecord(incident incidentRecord) incidentRecord {
	incident.PK = dynamodbschema.TenantPK(incident.TenantID)
	incident.SK = "INCIDENT#" + incident.OpenedAt + "#" + strings.ToUpper(incident.IncidentID)
	incident.EntityType = "IncidentRef"
	incident.TenantID = strings.ToUpper(incident.TenantID)
	incident.ServiceID = strings.ToLower(incident.ServiceID)
	incident.MonitorID = strings.ToLower(incident.MonitorID)
	incident.IncidentID = strings.ToUpper(incident.IncidentID)
	incident.OriginalIncidentID = strings.ToUpper(strings.TrimSpace(incident.OriginalIncidentID))
	return incident
}

func newIncidentMetaItemRecord(incident incidentRecord) incidentRecord {
	incident.PK = dynamodbschema.IncidentPK(incident.IncidentID)
	incident.SK = "META"
	incident.EntityType = dynamodbschema.EntityIncident
	incident.TenantID = strings.ToUpper(incident.TenantID)
	incident.ServiceID = strings.ToLower(incident.ServiceID)
	incident.MonitorID = strings.ToLower(incident.MonitorID)
	incident.IncidentID = strings.ToUpper(incident.IncidentID)
	incident.OriginalIncidentID = strings.ToUpper(strings.TrimSpace(incident.OriginalIncidentID))
	return incident
}

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
