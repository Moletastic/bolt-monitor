package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// NewDynamoDBClient returns a DynamoDB client configured from the default AWS config.
func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// NewDynamoDBAPI returns a DynamoDB facade configured from the default AWS config.
func NewDynamoDBAPI(ctx context.Context) (DynamoDBAPI, error) {
	client, err := NewDynamoDBClient(ctx)
	if err != nil {
		return nil, err
	}
	return NewDynamoDB(client), nil
}

// NewCognitoIdentityProviderClient returns a Cognito client configured from the default AWS config.
func NewCognitoIdentityProviderClient(ctx context.Context) (*cognitoidentityprovider.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return cognitoidentityprovider.NewFromConfig(cfg), nil
}

// NewCognitoIdentityProviderAPI returns a credentialed Cognito administration facade.
func NewCognitoIdentityProviderAPI(ctx context.Context) (CognitoIdentityProviderAPI, error) {
	client, err := NewCognitoIdentityProviderClient(ctx)
	if err != nil {
		return nil, err
	}
	return NewCognitoIdentityProvider(client), nil
}

// NewEventBridgeClient returns an EventBridge client configured from the default AWS config.
func NewEventBridgeClient(ctx context.Context) (*eventbridge.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return eventbridge.NewFromConfig(cfg), nil
}

// NewEventBridgeAPI returns an EventBridge facade configured from the default AWS config.
func NewEventBridgeAPI(ctx context.Context) (EventBridgeAPI, error) {
	client, err := NewEventBridgeClient(ctx)
	if err != nil {
		return nil, err
	}
	return NewEventBridge(client), nil
}

// NewSQSClient returns an SQS client configured from the default AWS config.
func NewSQSClient(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return sqs.NewFromConfig(cfg), nil
}

// NewSQSAPI returns an SQS facade configured from the default AWS config.
func NewSQSAPI(ctx context.Context) (SQSAPI, error) {
	client, err := NewSQSClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return NewSQS(client), nil
}

// NewSchedulerClient returns a Scheduler client configured from the default AWS config.
func NewSchedulerClient(ctx context.Context) (*scheduler.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return scheduler.NewFromConfig(cfg), nil
}

// NewSchedulerAPI returns a Scheduler facade configured from the default AWS config.
func NewSchedulerAPI(ctx context.Context) (SchedulerAPI, error) {
	client, err := NewSchedulerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return NewScheduler(client), nil
}
