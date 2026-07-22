package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/aws"
	"bolt-monitor/shared/outboundhttp"
)

const (
	defaultTenantID = "DEFAULT"
	modeWorker      = "worker"
	modeScheduler   = "scheduler"
)

func newProductionRuntimeHandler(dynamoClient aws.DynamoDBAPI, sqsClient aws.SQSAPI, config runtimeConfig) runtimeHandler {
	repo := newDynamoRuntimeRepositoryWithLease(dynamoClient, config.TableName, config.WorkLeaseDuration)
	return newRuntimeHandlerWithDependencies(
		repo,
		newAWSSQSClient(sqsClient),
		config.ExecutionQueueURL,
		config.EscalationQueueURL,
		defaultTenantID,
		config.Mode,
		runtimeHandlerDependencies{
			now:               time.Now,
			executor:          outboundhttp.NewExecutor(),
			resultClock:       systemExecutionResultClock{},
			resultIDs:         generatedExecutionResultIDs{},
			schedulerDeadline: defaultSchedulerDeadline,
		},
	)
}

func main() {
	ctx := context.Background()
	config, err := newRuntimeConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	dynamoClient, err := aws.NewDynamoDBAPI(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}
	sqsClient, err := aws.NewSQSAPI(ctx)
	if err != nil {
		log.Fatalf("create sqs client: %v", err)
	}

	handler := newProductionRuntimeHandler(dynamoClient, sqsClient, config)

	switch config.Mode {
	case modeScheduler:
		lambda.Start(func(ctx context.Context, event events.CloudWatchEvent) (runtimeSummary, error) {
			return handler.handle(ctx, event)
		})
	case modeWorker:
		lambda.Start(func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
			return handler.handleSQSEventBatch(ctx, event)
		})
	default:
		log.Fatalf("unsupported runtime mode %q", config.Mode)
	}
}
