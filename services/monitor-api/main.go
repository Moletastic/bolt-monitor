//go:build !inline_channel_migration

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/aws"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/notifications"
	"bolt-monitor/shared/outboundhttp"
)

func newProductionMonitorHandler(dynamoClient aws.DynamoDBAPI, config monitorAPIConfig) monitorHandler {
	executor := outboundhttp.NewExecutor()
	now := commandClock(time.Now)
	ids := productionIdentifierGenerator()
	repository := newDynamoMonitorRepository(dynamoClient, config.TableName)
	services := newServiceOperations(repository, repository, repository, repository, repository, repository, repository, repository, now, ids)
	monitors := newMonitorOperations(repository, repository, repository, repository, repository, repository, repository, repository, repository, repository, repository, repository, now, ids, executor, func(ctx context.Context, monitor monitorconfig.Monitor) error {
		if monitor.HTTP == nil {
			return nil
		}
		_, _, err := executor.ValidateDestination(ctx, monitor.HTTP.Target)
		if err != nil {
			return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.target", "reason": outboundhttp.SafeMessage(err)})
		}
		return nil
	})
	incidents := newIncidentOperations(repository, repository, repository, repository, repository, repository, repository, repository, repository, repository, now)
	scheduler := newSchedulerOperations(repository, repository, now)
	channels := newNotificationChannelOperations(repository, repository, repository, repository, repository, repository, notifications.NewSenderRegistry(), now, ids)
	policies := newEscalationPolicyOperations(repository, repository, repository, repository, repository, repository, repository, repository, now, ids)
	return newAuthorizedMonitorHandlerWithDependencies(
		newMonitorAPIOperations(services, monitors, incidents, scheduler, policies, channels, searchResourcesQuery{store: repository}),
		newCognitoPrincipalResolver(config.CognitoClientIDs),
		newAuthTableMembershipResolver(dynamoClient, config.AuthTableName),
		monitorHandlerDependencies{
			securityEvents:   emitMonitorSecurityEvent,
			newSecurityEvent: newMonitorSecurityEventFactory(config.SecurityEventStage, time.Now),
			tenantID:         defaultTenantID,
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
