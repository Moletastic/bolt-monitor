package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/aws"
	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/notifications"
)

func main() {
	ctx := context.Background()
	config, err := newEscalationRuntimeConfig(os.Getenv)
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
	schedulerClient, err := aws.NewSchedulerAPI(ctx)
	if err != nil {
		log.Fatalf("create scheduler client: %v", err)
	}

	handler, adapter, dispatcher := newProductionEscalationRuntime(dynamoClient, sqsClient, schedulerClient, config)

	lambda.Start(func(ctx context.Context, payload json.RawMessage) (any, error) {
		var sqsEvent events.SQSEvent
		if err := json.Unmarshal(payload, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
			response, err := handler.handleSQSEventResponse(ctx, sqsEvent)
			return response, err
		}
		var streamEvent events.DynamoDBEvent
		if err := json.Unmarshal(payload, &streamEvent); err == nil && len(streamEvent.Records) > 0 {
			return dispatcher.handle(ctx, streamEvent)
		}
		var scheduled scheduledInvocationEvent
		if err := json.Unmarshal(payload, &scheduled); err == nil && scheduled.IncidentID != "" {
			if adapter != nil {
				_ = adapter.Reenqueue(ctx, scheduled)
			}
			return nil, handler.handleScheduledInvocation(ctx, scheduled)
		}
		log.Printf("ignoring unsupported escalation-runtime payload")
		return nil, nil
	})
}

func newProductionEscalationRuntime(dynamoClient sharedaws.DynamoDBAPI, sqsClient sharedaws.SQSAPI, schedulerClient sharedaws.SchedulerAPI, config escalationRuntimeConfig) (*escalationHandler, *legacyScheduleAdapter, *streamDispatcher) {
	repo := newDynamoEscalationRepository(dynamoClient, config.TableName)
	handler := newEscalationHandlerWithDependencies(repo, buildScheduler(schedulerClient, config), escalationHandlerDependencies{
		senders: notifications.NewSenderRegistry(),
		now:     time.Now,
	})
	return handler, newLegacyScheduleAdapter(sqsClient, config.NotificationQueueURL), newStreamDispatcher(repo, sqsClient, config.NotificationQueueURL)
}

func buildScheduler(schedulerClient sharedaws.SchedulerAPI, config escalationRuntimeConfig) scheduleClient {
	group := config.ScheduleGroupName
	role := config.ScheduleExecutionRole
	queue := config.NotificationQueueARN
	dlq := config.NotificationDLQARN
	if group == "" || role == "" || queue == "" {
		log.Printf("scheduler group/role/queue missing; new schedules disabled")
		return nil
	}
	retry := sharedaws.SchedulerRetryPolicy{
		MaximumEventAgeInSeconds: aws.Int32(int32(notifications.SchedulerTargetRetryAge.Seconds())),
		MaximumRetryAttempts:     aws.Int32(3),
	}
	return newOneTimeScheduler(schedulerClient, group, role, queue, dlq, retry)
}
