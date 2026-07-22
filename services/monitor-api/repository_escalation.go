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
	"bolt-monitor/shared/escalation"
)

// EscalationStore is the narrow interface used by escalation-policy,
// notification-channel, and escalation-state handlers.
type EscalationStore interface {
	CreateEscalationPolicy(context.Context, escalation.EscalationPolicy) (escalation.EscalationPolicy, error)
	ListEscalationPolicies(context.Context, string) ([]escalation.EscalationPolicy, error)
	GetEscalationPolicy(context.Context, string, string) (*escalation.EscalationPolicy, error)
	UpdateEscalationPolicy(context.Context, escalation.EscalationPolicy) (escalation.EscalationPolicy, error)
	DeleteEscalationPolicy(context.Context, string, string) error
	CreateNotificationChannel(context.Context, escalation.NotificationChannel) (escalation.NotificationChannel, error)
	ListNotificationChannels(context.Context, string) ([]escalation.NotificationChannel, error)
	GetNotificationChannel(context.Context, string, string) (*escalation.NotificationChannel, error)
	UpdateNotificationChannel(context.Context, escalation.NotificationChannel) (escalation.NotificationChannel, error)
	DeleteNotificationChannel(context.Context, string, string) error
	ChannelsReferencedByRoutes(context.Context, string, string) ([]routeReference, error)
	RecordNotificationChannelTestAudit(context.Context, string, string, string, string, string, time.Time) error
	GetEscalationState(context.Context, string, string) (*escalation.EscalationState, error)
}

func (r *dynamoMonitorRepository) CreateNotificationChannel(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.NotificationChannel{}, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	channel.CreatedAt = now
	channel.UpdatedAt = now
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewNotificationChannelItemRecord(channel))
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	if err := r.replaceSearchIndex(ctx, channel.TenantID, searchResourceChannel, channel.ChannelID, "", buildChannelSearchRecords(channel)); err != nil {
		return escalation.NotificationChannel{}, err
	}
	return channel, nil
}

func (r *dynamoMonitorRepository) ListNotificationChannels(ctx context.Context, tenantID string) ([]escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "NOTIFICATION_CHANNEL#")
	if err != nil {
		return nil, err
	}
	channels := make([]escalation.NotificationChannel, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.NotificationChannelItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityNotificationChannel {
			continue
		}
		channels = append(channels, record.ToNotificationChannel())
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].Name < channels[j].Name })
	return channels, nil
}

func (r *dynamoMonitorRepository) GetNotificationChannel(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.NotificationChannelSK(channelID)}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record dynamodbrecord.NotificationChannelItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	channel := record.ToNotificationChannel()
	return &channel, nil
}

func (r *dynamoMonitorRepository) UpdateNotificationChannel(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.NotificationChannel{}, err
	}
	existing, err := r.GetNotificationChannel(ctx, channel.TenantID, channel.ChannelID)
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	if existing != nil && strings.TrimSpace(existing.CreatedAt) != "" {
		channel.CreatedAt = existing.CreatedAt
	}
	channel.UpdatedAt = r.now().UTC().Format(time.RFC3339)
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewNotificationChannelItemRecord(channel))
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.NotificationChannel{}, err
	}
	if err := r.replaceSearchIndex(ctx, channel.TenantID, searchResourceChannel, channel.ChannelID, "", buildChannelSearchRecords(channel)); err != nil {
		return escalation.NotificationChannel{}, err
	}
	return channel, nil
}

func (r *dynamoMonitorRepository) DeleteNotificationChannel(ctx context.Context, tenantID, channelID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.NotificationChannelSK(channelID)}}})
	if err != nil {
		return err
	}
	return r.deleteSearchIndex(ctx, tenantID, searchResourceChannel, channelID, "")
}

func (r *dynamoMonitorRepository) RecordNotificationChannelTestAudit(ctx context.Context, tenantID, channelID, channelType, outcome, reason string, now time.Time) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	auditID := newAuditID(now)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, tenantID, "NOTIFICATION_CHANNEL_TEST_SENT", "notification-channel", channelID)
	records := []any{
		auditEvent,
		dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "channelType", "", channelType),
		dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "outcome", "", outcome),
	}
	if strings.TrimSpace(reason) != "" {
		records = append(records, dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "reason", "", reason))
	}
	items, err := marshalPutItems(r.tableName, records...)
	if err != nil {
		return err
	}
	return r.writeTransaction(ctx, items)
}

func (r *dynamoMonitorRepository) ChannelsReferencedByRoutes(ctx context.Context, tenantID, channelID string) ([]routeReference, error) {
	policies, err := r.ListEscalationPolicies(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	references := []routeReference{}
	for _, policy := range policies {
		if policyReferencesChannel(policy, channelID) {
			references = append(references, routeReference{PolicyID: policy.PolicyID, Name: policy.Name})
		}
	}
	return references, nil
}

func policyReferencesChannel(policy escalation.EscalationPolicy, channelID string) bool {
	needle := strings.TrimSpace(channelID)
	for _, path := range []escalation.EscalationPath{policy.BusinessHoursPath, policy.OffHoursPath} {
		for _, step := range path.Steps {
			if strings.EqualFold(step.ChannelID, needle) {
				return true
			}
		}
	}
	return false
}

func (r *dynamoMonitorRepository) CreateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	now := r.now().UTC().Format(time.RFC3339)
	policy.CreatedAt = now
	policy.UpdatedAt = now
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPolicyItemRecord(policy))
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	if err := r.replaceSearchIndex(ctx, policy.TenantID, searchResourcePolicy, policy.PolicyID, "", buildPolicySearchRecords(policy)); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return policy, nil
}

func (r *dynamoMonitorRepository) ListEscalationPolicies(ctx context.Context, tenantID string) ([]escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "ESCALATION_POLICY#")
	if err != nil {
		return nil, err
	}
	policies := make([]escalation.EscalationPolicy, 0, len(out))
	for _, item := range out {
		var record dynamodbrecord.EscalationPolicyItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntityEscalationPolicy {
			continue
		}
		policy := record.ToEscalationPolicy()
		policies = append(policies, policy)
	}
	sort.Slice(policies, func(i, j int) bool { return policies[i].PolicyID < policies[j].PolicyID })
	return policies, nil
}

func (r *dynamoMonitorRepository) GetEscalationPolicy(ctx context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.EscalationPolicySK(policyID)}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record dynamodbrecord.EscalationPolicyItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	policy := record.ToEscalationPolicy()
	return &policy, nil
}

// MigrateRouteInlineChannels is an explicit operator migration command. Query
// methods must return legacy inline-channel records without modifying them.
func (r *dynamoMonitorRepository) MigrateRouteInlineChannels(ctx context.Context, policy *escalation.EscalationPolicy) error {
	if policy == nil {
		return nil
	}
	migrated := false
	stepIndex := 0
	paths := []*escalation.EscalationPath{&policy.BusinessHoursPath, &policy.OffHoursPath}
	for _, path := range paths {
		for i := range path.Steps {
			step := &path.Steps[i]
			if strings.TrimSpace(step.ChannelID) != "" || len(step.Channels) == 0 {
				stepIndex++
				continue
			}
			legacy := step.Channels[0]
			channelID := fmt.Sprintf("%s#%s#%d", strings.ToUpper(strings.TrimSpace(policy.TenantID)), strings.ToUpper(strings.TrimSpace(policy.PolicyID)), stepIndex)
			channel := escalation.NotificationChannel{
				TenantID:  policy.TenantID,
				ChannelID: channelID,
				Name:      fmt.Sprintf("Migrated channel %d", stepIndex+1),
				Type:      legacy.Type,
				Target:    legacy.Target,
				Config:    append(json.RawMessage(nil), legacy.Config...),
				CreatedAt: policy.CreatedAt,
				UpdatedAt: policy.UpdatedAt,
			}
			if _, err := r.CreateNotificationChannel(ctx, channel); err != nil {
				return err
			}
			step.ChannelID = channelID
			step.Channels = nil
			migrated = true
			stepIndex++
		}
	}
	if !migrated {
		return nil
	}
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPolicyItemRecord(*policy))
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	return err
}

func (r *dynamoMonitorRepository) UpdateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	if err := r.requireTableName(); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	existing, err := r.GetEscalationPolicy(ctx, policy.TenantID, policy.PolicyID)
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	if existing != nil && strings.TrimSpace(existing.CreatedAt) != "" {
		policy.CreatedAt = existing.CreatedAt
	}
	policy.UpdatedAt = r.now().UTC().Format(time.RFC3339)
	av, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPolicyItemRecord(policy))
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: av})
	if err != nil {
		return escalation.EscalationPolicy{}, err
	}
	if err := r.replaceSearchIndex(ctx, policy.TenantID, searchResourcePolicy, policy.PolicyID, "", buildPolicySearchRecords(policy)); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return policy, nil
}

func (r *dynamoMonitorRepository) DeleteEscalationPolicy(ctx context.Context, tenantID, policyID string) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	_, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)}, "SK": &sharedaws.AttributeValueMemberS{Value: dynamodbrecord.EscalationPolicySK(policyID)}}})
	if err != nil {
		return err
	}
	return r.deleteSearchIndex(ctx, tenantID, searchResourcePolicy, policyID, "")
}

func (r *dynamoMonitorRepository) GetEscalationState(ctx context.Context, tenantID, incidentID string) (*escalation.EscalationState, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)}, "SK": &sharedaws.AttributeValueMemberS{Value: "ESCALATION_STATE"}}})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var record dynamodbrecord.EscalationStateItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	state := record.ToEscalationState()
	return &state, nil
}
