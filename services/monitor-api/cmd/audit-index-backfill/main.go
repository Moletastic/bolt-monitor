package main

import (
	"context"
	"errors"
	"log"
	"os"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	ctx := context.Background()
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatal("TABLE_NAME is required")
	}
	client, err := sharedaws.NewDynamoDBAPI(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}

	var startKey map[string]sharedaws.AttributeValue
	updated := 0
	for {
		out, err := client.Scan(ctx, &sharedaws.DynamoDBScanInput{
			TableName:         sharedaws.String(tableName),
			ExclusiveStartKey: startKey,
			FilterExpression:  sharedaws.String("EntityType = :audit AND (attribute_not_exists(GSI3PK) OR attribute_not_exists(GSI3SK))"),
			ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
				":audit": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.EntityAuditEvent},
			},
		})
		if err != nil {
			log.Fatalf("scan audit events: %v", err)
		}
		for _, item := range out.Items {
			var event dynamodbrecord.AuditEventRecord
			if err := sharedaws.UnmarshalMap(item, &event); err != nil {
				log.Fatalf("decode audit event: %v", err)
			}
			resource := dynamodbschema.AuditResourceItem(event.TenantID, event.ServiceID, event.MonitorID, event.AuditID, event.Timestamp)
			_, err := client.UpdateItem(ctx, &sharedaws.DynamoDBUpdateItemInput{
				TableName: sharedaws.String(tableName),
				Key: map[string]sharedaws.AttributeValue{
					"PK": &sharedaws.AttributeValueMemberS{Value: event.PK},
					"SK": &sharedaws.AttributeValueMemberS{Value: event.SK},
				},
				UpdateExpression:    sharedaws.String("SET GSI3PK = :gsi3pk, GSI3SK = :gsi3sk"),
				ConditionExpression: sharedaws.String("attribute_not_exists(GSI3PK) OR attribute_not_exists(GSI3SK)"),
				ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
					":gsi3pk": &sharedaws.AttributeValueMemberS{Value: resource.GSI3PK},
					":gsi3sk": &sharedaws.AttributeValueMemberS{Value: resource.GSI3SK},
				},
			})
			if err != nil {
				var conditionFailed *ddbtypes.ConditionalCheckFailedException
				if !errors.As(err, &conditionFailed) {
					log.Fatalf("backfill audit event: %v", err)
				}
				continue
			}
			updated++
		}
		startKey = out.LastEvaluatedKey
		if len(startKey) == 0 {
			break
		}
	}
	log.Printf("audit resource index backfill complete: updated=%d", updated)
}
