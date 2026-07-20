package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/notifications"
)

// ListPendingDispatch reads one sparse recovery partition. It never scans the
// primary table and returns the DynamoDB cursor for bounded reconciliation.
func (r *dynamoEscalationRepository) ListPendingDispatch(ctx context.Context, tenantID, bucketShard string, limit int, cursor map[string]sharedaws.AttributeValue) ([]dynamodbrecord.DispatchPendingRecord, map[string]sharedaws.AttributeValue, error) {
	queryLimit := boundedQueryLimit(limit)
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk": &sharedaws.AttributeValueMemberS{Value: fmt.Sprintf("DISPATCH_PENDING#%s#%s", dynamodbschema.NormalizeToken(tenantID), dynamodbschema.NormalizeToken(bucketShard))},
		},
		Limit:             sharedaws.Int32(queryLimit),
		ExclusiveStartKey: cursor,
	})
	if err != nil {
		return nil, nil, err
	}
	records := make([]dynamodbrecord.DispatchPendingRecord, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.DispatchPendingRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, nil, err
		}
		if strings.EqualFold(record.TenantID, tenantID) {
			records = append(records, record)
		}
	}
	return records, out.LastEvaluatedKey, nil
}

func boundedQueryLimit(limit int) int32 {
	if limit <= 0 {
		return 25
	}
	if limit >= 2147483647 {
		return 2147483647
	}
	return int32(limit)
}

func (r *dynamoEscalationRepository) GetEscalationPlan(ctx context.Context, tenantID, incidentID, transitionID string) (*notifications.EscalationPlan, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key:       sharedaws.NewPrimaryKey(dynamodbschema.IncidentPK(incidentID), "ESCALATION_PLAN#"+dynamodbschema.NormalizeToken(transitionID)).AttributeMap(),
	})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.EscalationPlanItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	if !record.BelongsToTenant(tenantID, incidentID) {
		return nil, nil
	}
	plan := record.ToEscalationPlan()
	return &plan, nil
}

func (r *dynamoEscalationRepository) CreateEscalationPlan(ctx context.Context, plan notifications.EscalationPlan) error {
	if err := plan.Validate(); err != nil {
		return err
	}
	item, err := sharedaws.MarshalMap(dynamodbrecord.NewEscalationPlanItemRecord(plan))
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{
		TableName:           sharedaws.String(r.tableName),
		Item:                item,
		ConditionExpression: sharedaws.String("attribute_not_exists(PK)"),
	})
	if sharedaws.IsConditionalCheckFailure(err) {
		return nil
	}
	return err
}

func (r *dynamoEscalationRepository) CreateDelivery(ctx context.Context, delivery notifications.DeliveryRecord) error {
	if err := delivery.Validate(); err != nil {
		return err
	}
	item, err := sharedaws.MarshalMap(dynamodbrecord.NewDeliveryItemRecord(delivery))
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{
		TableName:           sharedaws.String(r.tableName),
		Item:                item,
		ConditionExpression: sharedaws.String("attribute_not_exists(PK)"),
	})
	if sharedaws.IsConditionalCheckFailure(err) {
		return nil
	}
	return err
}

func (r *dynamoEscalationRepository) ListIncidentDeliveries(ctx context.Context, tenantID, incidentID string) ([]notifications.DeliveryRecord, error) {
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
	result := make([]notifications.DeliveryRecord, 0, len(out.Items))
	for _, item := range out.Items {
		var record dynamodbrecord.DeliveryItemRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.BelongsToTenant(tenantID, incidentID) {
			result = append(result, record.ToDelivery())
		}
	}
	return result, nil
}

func (r *dynamoEscalationRepository) ClaimDelivery(ctx context.Context, tenantID, incidentID, deliveryID string, now time.Time, lease time.Duration) (*notifications.DeliveryRecord, string, error) {
	deliveries, err := r.ListIncidentDeliveries(ctx, tenantID, incidentID)
	if err != nil {
		return nil, "", err
	}
	var current *notifications.DeliveryRecord
	for i := range deliveries {
		if strings.EqualFold(deliveries[i].DeliveryID, deliveryID) {
			current = &deliveries[i]
			break
		}
	}
	if current == nil {
		return nil, "", nil
	}
	if current.State.IsTerminal() || (current.State == notifications.DeliveryInFlight && current.LeaseUntil != "" && current.LeaseUntil > now.UTC().Format(time.RFC3339)) {
		return current, "", nil
	}
	token, err := newFencingToken()
	if err != nil {
		return nil, "", err
	}
	leaseUntil := now.UTC().Add(lease).Format(time.RFC3339)
	key := dynamodbschema.IncidentPK(incidentID)
	// Delivery SK is stable once created; query above supplied the exact key.
	record := dynamodbrecord.NewDeliveryItemRecord(*current)
	_, err = r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
		TableName:                sharedaws.String(r.tableName),
		Key:                      sharedaws.NewPrimaryKey(key, record.SK).AttributeMap(),
		UpdateExpression:         sharedaws.String("SET #state = :inflight, FencingToken = :token, LeaseUntil = :leaseUntil, AttemptCount = :attempts, LastAttemptAt = :attemptAt, UpdatedAt = :updatedAt"),
		ConditionExpression:      sharedaws.String("TenantID = :tenant AND IncidentID = :incident AND (#state = :pending OR #state = :retryable OR #state = :ambiguous OR (#state = :inflight AND LeaseUntil <= :now))"),
		ExpressionAttributeNames: map[string]string{"#state": "State"},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":tenant": &sharedaws.AttributeValueMemberS{Value: tenantID}, ":incident": &sharedaws.AttributeValueMemberS{Value: incidentID},
			":pending": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryPending)}, ":retryable": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryRetryable)},
			":ambiguous": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryAmbiguous)}, ":inflight": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryInFlight)},
			":now": &sharedaws.AttributeValueMemberS{Value: now.UTC().Format(time.RFC3339)}, ":token": &sharedaws.AttributeValueMemberS{Value: token},
			":leaseUntil": &sharedaws.AttributeValueMemberS{Value: leaseUntil}, ":attempts": &sharedaws.AttributeValueMemberN{Value: fmt.Sprintf("%d", current.AttemptCount+1)},
			":attemptAt": &sharedaws.AttributeValueMemberS{Value: now.UTC().Format(time.RFC3339)}, ":updatedAt": &sharedaws.AttributeValueMemberS{Value: now.UTC().Format(time.RFC3339)},
		},
	})
	if sharedaws.IsConditionalCheckFailure(err) {
		return current, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	current.State = notifications.DeliveryInFlight
	current.FencingToken = token
	current.LeaseUntil = leaseUntil
	current.AttemptCount++
	current.LastAttemptAt = now.UTC().Format(time.RFC3339)
	current.UpdatedAt = current.LastAttemptAt
	return current, token, nil
}

func (r *dynamoEscalationRepository) CompleteDelivery(ctx context.Context, delivery notifications.DeliveryRecord, outcome notifications.SendOutcome, nextAttemptAt string) error {
	if err := outcome.Validate(); err != nil {
		return err
	}
	record := dynamodbrecord.NewDeliveryItemRecord(delivery)
	state := notifications.DeliveryTerminalFailed
	if outcome.Class == notifications.OutcomeAccepted {
		state = notifications.DeliveryDelivered
	} else if outcome.Retryable {
		state = notifications.DeliveryRetryable
	}
	_, err := r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
		TableName: sharedaws.String(r.tableName), Key: sharedaws.NewPrimaryKey(record.PK, record.SK).AttributeMap(),
		UpdateExpression:         sharedaws.String("SET #state = :state, LastOutcomeClass = :outcome, ProviderStatusClass = :status, ProviderRequestID = :requestID, RetryAfterSeconds = :retryAfter, NextAttemptAt = :next, UpdatedAt = :updated REMOVE LeaseUntil"),
		ConditionExpression:      sharedaws.String("FencingToken = :token AND #state = :inflight"),
		ExpressionAttributeNames: map[string]string{"#state": "State"},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":token": &sharedaws.AttributeValueMemberS{Value: delivery.FencingToken}, ":inflight": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryInFlight)},
			":state": &sharedaws.AttributeValueMemberS{Value: string(state)}, ":outcome": &sharedaws.AttributeValueMemberS{Value: string(outcome.Class)},
			":status": &sharedaws.AttributeValueMemberS{Value: outcome.Metadata.ProviderStatusClass}, ":requestID": &sharedaws.AttributeValueMemberS{Value: outcome.Metadata.ProviderRequestID},
			":retryAfter": &sharedaws.AttributeValueMemberN{Value: fmt.Sprintf("%d", outcome.Metadata.RetryAfterSeconds)}, ":next": &sharedaws.AttributeValueMemberS{Value: nextAttemptAt},
		},
	})
	return err
}

// PrepareReplay makes delivery reset, replay idempotency, canonical dispatch,
// and sparse pending recovery one atomic write. Existing keys converge on the
// stored result and never create another replay.
func (r *dynamoEscalationRepository) PrepareReplay(ctx context.Context, command notifications.ReplayCommand, fingerprint string, now time.Time, retention time.Duration) (string, error) {
	if err := dynamodbrecord.ValidateReplayIdempotencyKey(command.IdempotencyKey); err != nil {
		return "", err
	}
	address := dynamodbrecord.ReplayIdempotencyAddress(command.TenantID, command.IncidentID, command.DeliveryID, command.IdempotencyKey)
	idempotencyKey := dynamodbschema.ReplayIdempotencyItem(command.TenantID, address).SK
	if existing, err := r.getReplayIdempotency(ctx, command.TenantID, idempotencyKey); err != nil {
		return "", err
	} else if existing != nil {
		if !existing.MatchesFingerprint(fingerprint) {
			return "", fmt.Errorf("idempotency key conflicts with prior replay request")
		}
		return existing.ResultDeliveryID, nil
	}

	deliveries, err := r.ListIncidentDeliveries(ctx, command.TenantID, command.IncidentID)
	if err != nil {
		return "", err
	}
	var delivery *notifications.DeliveryRecord
	for i := range deliveries {
		if strings.EqualFold(deliveries[i].DeliveryID, command.DeliveryID) {
			delivery = &deliveries[i]
			break
		}
	}
	if delivery == nil || delivery.State != notifications.DeliveryTerminalFailed {
		return "", fmt.Errorf("delivery is not eligible for replay")
	}

	replayPrefix := fingerprint
	if len(replayPrefix) > 16 {
		replayPrefix = replayPrefix[:16]
	}
	replayID := "REPLAY_" + strings.ToUpper(replayPrefix)
	deliveryItem := dynamodbrecord.NewDeliveryItemRecord(*delivery)
	idempotency := dynamodbrecord.NewReplayIdempotencyItemRecord(notifications.ReplayIdempotencyRecord{
		TenantID: command.TenantID, IncidentID: command.IncidentID, DeliveryID: command.DeliveryID,
		Operation: "delivery_replay", IdempotencyKey: command.IdempotencyKey, RequestFingerprint: fingerprint,
		ResultDeliveryID: command.DeliveryID, CreatedAt: now.UTC().Format(time.RFC3339), ExpiresAt: now.Add(retention).Unix(),
	})
	outbox := dynamodbrecord.TransitionOutboxRecord{
		PK: dynamodbschema.TenantPK(command.TenantID), SK: "TRANSITION_OUTBOX#" + dynamodbschema.NormalizeToken(replayID),
		EntityType: dynamodbschema.EntityTransitionOutbox, TenantID: command.TenantID, EventID: replayID, TransitionID: command.TransitionID,
		IncidentID: command.IncidentID, DispatchStatus: dynamodbrecord.DispatchPending, Version: notifications.CanonicalEnvelopeVersion,
		Kind: notifications.CanonicalKindReplay, SourceKind: notifications.CanonicalSourceReplay, DeliveryID: command.DeliveryID,
		CreatedAt: now.UTC().Format(time.RFC3339),
	}
	pending := dynamodbrecord.NewDispatchPendingRecord(command.TenantID, replayID, now.UTC().Format(dynamodbrecord.DispatchPendingBucketFormat), "00")
	idempotencyAV, err := sharedaws.MarshalMap(idempotency)
	if err != nil {
		return "", err
	}
	outboxAV, err := sharedaws.MarshalMap(outbox)
	if err != nil {
		return "", err
	}
	pendingAV, err := sharedaws.MarshalMap(pending)
	if err != nil {
		return "", err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: []sharedaws.TransactWriteItem{
		{Update: &sharedaws.Update{TableName: sharedaws.String(r.tableName), Key: sharedaws.NewPrimaryKey(deliveryItem.PK, deliveryItem.SK).AttributeMap(),
			UpdateExpression:         sharedaws.String("SET #state = :pending, ReplayCount = if_not_exists(ReplayCount, :zero) + :one, UpdatedAt = :updated"),
			ConditionExpression:      sharedaws.String("TenantID = :tenant AND IncidentID = :incident AND DeliveryID = :delivery AND #state = :terminal"),
			ExpressionAttributeNames: map[string]string{"#state": "State"}, ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
				":tenant": &sharedaws.AttributeValueMemberS{Value: command.TenantID}, ":incident": &sharedaws.AttributeValueMemberS{Value: command.IncidentID}, ":delivery": &sharedaws.AttributeValueMemberS{Value: command.DeliveryID},
				":terminal": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryTerminalFailed)}, ":pending": &sharedaws.AttributeValueMemberS{Value: string(notifications.DeliveryPending)},
				":zero": &sharedaws.AttributeValueMemberN{Value: "0"}, ":one": &sharedaws.AttributeValueMemberN{Value: "1"}, ":updated": &sharedaws.AttributeValueMemberS{Value: now.UTC().Format(time.RFC3339)},
			}}},
		{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: idempotencyAV, ConditionExpression: sharedaws.String("attribute_not_exists(PK)")}},
		{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: outboxAV, ConditionExpression: sharedaws.String("attribute_not_exists(PK)")}},
		{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: pendingAV, ConditionExpression: sharedaws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)")}},
	}})
	if err != nil {
		return "", err
	}
	return command.DeliveryID, nil
}

func (r *dynamoEscalationRepository) getReplayIdempotency(ctx context.Context, tenantID, sortKey string) (*dynamodbrecord.ReplayIdempotencyItemRecord, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(r.tableName), Key: sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), sortKey).AttributeMap()})
	if err != nil || len(out.Item) == 0 {
		return nil, err
	}
	var record dynamodbrecord.ReplayIdempotencyItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *dynamoEscalationRepository) SuppressEscalation(ctx context.Context, tenantID, incidentID string, updatedAt string) error {
	_, err := r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
		TableName:                sharedaws.String(r.tableName),
		Key:                      sharedaws.NewPrimaryKey(dynamodbschema.IncidentPK(incidentID), "ESCALATION_STATE").AttributeMap(),
		UpdateExpression:         sharedaws.String("SET #status = :suppressed, UpdatedAt = :updated"),
		ConditionExpression:      sharedaws.String("TenantID = :tenant AND IncidentID = :incident AND #status = :active"),
		ExpressionAttributeNames: map[string]string{"#status": "Status"},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":tenant": &sharedaws.AttributeValueMemberS{Value: tenantID}, ":incident": &sharedaws.AttributeValueMemberS{Value: incidentID},
			":active": &sharedaws.AttributeValueMemberS{Value: "active"}, ":suppressed": &sharedaws.AttributeValueMemberS{Value: "suppressed"},
			":updated": &sharedaws.AttributeValueMemberS{Value: updatedAt},
		},
	})
	if sharedaws.IsConditionalCheckFailure(err) {
		return nil
	}
	return err
}

func (r *dynamoEscalationRepository) AdvanceStepOnce(ctx context.Context, tenantID, incidentID string, expectedStep, nextStep int, updatedAt string) error {
	_, err := r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
		TableName:                sharedaws.String(r.tableName),
		Key:                      sharedaws.NewPrimaryKey(dynamodbschema.IncidentPK(incidentID), "ESCALATION_STATE").AttributeMap(),
		UpdateExpression:         sharedaws.String("SET CurrentStep = :next, UpdatedAt = :updated"),
		ConditionExpression:      sharedaws.String("TenantID = :tenant AND IncidentID = :incident AND #status = :active AND CurrentStep = :expected"),
		ExpressionAttributeNames: map[string]string{"#status": "Status"},
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":tenant": &sharedaws.AttributeValueMemberS{Value: tenantID}, ":incident": &sharedaws.AttributeValueMemberS{Value: incidentID},
			":active": &sharedaws.AttributeValueMemberS{Value: "active"}, ":expected": &sharedaws.AttributeValueMemberN{Value: fmt.Sprintf("%d", expectedStep)},
			":next": &sharedaws.AttributeValueMemberN{Value: fmt.Sprintf("%d", nextStep)}, ":updated": &sharedaws.AttributeValueMemberS{Value: updatedAt},
		},
	})
	if sharedaws.IsConditionalCheckFailure(err) {
		return nil
	}
	return err
}

func newFencingToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}
