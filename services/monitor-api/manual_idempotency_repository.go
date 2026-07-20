package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbschema"
)

const manualIdempotencyRetentionSeconds = manualIdempotencyRetentionDays * 24 * 60 * 60

// ReserveManualIdempotency conditionally persists a record at the deterministic
// address. It returns the existing record when one already exists so the
// handler can resume or detect a conflict.
func (r *dynamoMonitorRepository) ReserveManualIdempotency(ctx context.Context, record manualIdempotencyRecord) (manualIdempotencyRecord, error) {
	if err := record.validate(); err != nil {
		return manualIdempotencyRecord{}, err
	}
	if err := r.requireTableName(); err != nil {
		return manualIdempotencyRecord{}, err
	}
	item := sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(record.TenantID), manualIdempotencyAddress(record.TenantID, record.ServiceID, record.MonitorID, record.Key))
	itemMap := item.AttributeMap()
	itemMap["Fingerprint"] = &sharedaws.AttributeValueMemberS{Value: record.Fingerprint}
	itemMap["Outcome"] = &sharedaws.AttributeValueMemberS{Value: string(record.Outcome)}
	itemMap["RunID"] = &sharedaws.AttributeValueMemberS{Value: record.RunID}
	itemMap["ServiceID"] = &sharedaws.AttributeValueMemberS{Value: record.ServiceID}
	itemMap["MonitorID"] = &sharedaws.AttributeValueMemberS{Value: record.MonitorID}
	itemMap["Key"] = &sharedaws.AttributeValueMemberS{Value: record.Key}
	itemMap["CreatedAt"] = &sharedaws.AttributeValueMemberS{Value: record.CreatedAt.Format(time.RFC3339)}
	_, err := r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{
		TableName:           sharedaws.String(r.tableName),
		Item:                itemMap,
		ConditionExpression: sharedaws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	})
	if err == nil {
		return record, nil
	}
	if !sharedaws.IsConditionalCheckFailure(err) {
		return manualIdempotencyRecord{}, err
	}
	existing, found, loadErr := r.LoadManualIdempotency(ctx, record.TenantID, record.ServiceID, record.MonitorID, record.Key)
	if loadErr != nil {
		return manualIdempotencyRecord{}, loadErr
	}
	if !found {
		return manualIdempotencyRecord{}, fmt.Errorf("idempotency record missing after conditional failure")
	}
	return existing, nil
}

// LoadManualIdempotency returns the existing record at the scoped address.
func (r *dynamoMonitorRepository) LoadManualIdempotency(ctx context.Context, tenantID, serviceID, monitorID, key string) (manualIdempotencyRecord, bool, error) {
	if err := r.requireTableName(); err != nil {
		return manualIdempotencyRecord{}, false, err
	}
	item, found, err := sharedaws.GetByPrimaryKey(ctx, r.client, r.tableName, sharedaws.NewPrimaryKey(dynamodbschema.TenantPK(tenantID), manualIdempotencyAddress(tenantID, serviceID, monitorID, key)))
	if err != nil || !found {
		return manualIdempotencyRecord{}, found, err
	}
	fingerprint := attrString(item, "Fingerprint")
	outcome := attrString(item, "Outcome")
	runID := attrString(item, "RunID")
	serviceIDValue := attrString(item, "ServiceID")
	monitorIDValue := attrString(item, "MonitorID")
	rawKey := attrString(item, "Key")
	createdAt := attrString(item, "CreatedAt")
	ttl := ttlNumber(item)
	record := manualIdempotencyRecord{
		TenantID:    tenantID,
		ServiceID:   serviceIDValue,
		MonitorID:   monitorIDValue,
		Key:         rawKey,
		Fingerprint: fingerprint,
		Outcome:     manualIdempotencyOutcome(outcome),
		RunID:       runID,
		CreatedAt:   parseTimeOrZero(createdAt),
		TTL:         ttl,
	}
	return record, true, nil
}

func attrString(item map[string]sharedaws.AttributeValue, key string) string {
	value, ok := item[key].(*sharedaws.AttributeValueMemberS)
	if !ok {
		return ""
	}
	return value.Value
}

func ttlNumber(item map[string]sharedaws.AttributeValue) int64 {
	value, ok := item["TTL"].(*sharedaws.AttributeValueMemberN)
	if !ok {
		return 0
	}
	parsed, err := strconv.ParseInt(value.Value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseTimeOrZero(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
