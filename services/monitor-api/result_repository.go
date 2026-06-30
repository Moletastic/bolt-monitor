package main

import (
	"context"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/resultstatus"
)

func (r *dynamoMonitorRepository) PersistCheckRunAndStatus(ctx context.Context, run resultstatus.CheckRun, status resultstatus.MonitorStatus) error {
	if err := r.requireTableName(); err != nil {
		return err
	}
	runItem, err := sharedaws.MarshalMap(run.ToRecord())
	if err != nil {
		return err
	}
	statusItem, err := sharedaws.MarshalMap(status.ToRecord())
	if err != nil {
		return err
	}
	_, err = r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{
		TransactItems: []sharedaws.TransactWriteItem{
			{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: runItem}},
			{Put: &sharedaws.Put{TableName: sharedaws.String(r.tableName), Item: statusItem}},
		},
	})
	return err
}
