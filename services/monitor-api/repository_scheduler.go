package main

import (
	"context"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
)

// SchedulerStore is the narrow interface used by the admin scheduler view.
// Admin handlers do not need monitor or incident reads so they consume only
// the scheduler methods.
type SchedulerStore interface {
	GetSchedulerConfig(context.Context, string) (dynamodbrecord.SchedulerConfigRecord, error)
	UpdateSchedulerConfig(context.Context, string, checkexecution.SchedulerConfig, time.Time) (dynamodbrecord.SchedulerConfigRecord, error)
}

func (r *dynamoMonitorRepository) GetSchedulerConfig(ctx context.Context, tenantID string) (dynamodbrecord.SchedulerConfigRecord, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName: sharedaws.String(r.tableName),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			"SK": &sharedaws.AttributeValueMemberS{Value: "SCHEDULER_CONFIG"},
		},
	})
	if err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	if len(out.Item) == 0 {
		return dynamodbrecord.SchedulerConfigRecord{Config: checkexecution.SchedulerConfig{}}, nil
	}
	var record dynamodbrecord.SchedulerConfigItemRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	return record.ToSchedulerConfig(), nil
}

func (r *dynamoMonitorRepository) UpdateSchedulerConfig(ctx context.Context, tenantID string, config checkexecution.SchedulerConfig, now time.Time) (dynamodbrecord.SchedulerConfigRecord, error) {
	if err := r.requireTableName(); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	record := dynamodbrecord.NewSchedulerConfigItemRecord(tenantID, config, now)
	items, err := marshalPutItems(r.tableName, record)
	if err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	if err := r.writeTransaction(ctx, items); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	return record.ToSchedulerConfig(), nil
}
