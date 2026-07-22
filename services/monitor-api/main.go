//go:build !inline_channel_migration

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/aws"
	"bolt-monitor/shared/notifications"
	"bolt-monitor/shared/outboundhttp"
)

func newProductionMonitorHandler(dynamoClient aws.DynamoDBAPI, config monitorAPIConfig) monitorHandler {
	executor := outboundhttp.NewExecutor()
	repository := newDynamoMonitorRepository(dynamoClient, config.TableName)
	return newAuthorizedMonitorHandlerWithDependencies(
		newMonitorAPIOperations(repository, repository, repository, repository, repository, repository, repository, repository, repository),
		newCognitoPrincipalResolver(config.CognitoClientIDs),
		newAuthTableMembershipResolver(dynamoClient, config.AuthTableName),
		monitorHandlerDependencies{
			securityEvents:   emitMonitorSecurityEvent,
			newSecurityEvent: newMonitorSecurityEventFactory(config.SecurityEventStage, time.Now),
			now:              time.Now,
			senders:          notifications.NewSenderRegistry(),
			executor:         executor,
			validateDestination: func(ctx context.Context, target string) error {
				_, _, err := executor.ValidateDestination(ctx, target)
				return err
			},
		},
	)
}

func main() {
	ctx := context.Background()
	config, err := newMonitorAPIConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	dynamoClient, err := aws.NewDynamoDBAPI(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}

	handler := newProductionMonitorHandler(dynamoClient, config)

	lambda.Start(handler.handleRequest)
}
