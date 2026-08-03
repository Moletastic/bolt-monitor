package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type apiEnvelope struct {
	Status string `json:"status"`
}

func main() {
	apiURL := flag.String("api-url", "", "monitor API base URL")
	clientID := flag.String("client-id", "", "Cognito readiness app client ID")
	username := flag.String("username", "", "synthetic readiness username")
	passwordParameter := flag.String("password-parameter", "", "SSM SecureString password parameter name")
	flag.Parse()

	if *apiURL == "" || *clientID == "" || *username == "" || *passwordParameter == "" {
		fmt.Fprintln(os.Stderr, "api-url, client-id, username, and password-parameter are required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fail("load AWS configuration", err)
	}
	parameter, err := ssm.NewFromConfig(cfg).GetParameter(ctx, &ssm.GetParameterInput{
		Name:           passwordParameter,
		WithDecryption: aws.Bool(true),
	})
	if err != nil || parameter.Parameter == nil || parameter.Parameter.Value == nil || *parameter.Parameter.Value == "" {
		fail("read readiness password", err)
	}
	auth, err := cognitoidentityprovider.NewFromConfig(cfg).InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: clientID,
		AuthParameters: map[string]string{
			"USERNAME": *username,
			"PASSWORD": *parameter.Parameter.Value,
		},
	})
	if err != nil || auth.AuthenticationResult == nil || auth.AuthenticationResult.AccessToken == nil {
		fail("authenticate readiness user", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(*apiURL, "/")+"/api/v1/services?limit=1", nil)
	if err != nil {
		fail("construct protected readiness request", err)
	}
	request.Header.Set("Authorization", "Bearer "+*auth.AuthenticationResult.AccessToken)
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		fail("call protected readiness endpoint", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		fail("protected readiness endpoint returned non-success status", nil)
	}
	var envelope apiEnvelope
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil || envelope.Status != "success" {
		fail("protected readiness endpoint returned an unsuccessful envelope", err)
	}
	fmt.Println("Protected API readiness validated")
}

func fail(operation string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", operation, err)
	} else {
		fmt.Fprintf(os.Stderr, "%s failed\n", operation)
	}
	os.Exit(1)
}
