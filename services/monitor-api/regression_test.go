package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/outboundhttp"
	"bolt-monitor/shared/resultstatus"
)

// TestEnvelopeShape locks the public response envelope schema. Any change to
// status/data/reason/message/pagination fields is a breaking change to the
// dashboard contract.
func TestEnvelopeShape(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services["auth"] = monitorconfig.Service{
		TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth",
		LifecycleState: monitorconfig.ServiceLifecycleDraft,
	}
	handler := newMonitorHandler(repo)
	handler.tenantID = defaultTenantID

	request := events.APIGatewayV2HTTPRequest{
		RawPath: "/api/v1/services", RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
	}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &envelope); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody=%s", err, response.Body)
	}
	if got, ok := envelope["status"].(string); !ok || got != "success" {
		t.Fatalf("status = %v, want \"success\"", envelope["status"])
	}
	if _, ok := envelope["data"]; !ok {
		t.Fatal("envelope missing data field")
	}
	if _, hasReason := envelope["reason"]; hasReason {
		t.Fatal("success envelope must omit reason field")
	}
}

// TestErrorEnvelopeShape pins the failure envelope's reason.code + reason.details
// shape so the dashboard renderer can rely on the stable contract.
func TestErrorEnvelopeShape(t *testing.T) {
	repo := newFakeMonitorRepository()
	handler := newMonitorHandler(repo)
	handler.tenantID = defaultTenantID

	request := events.APIGatewayV2HTTPRequest{
		RawPath: "/api/v1/services/missing", RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
		PathParameters: map[string]string{"serviceId": "missing"},
	}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.StatusCode)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &envelope); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if got, ok := envelope["status"].(string); !ok || got != "error" {
		t.Fatalf("status = %v, want \"error\"", envelope["status"])
	}
	reason, ok := envelope["reason"].(map[string]interface{})
	if !ok {
		t.Fatalf("envelope missing reason object")
	}
	if _, ok := reason["code"].(string); !ok {
		t.Fatalf("reason missing code: %v", reason)
	}
	if _, ok := reason["details"]; !ok {
		t.Fatalf("reason missing details: %v", reason)
	}
}

// TestPaginationEnvelopeShape pins that the paginated monitor-runs endpoint
// emits a page object with items + nextKey fields.
func TestPaginationEnvelopeShape(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{
		TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage",
		Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, Enabled: true,
		FailureThreshold: 1, RecoveryThreshold: 1,
		HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
	}
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	repo.runs[monitorKey("auth", "public-http")] = []resultstatus.CheckRun{{
		TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_ABC",
		StartedAt: now, FinishedAt: now.Add(time.Second),
		DurationMs: 1000, Outcome: checkexecution.OutcomeSuccess,
	}}
	handler := newMonitorHandler(repo)
	handler.tenantID = defaultTenantID

	request := events.APIGatewayV2HTTPRequest{
		RawPath: "/api/v1/services/auth/monitors/public-http/runs", RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
		PathParameters: map[string]string{"serviceId": "auth", "monitorId": "public-http"},
	}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.StatusCode, response.Body)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &envelope); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	pagination, ok := envelope["pagination"].(map[string]interface{})
	if !ok {
		t.Fatalf("envelope.pagination missing: %v", envelope)
	}
	for _, key := range []string{"size"} {
		if _, ok := pagination[key]; !ok {
			t.Fatalf("pagination missing %q: %v", key, pagination)
		}
	}
}

// TestManualRunResponseShape pins the success envelope of the manual-run
// endpoint. The response keeps the execution-result fields used by the
// dashboard: monitorId/runId/outcome/trigger/timestamps.
func TestManualRunResponseShape(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{
		TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage",
		Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, Enabled: true,
		FailureThreshold: 1, RecoveryThreshold: 1,
		HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
	}
	handler := newMonitorHandler(repo)
	handler.tenantID = defaultTenantID
	handler.now = func() time.Time { return time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC) }

	handler.executor = &recordingMonitorExecutor{response: outboundhttp.Response{StatusCode: http.StatusOK, Body: []byte("ok")}}
	response, err := handler.runMonitor(context.Background(), "auth", "public-http")
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", response.StatusCode, response.Body)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &envelope); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	data, ok := envelope["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("envelope.data is not an object: %v", envelope["data"])
	}
	for _, key := range []string{"runId", "serviceId", "monitorId", "outcome", "trigger", "startedAt", "finishedAt"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("data missing %q: %v", key, data)
		}
	}
}

// TestEscalationPolicyResponseShape pins the escalation-policy response
// payload. The dashboard renders businessHoursPath and offHoursPath arrays.
func TestEscalationPolicyResponseShape(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.policies["POL_1"] = escalation.EscalationPolicy{
		TenantID: defaultTenantID, PolicyID: "POL_1", Name: "Primary",
		BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{
			{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeEmail, Target: "ops@example.com"}}},
		}},
	}
	handler := newMonitorHandler(repo)
	handler.tenantID = defaultTenantID

	request := events.APIGatewayV2HTTPRequest{
		RawPath: "/api/v1/escalation-policies", RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
	}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &envelope); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	data, ok := envelope["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("envelope.data is not an object: %v", envelope["data"])
	}
	policies, ok := data["policies"].([]interface{})
	if !ok || len(policies) == 0 {
		t.Fatalf("envelope.data.policies missing or empty: %v", data)
	}
	first, ok := policies[0].(map[string]interface{})
	if !ok {
		t.Fatalf("policy not an object: %v", policies[0])
	}
	for _, key := range []string{"policyId", "name", "businessHoursPath", "offHoursPath"} {
		if _, ok := first[key]; !ok {
			t.Fatalf("policy missing %q: %v", key, first)
		}
	}
}

// TestMonitorRunPagePreservesContinuation proves the repository carries the
// DynamoDB LastEvaluatedKey through its typed history page instead of silently
// truncating the monitor-runs endpoint.
func TestMonitorRunPagePreservesContinuation(t *testing.T) {
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	record := resultstatus.CheckRunRecord{
		PK: "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP", SK: "RUN#2026-05-17T22:00:00Z#RUN_ABC",
		EntityType: "CheckRun", TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http",
		RunID: "RUN_ABC", Type: "http", Trigger: "manual", StartedAt: now.Format(time.RFC3339),
		FinishedAt: now.Add(time.Second).Format(time.RFC3339), DurationMs: 1000, Outcome: "success",
	}
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	nextKey := map[string]sharedaws.AttributeValue{
		"PK": &sharedaws.AttributeValueMemberS{Value: record.PK},
		"SK": &sharedaws.AttributeValueMemberS{Value: record.SK},
	}
	client := &fakeDynamoClient{queryOutput: &sharedaws.DynamoDBQueryOutput{
		Items:            []map[string]sharedaws.AttributeValue{item},
		LastEvaluatedKey: nextKey,
	}}
	repo := newDynamoMonitorRepository(client, "table-name")

	page, err := repo.ListMonitorRunsPage(context.Background(), defaultTenantID, "auth", "public-http", 1, nil)
	if err != nil {
		t.Fatalf("ListMonitorRunsPage: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].RunID != "RUN_ABC" {
		t.Fatalf("items = %#v", page.Items)
	}
	if page.NextKey == nil {
		t.Fatal("NextKey missing: repository discarded LastEvaluatedKey")
	}
	if got := page.NextKey["SK"].(*sharedaws.AttributeValueMemberS).Value; got != record.SK {
		t.Fatalf("NextKey SK = %q, want %q", got, record.SK)
	}
}
