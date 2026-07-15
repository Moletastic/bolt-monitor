package main

import (
	"context"
	"flag"
	"log"
	"time"

	sharedaws "bolt-monitor/shared/aws"
)

func main() {
	email := flag.String("email", "", "operator email")
	userPoolID := flag.String("user-pool-id", "", "Cognito user pool ID")
	authTable := flag.String("auth-table", "", "AuthTable name")
	flag.Parse()
	if *email == "" || *userPoolID == "" || *authTable == "" {
		log.Fatal("email, user-pool-id, and auth-table are required")
	}

	ctx := context.Background()
	cognito, err := sharedaws.NewCognitoIdentityProviderAPI(ctx)
	if err != nil {
		log.Fatalf("create cognito client: %v", err)
	}
	dynamo, err := sharedaws.NewDynamoDBAPI(ctx)
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}
	if err := (bootstrapper{cognito: cognito, dynamo: dynamo, userPoolID: *userPoolID, authTable: *authTable, now: time.Now, membershipID: newMembershipID}).bootstrap(ctx, *email); err != nil {
		log.Fatalf("bootstrap administrator: %v", err)
	}
}
