package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/aws"
)

func main() {
	ctx := context.Background()
	dynamoClient, err := aws.NewDynamoDBAPI(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}

	repo := newDynamoMonitorRepository(dynamoClient, os.Getenv("TABLE_NAME"))
	principalResolver := newCognitoPrincipalResolver(strings.Split(os.Getenv("COGNITO_CLIENT_IDS"), ","))
	membershipResolver := newAuthTableMembershipResolver(dynamoClient, os.Getenv("AUTH_TABLE_NAME"))
	handler := newAuthorizedMonitorHandler(repo, principalResolver, membershipResolver)

	lambda.Start(handler.handleRequest)
}
