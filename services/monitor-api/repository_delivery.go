package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/notifications"
)

func (r *dynamoMonitorRepository) ListIncidentDeliveries(ctx context.Context, tenantID, incidentID string) ([]notifications.DeliveryRecord, error) {
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.IncidentPK(incidentID)},
			":prefix": &sharedaws.AttributeValueMemberS{Value: "DELIVERY#"},
		},
		ScanIndexForward: sharedaws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	out2 := make([]notifications.DeliveryRecord, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.DeliveryItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if !record.BelongsToTenant(tenantID, incidentID) {
			continue
		}
		out2 = append(out2, record.ToDelivery())
	}
	return out2, nil
}

func (r *dynamoMonitorRepository) PrepareDeliveryReplay(ctx context.Context, command notifications.ReplayCommand, fingerprint string, now time.Time, retention time.Duration) (string, error) {
	if err := dynamodbrecord.ValidateReplayIdempotencyKey(command.IdempotencyKey); err != nil {
		return "", err
	}
	deliveries, err := r.ListIncidentDeliveries(ctx, command.TenantID, command.IncidentID)
	if err != nil {
		return "", err
	}
	var delivery *notifications.DeliveryRecord
	for i := range deliveries {
		if deliveries[i].DeliveryID == command.DeliveryID {
			delivery = &deliveries[i]
			break
		}
	}
	if delivery == nil {
		return "", errors.New("delivery not found")
	}
	if delivery.State != notifications.DeliveryTerminalFailed {
		return "", errors.New("delivery is not eligible for replay")
	}
	record := dynamodbrecord.NewDeliveryItemRecord(*delivery)
	idempotency := dynamodbrecord.NewReplayIdempotencyItemRecord(notifications.ReplayIdempotencyRecord{
		TenantID: command.TenantID, IncidentID: command.IncidentID, DeliveryID: command.DeliveryID,
		Operation: "delivery_replay", IdempotencyKey: command.IdempotencyKey, RequestFingerprint: fingerprint,
		ResultDeliveryID: command.DeliveryID, CreatedAt: now.UTC().Format(time.RFC3339), ExpiresAt: now.Add(retention).Unix(),
	})
	bucket := now.UTC().Format(dynamodbrecord.DispatchPendingBucketFormat)
	pending := dynamodbrecord.NewDispatchPendingRecord(command.TenantID, command.DeliveryID, bucket, "00")
	deliveryAV, err := sharedaws.MarshalMap(record)
	if err != nil {
		return "", err
	}
	_ = deliveryAV
	idempotencyAV, err := sharedaws.MarshalMap(idempotency)
	if err != nil {
		return "", err
	}
	pendingAV, err := sharedaws.MarshalMap(pending)
	if err != nil {
		return "", err
	}
	replayID := "REPLAY_" + strings.ToUpper(command.DeliveryID)
	outbox := dynamodbrecord.TransitionOutboxRecord{
		PK: dynamodbschema.TenantPK(command.TenantID), SK: "TRANSITION_OUTBOX#" + dynamodbschema.NormalizeToken(replayID),
		EntityType: dynamodbschema.EntityTransitionOutbox, TenantID: command.TenantID,
		EventID: replayID, TransitionID: command.TransitionID, IncidentID: command.IncidentID,
		DispatchStatus: dynamodbrecord.DispatchPending, Version: notifications.CanonicalEnvelopeVersion,
		Kind: notifications.CanonicalKindReplay, SourceKind: notifications.CanonicalSourceReplay, DeliveryID: command.DeliveryID,
		CreatedAt: now.UTC().Format(time.RFC3339),
	}
	outboxAV, err := sharedaws.MarshalMap(outbox)
	if err != nil {
		return "", err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: []sharedaws.TransactWriteItem{
		{Update: &sharedaws.Update{TableName: sharedaws.String(r.tableName), Key: sharedaws.NewPrimaryKey(record.PK, record.SK).AttributeMap(),
			UpdateExpression:         sharedaws.String("SET #state = :pending, UpdatedAt = :updated"),
			ConditionExpression:      sharedaws.String("TenantID = :tenant AND IncidentID = :incident AND DeliveryID = :delivery AND #state = :terminal"),
			ExpressionAttributeNames: map[string]string{"#state": "State"},
			ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
				":tenant": &sharedaws.AttributeValueMemberS{Value: command.TenantID}, ":incident": &sharedaws.AttributeValueMemberS{Value: command.IncidentID},
				":delivery": &sharedaws.AttributeValueMemberS{Value: command.DeliveryID}, ":terminal": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryTerminalFailed)},
				":pending": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryPending)},
				":updated": &sharedaws.AttributeValueMemberS{Value: now.UTC().Format(time.RFC3339)},
			},
		}},
		{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: idempotencyAV, ConditionExpression: sharedaws.String("attribute_not_exists(PK)")}},
		{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: pendingAV, ConditionExpression: sharedaws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)")}},
		{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: outboxAV, ConditionExpression: sharedaws.String("attribute_not_exists(PK)")}},
	}})
	if err != nil {
		return "", fmt.Errorf("prepare replay: %w", err)
	}
	return command.DeliveryID, nil
}

func (r *dynamoMonitorRepository) LookupReplayIdempotency(ctx context.Context, tenantID, incidentID, deliveryID, key string) (*notifications.ReplayIdempotencyRecord, error) {
	address := dynamodbrecord.ReplayIdempotencyAddress(tenantID, incidentID, deliveryID, key)
	item := dynamodbschema.ReplayIdempotencyItem(tenantID, address)
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: sharedaws.NewPrimaryKey(item.PK, item.SK).AttributeMap()})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.ReplayIdempotencyItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	r2 := record.ToRecord()
	return &r2, nil
}
