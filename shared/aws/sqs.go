package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSAPI interface {
	SendMessage(ctx context.Context, params *SQSSendMessageInput) (*SQSSendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *SQSReceiveMessageInput) (*SQSReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *SQSDeleteMessageInput) (*SQSDeleteMessageOutput, error)
	ChangeMessageVisibility(ctx context.Context, params *SQSChangeMessageVisibilityInput) (*SQSChangeMessageVisibilityOutput, error)
}

type SQSSendMessageInput = sqs.SendMessageInput
type SQSSendMessageOutput = sqs.SendMessageOutput
type SQSReceiveMessageInput = sqs.ReceiveMessageInput
type SQSReceiveMessageOutput = sqs.ReceiveMessageOutput
type SQSDeleteMessageInput = sqs.DeleteMessageInput
type SQSDeleteMessageOutput = sqs.DeleteMessageOutput
type SQSChangeMessageVisibilityInput = sqs.ChangeMessageVisibilityInput
type SQSChangeMessageVisibilityOutput = sqs.ChangeMessageVisibilityOutput
type SQSMessage = sqstypes.Message

type sqsAPI struct {
	client *sqs.Client
}

func NewSQS(client *sqs.Client) SQSAPI {
	return &sqsAPI{client: client}
}

func (s *sqsAPI) SendMessage(ctx context.Context, params *SQSSendMessageInput) (*SQSSendMessageOutput, error) {
	return s.client.SendMessage(ctx, params)
}

func (s *sqsAPI) ReceiveMessage(ctx context.Context, params *SQSReceiveMessageInput) (*SQSReceiveMessageOutput, error) {
	return s.client.ReceiveMessage(ctx, params)
}

func (s *sqsAPI) DeleteMessage(ctx context.Context, params *SQSDeleteMessageInput) (*SQSDeleteMessageOutput, error) {
	return s.client.DeleteMessage(ctx, params)
}

func (s *sqsAPI) ChangeMessageVisibility(ctx context.Context, params *SQSChangeMessageVisibilityInput) (*SQSChangeMessageVisibilityOutput, error) {
	return s.client.ChangeMessageVisibility(ctx, params)
}
