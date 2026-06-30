package main

import (
	"context"

	sharedaws "bolt-monitor/shared/aws"
)

type sqsClient interface {
	SendMessage(ctx context.Context, queueURL string, body string) error
}

type awsSQSClient struct {
	client sharedaws.SQSAPI
}

func newAWSSQSClient(client sharedaws.SQSAPI) *awsSQSClient {
	return &awsSQSClient{client: client}
}

func (c *awsSQSClient) SendMessage(ctx context.Context, queueURL string, body string) error {
	_, err := c.client.SendMessage(ctx, &sharedaws.SQSSendMessageInput{
		QueueUrl:    sharedaws.String(queueURL),
		MessageBody: sharedaws.String(body),
	})
	return err
}
