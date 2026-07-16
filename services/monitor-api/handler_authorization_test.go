package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"
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
	identity := &stubPrincipalResolver{err: errors.New("malformed auth_time")}
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
	envelope := decodeEnvelope(t, response.Body)
	if envelope.Reason.Code != string(sharederrors.CodeAuthenticationRequired) || envelope.Reason.Details != nil {
		t.Fatalf("reason = %#v, want redacted AUTHENTICATION_REQUIRED", envelope.Reason)
	}
	if strings.Contains(response.Body, "auth_time") {
		t.Fatalf("principal failure details leaked: %s", response.Body)
	}
	if membership.calls != 0 || len(repo.services) != 0 {
		t.Fatalf("authorization failure dispatched request: membership calls=%d services=%d", membership.calls, len(repo.services))
	}
}

func TestHandleRequestStopsBeforeDispatchWhenMembershipAuthorizationFails(t *testing.T) {
	for _, denial := range []string{"missing", "non-active", "wrong tenant", "unsupported role", "old auth_time"} {
		t.Run(denial, func(t *testing.T) {
			repo := newFakeMonitorRepository()
			identity := &stubPrincipalResolver{identity: auth.AuthenticatedIdentity{Subject: "operator", AuthTime: 2}}
			membership := &stubMembershipResolver{err: sharederrors.New(sharederrors.CodeAuthorizationDenied, map[string]any{"membership": denial})}
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
			envelope := decodeEnvelope(t, response.Body)
			if envelope.Reason.Code != string(sharederrors.CodeAuthorizationDenied) || envelope.Reason.Details != nil {
				t.Fatalf("reason = %#v, want redacted AUTHORIZATION_DENIED", envelope.Reason)
			}
			if strings.Contains(response.Body, denial) {
				t.Fatalf("membership state leaked: %s", response.Body)
			}
			if identity.calls != 1 || membership.calls != 1 || len(repo.services) != 0 {
				t.Fatalf("membership denial dispatched request: principal calls=%d membership calls=%d services=%d", identity.calls, membership.calls, len(repo.services))
			}
		})
	}
}

func TestHandleRequestEmitsSecretSafeAuthorizationDenial(t *testing.T) {
	repo := newFakeMonitorRepository()
	identity := &stubPrincipalResolver{identity: auth.AuthenticatedIdentity{Subject: "operator", AuthTime: 2}}
	membership := &stubMembershipResolver{err: sharederrors.New(sharederrors.CodeAuthorizationDenied, nil)}
	handler := newAuthorizedMonitorHandler(repo, identity, membership)
	var emitted []securityEvent
	handler.securityEvents = func(event securityEvent) { emitted = append(emitted, event) }

	_, err := handler.handleRequest(context.Background(), events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/services",
		RequestContext: events.APIGatewayV2HTTPRequestContext{RequestID: "request-1", HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}},
	})
	if err != nil {
		t.Fatalf("handleRequest() error = %v", err)
	}
	if len(emitted) != 1 {
		t.Fatalf("events = %d, want 1", len(emitted))
	}
	event := emitted[0]
	if event.Event != auth.EventAuthorizationDenied || event.Outcome != "failure" || event.Subject != "operator" || event.CorrelationID != "request-1" {
		t.Fatalf("event = %#v, want authorization-denial context", event)
	}
	if event.Component != "monitor-api" || event.Stage == "" || event.Timestamp == "" {
		t.Fatalf("event = %#v, want complete fixed schema", event)
	}
	if event.Operation != "authorization" || event.Events != 1 || event.EMF == nil {
		t.Fatalf("event = %#v, want bounded authorization metric", event)
	}
	metricSet := event.EMF.CloudWatchMetrics[0]
	if metricSet.Namespace != "BoltMonitor/Auth" || !reflect.DeepEqual(metricSet.Dimensions, [][]string{{"stage", "component", "operation", "outcome"}}) || !reflect.DeepEqual(metricSet.Metrics, []metric{{Name: "AuthenticationEvents", Unit: "Count"}}) {
		t.Fatalf("metric = %#v, want fixed auth dimensions and counter", metricSet)
	}
}

func TestHandleRequestPropagatesDashboardCorrelationWithoutLoggingSensitiveRequestData(t *testing.T) {
	const correlationID = "a1b2c3d4-e5f6-4789-abcd-0123456789ab"
	sensitive := []string{
		"password-value", "recovery-code-value", "totp-secret-value", "transaction-id-value",
		"session-hash-value", "eyJhbGciOiJIUzI1NiJ9.payload.signature", "access-token-value",
		"refresh-token-value", "cookie-value", "encryption-key-value", "request-body-value", "provider-session-value",
	}
	repo := newFakeMonitorRepository()
	identity := &stubPrincipalResolver{identity: auth.AuthenticatedIdentity{Subject: "operator", AuthTime: 2}}
	membership := &stubMembershipResolver{err: sharederrors.New(sharederrors.CodeAuthorizationDenied, nil)}
	handler := newAuthorizedMonitorHandler(repo, identity, membership)
	var output bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previousOutput) })

	_, err := handler.handleRequest(context.Background(), events.APIGatewayV2HTTPRequest{
		RawPath: "/api/v1/services",
		Body:    `{"password":"password-value","code":"recovery-code-value","totpSecret":"totp-secret-value","transactionId":"transaction-id-value","sessionHash":"session-hash-value","encryptionKey":"encryption-key-value","requestBody":"request-body-value","providerPayload":{"Session":"provider-session-value"}}`,
		Headers: map[string]string{
			"x-correlation-id": correlationID,
			"authorization":    "Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature",
			"cookie":           "session=cookie-value; refresh=refresh-token-value; access=access-token-value",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{RequestID: "gateway-request-id", HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	})
	if err != nil {
		t.Fatalf("handleRequest() error = %v", err)
	}
	serialized := output.String()
	if !strings.Contains(serialized, correlationID) {
		t.Fatalf("log = %s, want propagated correlation ID", serialized)
	}
	for _, value := range sensitive {
		if strings.Contains(serialized, value) {
			t.Fatalf("security event leaked sensitive request value %q: %s", value, serialized)
		}
	}
}

func TestMonitorCorrelationIDFallsBackToGatewayRequestIDForUnsafeHeader(t *testing.T) {
	if got := monitorCorrelationID("gateway-request-id", "Bearer access-token-value"); got != "gateway-request-id" {
		t.Fatalf("correlation ID = %q, want gateway request ID", got)
	}
}
