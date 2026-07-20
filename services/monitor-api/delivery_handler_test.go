package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/notifications"
	"github.com/aws/aws-lambda-go/events"
)

func TestListIncidentDeliveriesReturnsAllSixStates(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	repo.deliveries["dlv_pending"] = makeDelivery("INC_1", "dlv_pending", notifications.DeliveryPending)
	repo.deliveries["dlv_inflight"] = makeDelivery("INC_1", "dlv_inflight", notifications.DeliveryInFlight)
	repo.deliveries["dlv_retryable"] = makeDelivery("INC_1", "dlv_retryable", notifications.DeliveryRetryable)
	repo.deliveries["dlv_ambiguous"] = makeDelivery("INC_1", "dlv_ambiguous", notifications.DeliveryAmbiguous)
	repo.deliveries["dlv_delivered"] = makeDelivery("INC_1", "dlv_delivered", notifications.DeliveryDelivered)
	repo.deliveries["dlv_terminal"] = makeDelivery("INC_1", "dlv_terminal", notifications.DeliveryTerminalFailed)
	repo.deliveries["dlv_other"] = makeDelivery("INC_2", "dlv_other", notifications.DeliveryDelivered)
	handler := newAuthorizedHandler(repo)
	response, err := handler.listIncidentDeliveries(context.Background(), "INC_1", events.APIGatewayV2HTTPRequest{RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if !strings.Contains(response.Body, "dlv_pending") || !strings.Contains(response.Body, "dlv_terminal") || !strings.Contains(response.Body, "dlv_delivered") {
		t.Fatalf("body missing expected deliveries: %s", response.Body)
	}
	if strings.Contains(response.Body, "dlv_other") {
		t.Fatalf("body must not include deliveries from another incident: %s", response.Body)
	}
}

func TestListIncidentDeliveriesReturnsNotFoundForUnknownIncident(t *testing.T) {
	repo := newFakeMonitorRepository()
	handler := newAuthorizedHandler(repo)
	response, _ := handler.listIncidentDeliveries(context.Background(), "INC_MISSING", events.APIGatewayV2HTTPRequest{RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}})
	if response.StatusCode != http.StatusNotFound || !strings.Contains(response.Body, "INCIDENT_NOT_FOUND") {
		t.Fatalf("expected 404 INCIDENT_NOT_FOUND, got status=%d body=%s", response.StatusCode, response.Body)
	}
}

func TestReplayIncidentDeliveryResetsTerminalFailedToPending(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	repo.deliveries["dlv_terminal"] = makeDelivery("INC_1", "dlv_terminal", notifications.DeliveryTerminalFailed)
	handler := newAuthorizedHandler(repo)
	request := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	response, err := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_terminal", request)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.StatusCode, response.Body)
	}
	if repo.deliveries["dlv_terminal"].State != notifications.DeliveryPending {
		t.Fatalf("state did not reset: %s", repo.deliveries["dlv_terminal"].State)
	}
	if _, ok := repo.replayIdempotency["key-1"]; !ok {
		t.Fatalf("idempotency record missing")
	}
}

func TestReplayIncidentDeliverySameKeySameRequestIsIdempotent(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	repo.deliveries["dlv_terminal"] = makeDelivery("INC_1", "dlv_terminal", notifications.DeliveryTerminalFailed)
	handler := newAuthorizedHandler(repo)
	request := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	if _, err := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_terminal", request); err != nil {
		t.Fatalf("first replay failed: %v", err)
	}
	count := len(repo.replayIdempotency)
	if _, err := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_terminal", request); err != nil {
		t.Fatalf("second replay failed: %v", err)
	}
	if len(repo.replayIdempotency) != count {
		t.Fatalf("idempotency record duplicated: %+v", repo.replayIdempotency)
	}
}

func TestReplayIncidentDeliverySameKeyDifferentRequestConflicts(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	repo.deliveries["dlv_terminal"] = makeDelivery("INC_1", "dlv_terminal", notifications.DeliveryTerminalFailed)
	handler := newAuthorizedHandler(repo)
	first := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, Body: `{"reason":"first"}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	second := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, Body: `{"reason":"second"}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	if _, err := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_terminal", first); err != nil {
		t.Fatalf("first replay failed: %v", err)
	}
	response, _ := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_terminal", second)
	if response.StatusCode != http.StatusConflict || !strings.Contains(response.Body, "IDEMPOTENCY_CONFLICT") {
		t.Fatalf("expected idempotency conflict, got status=%d body=%s", response.StatusCode, response.Body)
	}
}

func TestReplayIncidentDeliveryRejectsNonTerminalState(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	repo.deliveries["dlv_delivered"] = makeDelivery("INC_1", "dlv_delivered", notifications.DeliveryDelivered)
	handler := newAuthorizedHandler(repo)
	request := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	response, _ := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_delivered", request)
	if response.StatusCode != http.StatusConflict || !strings.Contains(response.Body, "DELIVERY_NOT_REPLAYABLE") {
		t.Fatalf("expected not replayable, got status=%d body=%s", response.StatusCode, response.Body)
	}
}

func TestReplayIncidentDeliveryRequiresIdempotencyKey(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	repo.deliveries["dlv_terminal"] = makeDelivery("INC_1", "dlv_terminal", notifications.DeliveryTerminalFailed)
	handler := newAuthorizedHandler(repo)
	request := events.APIGatewayV2HTTPRequest{RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	response, _ := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_terminal", request)
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(response.Body, "VALIDATION_FAILED") {
		t.Fatalf("expected validation failure, got status=%d body=%s", response.StatusCode, response.Body)
	}
}

func TestReplayIncidentDeliveryReturnsNotFoundForMissingDelivery(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_1"] = dynamodbrecordIncident(t)
	handler := newAuthorizedHandler(repo)
	request := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	response, _ := handler.replayIncidentDelivery(context.Background(), "INC_1", "dlv_missing", request)
	if response.StatusCode != http.StatusNotFound || !strings.Contains(response.Body, "DELIVERY_NOT_FOUND") {
		t.Fatalf("expected delivery not found, got status=%d body=%s", response.StatusCode, response.Body)
	}
}

func TestReplayRouteCrossTenantAccess(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.incidents["INC_OTHER"] = dynamodbrecordIncidentForTenant("INC_OTHER", "OTHER")
	repo.deliveries["dlv_terminal"] = makeDeliveryForTenant("INC_OTHER", "dlv_terminal", notifications.DeliveryTerminalFailed, "OTHER")
	handler := newAuthorizedHandler(repo)
	request := events.APIGatewayV2HTTPRequest{Headers: map[string]string{"Idempotency-Key": "key-1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	response, _ := handler.replayIncidentDelivery(context.Background(), "INC_OTHER", "dlv_terminal", request)
	if response.StatusCode != http.StatusNotFound || !strings.Contains(response.Body, "INCIDENT_NOT_FOUND") {
		t.Fatalf("expected cross-tenant incident not found, got status=%d body=%s", response.StatusCode, response.Body)
	}
}

func dynamodbrecordIncident(t *testing.T) dynamodbrecord.IncidentRecord {
	t.Helper()
	return dynamodbrecord.IncidentRecord{TenantID: defaultTenantID, ServiceID: "svc", MonitorID: "mon", IncidentID: "INC_1", Status: "open", OpenedAt: "2026-07-19T00:00:00Z", UpdatedAt: "2026-07-19T00:00:00Z"}
}

func dynamodbrecordIncidentForTenant(id, tenant string) dynamodbrecord.IncidentRecord {
	return dynamodbrecord.IncidentRecord{TenantID: tenant, ServiceID: "svc", MonitorID: "mon", IncidentID: id, Status: "open", OpenedAt: "2026-07-19T00:00:00Z", UpdatedAt: "2026-07-19T00:00:00Z"}
}

func makeDelivery(incidentID, deliveryID string, state notifications.DeliveryState) notifications.DeliveryRecord {
	return makeDeliveryForTenant(incidentID, deliveryID, state, defaultTenantID)
}

func makeDeliveryForTenant(incidentID, deliveryID string, state notifications.DeliveryState, tenant string) notifications.DeliveryRecord {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	return notifications.DeliveryRecord{
		TenantID:     tenant,
		IncidentID:   incidentID,
		TransitionID: "TRN_" + deliveryID,
		DeliveryID:   deliveryID,
		ChannelID:    "CH_1",
		ChannelType:  "telegram",
		StepNumber:   1,
		State:        state,
		AttemptCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// newAuthorizedHandler constructs a monitor handler bypassing authorization for
// focused delivery tests; the production wiring remains unchanged.
func newAuthorizedHandler(repo monitorRepository) monitorHandler {
	handler := newMonitorHandler(repo)
	handler.tenantID = defaultTenantID
	handler.now = func() time.Time { return time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC) }
	return handler
}

var _ = json.Marshal
var _ = escalation.EscalationStatusActive
