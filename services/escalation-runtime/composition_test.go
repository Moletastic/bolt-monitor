package main

import (
	"context"
	"testing"

	sharedaws "bolt-monitor/shared/aws"
)

type compositionSQSClient struct{}

func (compositionSQSClient) SendMessage(context.Context, *sharedaws.SQSSendMessageInput) (*sharedaws.SQSSendMessageOutput, error) {
	return &sharedaws.SQSSendMessageOutput{}, nil
}

func (compositionSQSClient) ReceiveMessage(context.Context, *sharedaws.SQSReceiveMessageInput) (*sharedaws.SQSReceiveMessageOutput, error) {
	return &sharedaws.SQSReceiveMessageOutput{}, nil
}

func (compositionSQSClient) DeleteMessage(context.Context, *sharedaws.SQSDeleteMessageInput) (*sharedaws.SQSDeleteMessageOutput, error) {
	return &sharedaws.SQSDeleteMessageOutput{}, nil
}

func (compositionSQSClient) ChangeMessageVisibility(context.Context, *sharedaws.SQSChangeMessageVisibilityInput) (*sharedaws.SQSChangeMessageVisibilityOutput, error) {
	return &sharedaws.SQSChangeMessageVisibilityOutput{}, nil
}

type compositionSchedulerClient struct{}

func (compositionSchedulerClient) CreateSchedule(context.Context, *sharedaws.SchedulerCreateScheduleInput) (*sharedaws.SchedulerCreateScheduleOutput, error) {
	return &sharedaws.SchedulerCreateScheduleOutput{}, nil
}

func (compositionSchedulerClient) UpdateSchedule(context.Context, *sharedaws.SchedulerUpdateScheduleInput) (*sharedaws.SchedulerUpdateScheduleOutput, error) {
	return &sharedaws.SchedulerUpdateScheduleOutput{}, nil
}

func (compositionSchedulerClient) DeleteSchedule(context.Context, *sharedaws.SchedulerDeleteScheduleInput) (*sharedaws.SchedulerDeleteScheduleOutput, error) {
	return &sharedaws.SchedulerDeleteScheduleOutput{}, nil
}

func (compositionSchedulerClient) GetSchedule(context.Context, *sharedaws.SchedulerGetScheduleInput) (*sharedaws.SchedulerGetScheduleOutput, error) {
	return &sharedaws.SchedulerGetScheduleOutput{}, nil
}

func TestNewProductionEscalationRuntimeAssemblesDependenciesWithoutAWS(t *testing.T) {
	handler, adapter, dispatcher := newProductionEscalationRuntime(&fakeDynamoClient{}, compositionSQSClient{}, compositionSchedulerClient{}, escalationRuntimeConfig{TableName: "app-table", NotificationQueueURL: "notification-queue"})
	if handler.repo == nil || handler.senders == nil || handler.now == nil || adapter == nil || dispatcher == nil {
		t.Fatal("production escalation runtime has missing dependency")
	}
}
