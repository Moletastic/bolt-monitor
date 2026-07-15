package main

import (
	"context"
	"strconv"
	"strings"

	"bolt-monitor/shared/auth"
	sharederrors "bolt-monitor/shared/errors"
	"github.com/aws/aws-lambda-go/events"
)

const (
	cognitoTokenUseClaim = "token_use"
	cognitoClientIDClaim = "client_id"
	cognitoSubjectClaim  = "sub"
	cognitoAuthTimeClaim = "auth_time"
	cognitoIssuedAtClaim = "iat"
)

// cognitoPrincipalResolver adapts API Gateway's validated JWT claims to shared auth types.
// Gateway verifies JWT cryptography; this validates the integration claim shape defensively.
type cognitoPrincipalResolver struct {
	approvedClient map[string]struct{}
}

type PrincipalResolver interface {
	Resolve(context.Context, events.APIGatewayV2HTTPRequest) (auth.AuthenticatedIdentity, error)
}

func newCognitoPrincipalResolver(approvedClientIDs []string) PrincipalResolver {
	approvedClients := make(map[string]struct{}, len(approvedClientIDs))
	for _, clientID := range approvedClientIDs {
		if clientID = strings.TrimSpace(clientID); clientID != "" {
			approvedClients[clientID] = struct{}{}
		}
	}

	return cognitoPrincipalResolver{approvedClient: approvedClients}
}

func (r cognitoPrincipalResolver) Resolve(_ context.Context, request events.APIGatewayV2HTTPRequest) (auth.AuthenticatedIdentity, error) {
	if request.RequestContext.Authorizer == nil || request.RequestContext.Authorizer.JWT == nil {
		return auth.AuthenticatedIdentity{}, authenticationRequired()
	}
	claims := request.RequestContext.Authorizer.JWT.Claims
	if claims == nil || claims[cognitoTokenUseClaim] != "access" {
		return auth.AuthenticatedIdentity{}, authenticationRequired()
	}
	if _, ok := r.approvedClient[claims[cognitoClientIDClaim]]; !ok {
		return auth.AuthenticatedIdentity{}, authenticationRequired()
	}

	subject := strings.TrimSpace(claims[cognitoSubjectClaim])
	if subject == "" {
		return auth.AuthenticatedIdentity{}, authenticationRequired()
	}
	authTime, ok := parseNonNegativeNumericDate(claims[cognitoAuthTimeClaim])
	if !ok {
		return auth.AuthenticatedIdentity{}, authenticationRequired()
	}

	identity := auth.AuthenticatedIdentity{Subject: auth.Subject(subject), AuthTime: authTime}
	if issuedAt, ok := parseNonNegativeNumericDate(claims[cognitoIssuedAtClaim]); ok {
		identity.IssuedAt = &issuedAt
	}
	return identity, nil
}

func parseNonNegativeNumericDate(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || !auth.IsUnixSecond(parsed) {
		return 0, false
	}
	return parsed, true
}

func authenticationRequired() error {
	return sharederrors.New(sharederrors.CodeAuthenticationRequired, nil)
}
