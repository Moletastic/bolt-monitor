package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"bolt-monitor/shared/dynamodbrecord"
	"github.com/aws/aws-lambda-go/events"
)

func incidentCommandRequest(path, key, body string) events.APIGatewayV2HTTPRequest {
	headers := map[string]string{}
	if key != "" {
		headers["idempotency-key"] = key
	}
	return events.APIGatewayV2HTTPRequest{
		Headers:        headers,
		Body:           body,
		RawPath:        path,
		PathParameters: map[string]string{"incidentId": strings.Split(strings.TrimPrefix(path, "/api/v1/incidents/"), "/")[0]},
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}
}

func TestAcknowledgeIncidentRequiresAndReplaysIdempotencyKey(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecord.IncidentRecord{TenantID: defaultTenantID, IncidentID: "INC_1", Status: incidentStatusOpen, OpenedAt: "2026-07-22T09:00:00Z"}
	handler := newMonitorHandler(repo, monitorHandlerTestDependencies{now: func() time.Time { return time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC) }})

	missing, err := handler.handleRequest(context.Background(), incidentCommandRequest("/api/v1/incidents/INC_1/ack", "", ""))
	if err != nil || missing.StatusCode != http.StatusBadRequest || len(repo.commandIdempotency) != 0 {
		t.Fatalf("missing-key response = %d %s, records = %d, err = %v", missing.StatusCode, missing.Body, len(repo.commandIdempotency), err)
	}
	request := incidentCommandRequest("/api/v1/incidents/INC_1/ack", "incident-ack-123", `{"note":"first"}`)
	first, err := handler.handleRequest(context.Background(), request)
	if err != nil || first.StatusCode != http.StatusOK {
		t.Fatalf("first response = %d %s, err = %v", first.StatusCode, first.Body, err)
	}
	second, err := handler.handleRequest(context.Background(), request)
	if err != nil || second.StatusCode != http.StatusOK || second.Body != first.Body {
		t.Fatalf("retry response = %d %s, want canonical %s, err = %v", second.StatusCode, second.Body, first.Body, err)
	}
	if got := repo.incidents["INC_1"].Status; got != incidentStatusAcknowledged || len(repo.commandIdempotency) != 1 {
		t.Fatalf("incident status = %q, records = %d", got, len(repo.commandIdempotency))
	}

	conflict, err := handler.handleRequest(context.Background(), incidentCommandRequest("/api/v1/incidents/INC_1/ack", "incident-ack-123", `{"note":"different"}`))
	if err != nil || conflict.StatusCode != http.StatusConflict || !strings.Contains(conflict.Body, "IDEMPOTENCY_CONFLICT") {
		t.Fatalf("conflict response = %d %s, err = %v", conflict.StatusCode, conflict.Body, err)
	}
}

func TestResolveIncidentReplaysCanonicalIdempotentResponse(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_2"] = dynamodbrecord.IncidentRecord{TenantID: defaultTenantID, IncidentID: "INC_2", Status: incidentStatusAcknowledged, OpenedAt: "2026-07-22T09:00:00Z"}
	handler := newMonitorHandler(repo, monitorHandlerTestDependencies{now: func() time.Time { return time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC) }})
	request := incidentCommandRequest("/api/v1/incidents/INC_2/resolve", "incident-resolve-123", "")

	first, err := handler.handleRequest(context.Background(), request)
	if err != nil || first.StatusCode != http.StatusOK {
		t.Fatalf("first response = %d %s, err = %v", first.StatusCode, first.Body, err)
	}
	second, err := handler.handleRequest(context.Background(), request)
	if err != nil || second.StatusCode != http.StatusOK || second.Body != first.Body {
		t.Fatalf("retry response = %d %s, want canonical %s, err = %v", second.StatusCode, second.Body, first.Body, err)
	}
	if got := repo.incidents["INC_2"].Status; got != incidentStatusResolved || len(repo.commandIdempotency) != 1 {
		t.Fatalf("incident status = %q, records = %d", got, len(repo.commandIdempotency))
	}
}
