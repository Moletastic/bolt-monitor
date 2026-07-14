package main

import (
	"context"
	"log"
	"os"

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
	handler := newMonitorHandler(repo, defaultTenantID)

	lambda.Start(handler.handleRequest)
}
