package main

import (
	"context"
	"testing"

	"bolt-monitor/shared/auth"
	sharederrors "bolt-monitor/shared/errors"
	"github.com/aws/aws-lambda-go/events"
)

func TestCognitoPrincipalResolverAcceptsValidatedAccessClaims(t *testing.T) {
	resolver := newCognitoPrincipalResolver([]string{"dashboard-client", "direct-client"})
	identity, err := resolver.Resolve(context.Background(), requestWithClaims(map[string]string{
		"token_use": "access",
		"client_id": "dashboard-client",
		"sub":       "operator-subject",
		"auth_time": "1700000000",
		"iat":       "1700000010",
	}))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if identity.Subject != auth.Subject("operator-subject") || identity.AuthTime != 1700000000 {
		t.Fatalf("Resolve() identity = %#v", identity)
	}
	if identity.IssuedAt == nil || *identity.IssuedAt != 1700000010 {
		t.Fatalf("Resolve() IssuedAt = %v, want diagnostic iat", identity.IssuedAt)
	}
}

func TestCognitoPrincipalResolverFailsClosedForInvalidRequiredClaims(t *testing.T) {
	tests := map[string]map[string]string{
		"missing claims":       nil,
		"ID token":             {"token_use": "id", "client_id": "dashboard-client", "sub": "subject", "auth_time": "1"},
		"missing client":       {"token_use": "access", "sub": "subject", "auth_time": "1"},
		"unapproved client":    {"token_use": "access", "client_id": "other", "sub": "subject", "auth_time": "1"},
		"empty subject":        {"token_use": "access", "client_id": "dashboard-client", "sub": " ", "auth_time": "1"},
		"missing auth time":    {"token_use": "access", "client_id": "dashboard-client", "sub": "subject"},
		"fractional auth time": {"token_use": "access", "client_id": "dashboard-client", "sub": "subject", "auth_time": "1.5"},
		"negative auth time":   {"token_use": "access", "client_id": "dashboard-client", "sub": "subject", "auth_time": "-1"},
		"string auth time":     {"token_use": "access", "client_id": "dashboard-client", "sub": "subject", "auth_time": "one"},
		"overflowed auth time": {"token_use": "access", "client_id": "dashboard-client", "sub": "subject", "auth_time": "9223372036854775808"},
	}

	for name, claims := range tests {
		t.Run(name, func(t *testing.T) {
			resolver := newCognitoPrincipalResolver([]string{"dashboard-client"})
			_, err := resolver.Resolve(context.Background(), requestWithClaims(claims))
			assertAuthenticationRequired(t, err)
		})
	}
}

func TestCognitoPrincipalResolverTreatsIssuedAtAsDiagnosticOnly(t *testing.T) {
	validClaims := map[string]string{
		"token_use": "access",
		"client_id": "dashboard-client",
		"sub":       "subject",
		"auth_time": "10",
		"iat":       "not-a-numeric-date",
	}
	resolver := newCognitoPrincipalResolver([]string{"dashboard-client"})
	identity, err := resolver.Resolve(context.Background(), requestWithClaims(validClaims))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if identity.IssuedAt != nil {
		t.Fatalf("Resolve() IssuedAt = %v, want nil for malformed diagnostic iat", identity.IssuedAt)
	}

	delete(validClaims, "auth_time")
	validClaims["iat"] = "11"
	_, err = resolver.Resolve(context.Background(), requestWithClaims(validClaims))
	assertAuthenticationRequired(t, err)
}

func requestWithClaims(claims map[string]string) events.APIGatewayV2HTTPRequest {
	request := events.APIGatewayV2HTTPRequest{}
	if claims != nil {
		request.RequestContext.Authorizer = &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
			JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{Claims: claims},
		}
	}
	return request
}

func assertAuthenticationRequired(t *testing.T, err error) {
	t.Helper()
	typed, ok := sharederrors.As(err)
	if !ok || typed.Code != sharederrors.CodeAuthenticationRequired {
		t.Fatalf("Resolve() error = %v, want AUTHENTICATION_REQUIRED", err)
	}
	if typed.Details != nil {
		t.Fatalf("Resolve() error details = %#v, want no claim details", typed.Details)
	}
}
