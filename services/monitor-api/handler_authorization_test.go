package main

import (
	"context"
	"net/http"
	"testing"

	"bolt-monitor/shared/auth"
	sharederrors "bolt-monitor/shared/errors"
	"github.com/aws/aws-lambda-go/events"
)

type stubPrincipalResolver struct {
	identity auth.AuthenticatedIdentity
	err      error
	calls    int
}

func (r *stubPrincipalResolver) Resolve(_ context.Context, _ events.APIGatewayV2HTTPRequest) (auth.AuthenticatedIdentity, error) {
	r.calls++
	return r.identity, r.err
}

type stubMembershipResolver struct {
	principal auth.Principal
	err       error
	calls     int
}

func (r *stubMembershipResolver) Resolve(_ context.Context, _ auth.AuthenticatedIdentity) (auth.Principal, error) {
	r.calls++
	return r.principal, r.err
}

func TestHandleRequestAuthorizesOnceAndUsesPrincipalTenant(t *testing.T) {
	repo := newFakeMonitorRepository()
	identity := &stubPrincipalResolver{identity: auth.AuthenticatedIdentity{Subject: "operator", AuthTime: 2}}
	membership := &stubMembershipResolver{principal: auth.Principal{Subject: "operator", TenantID: auth.DefaultTenantID, Role: auth.RoleAdmin}}
	handler := newAuthorizedMonitorHandler(repo, identity, membership)
	request := events.APIGatewayV2HTTPRequest{
		RawPath: "/api/v1/services",
		Body:    `{"name":"Auth","tenantId":"OTHER"}`,
		Headers: map[string]string{"X-Tenant-ID": "OTHER"},
		QueryStringParameters: map[string]string{
			"tenantId": "OTHER",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest() error = %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", response.StatusCode, http.StatusCreated, response.Body)
	}
	if identity.calls != 1 || membership.calls != 1 {
		t.Fatalf("resolver calls = principal:%d membership:%d, want 1 each", identity.calls, membership.calls)
	}
	for _, service := range repo.services {
		if service.TenantID != string(auth.DefaultTenantID) {
			t.Fatalf("service tenant = %q, want principal tenant %q", service.TenantID, auth.DefaultTenantID)
		}
	}
}

func TestHandleRequestStopsBeforeDispatchWhenPrincipalResolutionFails(t *testing.T) {
	repo := newFakeMonitorRepository()
	identity := &stubPrincipalResolver{err: sharederrors.New(sharederrors.CodeAuthenticationRequired, nil)}
	membership := &stubMembershipResolver{}
	handler := newAuthorizedMonitorHandler(repo, identity, membership)
	request := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/services",
		Body:           `{"name":"Auth"}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest() error = %v", err)
	}
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
	if membership.calls != 0 || len(repo.services) != 0 {
		t.Fatalf("authorization failure dispatched request: membership calls=%d services=%d", membership.calls, len(repo.services))
	}
}

func TestHandleRequestStopsBeforeDispatchWhenMembershipAuthorizationFails(t *testing.T) {
	repo := newFakeMonitorRepository()
	identity := &stubPrincipalResolver{identity: auth.AuthenticatedIdentity{Subject: "operator", AuthTime: 2}}
	membership := &stubMembershipResolver{err: sharederrors.New(sharederrors.CodeAuthorizationDenied, nil)}
	handler := newAuthorizedMonitorHandler(repo, identity, membership)
	request := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/services",
		Body:           `{"name":"Auth"}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest() error = %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	if identity.calls != 1 || membership.calls != 1 || len(repo.services) != 0 {
		t.Fatalf("membership denial dispatched request: principal calls=%d membership calls=%d services=%d", identity.calls, membership.calls, len(repo.services))
	}
}
