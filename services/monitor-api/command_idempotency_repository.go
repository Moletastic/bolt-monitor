package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbschema"
)

const commandIdempotencyRetention = 24 * 7 * 60 * 60

type commandIdempotencyStore interface {
	ReserveCommandIdempotency(context.Context, commandIdempotencyRecord) (commandIdempotencyRecord, error)
	CompleteCommandIdempotency(context.Context, commandIdempotencyRecord, string) error
}

func (r *dynamoMonitorRepository) ReserveCommandIdempotency(ctx context.Context, record commandIdempotencyRecord) (commandIdempotencyRecord, error) {
	if record.TenantID == "" || record.Operation == "" || record.ResourceID == "" || record.Key == "" || record.Fingerprint == "" {
		return commandIdempotencyRecord{}, fmt.Errorf("command idempotency record is incomplete")
	}
	if err := r.requireTableName(); err != nil {
		return commandIdempotencyRecord{}, err
	}
	key := sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(record.TenantID), commandIdempotencyAddress(record.TenantID, record.Operation, record.ResourceID, record.Key))
	item := key.AttributeMap()
	for name, value := range map[string]string{"Operation": record.Operation, "ResourceID": record.ResourceID, "Key": record.Key, "Fingerprint": record.Fingerprint, "ReservationToken": record.ReservationToken, "State": string(record.State), "CreatedAt": record.CreatedAt.Format(time.RFC3339)} {
		item[name] = &sharedaws.AttributeValueMemberS{Value: value}
	}
	item["TTL"] = &sharedaws.AttributeValueMemberN{Value: strconv.FormatInt(record.TTL, 10)}
	_, err := r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: item, ConditionExpression: sharedaws.String("attribute_not_exists(PK) OR TTL <= :now"), ExpressionAttributeValues: map[string]sharedaws.AttributeValue{":now": &sharedaws.AttributeValueMemberN{Value: strconv.FormatInt(record.CreatedAt.Unix(), 10)}}})
	if err == nil {
		return record, nil
	}
	if !sharedaws.IsConditionalCheckFailure(err) {
		return commandIdempotencyRecord{}, err
	}
	existing, found, err := r.loadCommandIdempotency(ctx, record)
	if err != nil || !found {
		return commandIdempotencyRecord{}, err
	}
	return existing, nil
}

func (r *dynamoMonitorRepository) loadCommandIdempotency(ctx context.Context, input commandIdempotencyRecord) (commandIdempotencyRecord, bool, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName:      sharedaws.String(r.tableName),
		Key:            sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(input.TenantID), commandIdempotencyAddress(input.TenantID, input.Operation, input.ResourceID, input.Key)).AttributeMap(),
		ConsistentRead: sharedaws.Bool(true),
	})
	if err != nil || len(out.Item) == 0 {
		return commandIdempotencyRecord{}, len(out.Item) != 0, err
	}
	return commandIdempotencyRecord{TenantID: input.TenantID, Operation: attrString(out.Item, "Operation"), ResourceID: attrString(out.Item, "ResourceID"), Key: attrString(out.Item, "Key"), Fingerprint: attrString(out.Item, "Fingerprint"), ReservationToken: attrString(out.Item, "ReservationToken"), Response: attrString(out.Item, "Response"), State: commandIdempotencyState(attrString(out.Item, "State")), CreatedAt: parseTimeOrZero(attrString(out.Item, "CreatedAt")), TTL: ttlNumber(out.Item)}, true, nil
}

func (r *dynamoMonitorRepository) CompleteCommandIdempotency(ctx context.Context, record commandIdempotencyRecord, publicResponse string) error {
	_, err := r.client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{TableName: sharedaws.String(r.tableName), Key: sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(record.TenantID), commandIdempotencyAddress(record.TenantID, record.Operation, record.ResourceID, record.Key)).AttributeMap(), UpdateExpression: sharedaws.String("SET #state = :state, Response = :response"), ConditionExpression: sharedaws.String("Fingerprint = :fingerprint"), ExpressionAttributeNames: map[string]string{"#state": "State"}, ExpressionAttributeValues: map[string]sharedaws.AttributeValue{":state": &sharedaws.AttributeValueMemberS{Value: string(commandIdempotencyCompleted)}, ":response": &sharedaws.AttributeValueMemberS{Value: publicResponse}, ":fingerprint": &sharedaws.AttributeValueMemberS{Value: record.Fingerprint}}})
	return err
}
