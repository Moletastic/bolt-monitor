package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"bolt-monitor/shared/auth"
	"github.com/aws/aws-lambda-go/events"
)

func TestAuthorizationIntegrationEnforcesAuthTimeMembershipAndFixedTenant(t *testing.T) {
	repo := newFakeMonitorRepository()
	client := &fakeAuthTableDynamoClient{}
	handler := newAuthorizedMonitorHandler(
		repo,
		newCognitoPrincipalResolver([]string{"direct-client"}),
		newAuthTableMembershipResolver(client, "auth-table"),
	)

	request := func(authTime int64) events.APIGatewayV2HTTPRequest {
		return events.APIGatewayV2HTTPRequest{
			RawPath: "/api/v1/services",
			Body:    `{"name":"Authorized service","tenantId":"OTHER"}`,
			Headers: map[string]string{"X-Tenant-ID": "OTHER"},
			QueryStringParameters: map[string]string{
				"tenantId": "OTHER",
			},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
					JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{Claims: map[string]string{
						"token_use": "access",
						"client_id": "direct-client",
						"sub":       "operator-subject",
						"auth_time": fmt.Sprintf("%d", authTime),
						"iat":       "2000000000",
					}},
				},
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost},
			},
		}
	}

	denied := validAuthMembershipRecord()
	denied.AuthValidAfter = 100
	for _, authTime := range []int64{99, 100} {
		client.getItemOutput = membershipGetItemOutput(t, denied)
		response, err := handler.handleRequest(context.Background(), request(authTime))
		if err != nil {
			t.Fatalf("handleRequest(auth_time=%d) error = %v", authTime, err)
		}
		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("status for auth_time=%d = %d, want %d", authTime, response.StatusCode, http.StatusForbidden)
		}
	}

	client.getItemOutput = membershipGetItemOutput(t, denied)
	response, err := handler.handleRequest(context.Background(), request(101))
	if err != nil {
		t.Fatalf("handleRequest(later full auth) error = %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status for later full auth = %d, want %d: %s", response.StatusCode, http.StatusCreated, response.Body)
	}
	if len(repo.services) != 1 {
		t.Fatalf("services = %d, want one authorized creation", len(repo.services))
	}
	for _, service := range repo.services {
		if service.TenantID != string(auth.DefaultTenantID) {
			t.Fatalf("service tenant = %q, want fixed %q", service.TenantID, auth.DefaultTenantID)
		}
	}

	inactive := denied
	inactive.Status = "DISABLED"
	client.getItemOutput = membershipGetItemOutput(t, inactive)
	response, err = handler.handleRequest(context.Background(), request(101))
	if err != nil {
		t.Fatalf("handleRequest(unexpired JWT after disable) error = %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status for inactive membership = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	if len(repo.services) != 1 {
		t.Fatalf("inactive membership dispatched request: services = %d, want one", len(repo.services))
	}
	if len(client.getItemInputs) != 4 {
		t.Fatalf("AuthTable reads = %d, want one strong read per request", len(client.getItemInputs))
	}
	for _, input := range client.getItemInputs {
		if input.ConsistentRead == nil || !*input.ConsistentRead {
			t.Fatal("membership authorization must use a strongly consistent AuthTable read")
		}
	}
}
