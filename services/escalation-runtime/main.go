package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"bolt-monitor/shared/aws"
)

func main() {
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatalf("TABLE_NAME is required")
	}

	dynamoClient, err := aws.NewDynamoDBClient(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}
	repo := newDynamoEscalationRepository(dynamoClient, tableName)
	scheduler := newCloudWatchScheduler(
		eventbridge.NewFromConfig(awsCfg),
		awslambda.NewFromConfig(awsCfg),
		sts.NewFromConfig(awsCfg),
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
	)
	handler := newEscalationHandler(repo, scheduler)

	lambda.Start(func(ctx context.Context, payload json.RawMessage) error {
		var sqsEvent events.SQSEvent
		if err := json.Unmarshal(payload, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
			return handler.handleSQSEvent(ctx, sqsEvent)
		}
		var scheduled scheduledInvocationEvent
		if err := json.Unmarshal(payload, &scheduled); err == nil && scheduled.IncidentID != "" {
			return handler.handleScheduledInvocation(ctx, scheduled)
		}
		log.Printf("ignoring unsupported escalation-runtime payload")
		return nil
	})
}
