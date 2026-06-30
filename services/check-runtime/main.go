package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/aws"
)

const (
	defaultTenantID = "DEFAULT"
	modeWorker      = "worker"
	modeScheduler   = "scheduler"
)

func main() {
	ctx := context.Background()
	dynamoClient, err := aws.NewDynamoDBAPI(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}
	sqsClient, err := aws.NewSQSAPI(ctx)
	if err != nil {
		log.Fatalf("create sqs client: %v", err)
	}

	awsSQSClient := newAWSSQSClient(sqsClient)
	escalationQueueURL := os.Getenv("ESCALATION_QUEUE_URL")
	handler := newRuntimeHandler(
		newDynamoRuntimeRepository(dynamoClient, os.Getenv("TABLE_NAME")),
		awsSQSClient,
		os.Getenv("EXECUTION_QUEUE_URL"),
		escalationQueueURL,
		defaultProbeLocationCatalog(),
		defaultTenantID,
		os.Getenv("RUNTIME_MODE"),
	)

	switch os.Getenv("RUNTIME_MODE") {
	case modeScheduler:
		lambda.Start(func(ctx context.Context, event events.CloudWatchEvent) (runtimeSummary, error) {
			return handler.handle(ctx, event)
		})
	case modeWorker:
		lambda.Start(func(ctx context.Context, event events.SQSEvent) (runtimeSummary, error) {
			return handler.handleSQSEvent(ctx, event)
		})
	default:
		log.Fatalf("unsupported runtime mode %q", os.Getenv("RUNTIME_MODE"))
	}
}
