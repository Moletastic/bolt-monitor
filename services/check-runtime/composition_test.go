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

func TestNewProductionRuntimeHandlerAssemblesDependenciesWithoutAWS(t *testing.T) {
	handler := newProductionRuntimeHandler(&fakeDynamoClient{}, compositionSQSClient{}, runtimeConfig{
		TableName: "app-table", ExecutionQueueURL: "execution-queue", Mode: modeScheduler, WorkLeaseDuration: defaultExecutionWorkLeaseDuration,
	})
	if handler.repo == nil || handler.sqsClient == nil || handler.now == nil || handler.executor == nil || handler.resultCommand.clock == nil || handler.resultCommand.ids == nil {
		t.Fatal("production runtime handler has missing dependency")
	}
}
