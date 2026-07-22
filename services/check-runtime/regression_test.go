package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/monitorconfig"
)

// TestSQSEscalationPayloadShape pins the JSON shape that the check-runtime
// worker emits onto the escalation SQS queue on incident transitions. Any
// change here is a breaking change to the escalation runtime contract.
func TestSQSEscalationPayloadShape(t *testing.T) {
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	monitor := monitorconfig.Monitor{
		TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http",
		Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60,
		HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
	}
	service := monitorconfig.Service{TenantID: "DEFAULT", ServiceID: "auth", Name: "Auth API"}

	cases := []struct {
		name       string
		transition string
		result     checkexecution.ExecutionResult
		wantFields map[string]string
	}{
		{
			name:       "incident opened",
			transition: "incident.down",
			result:     checkexecution.ExecutionResult{Outcome: checkexecution.OutcomeFailure, Error: "timeout", FinishedAt: now},
			wantFields: map[string]string{
				"eventType":   "incident.down",
				"tenantId":    "DEFAULT",
				"serviceId":   "auth",
				"monitorId":   "public-http",
				"monitorName": "Homepage",
				"serviceName": "Auth API",
				"timestamp":   "2026-05-17T22:00:00Z",
				"message":     "🚨 Incident Opened: Homepage is DOWN",
			},
		},
		{
			name:       "incident resolved",
			transition: "incident.up",
			result:     checkexecution.ExecutionResult{Outcome: checkexecution.OutcomeSuccess, FinishedAt: now},
			wantFields: map[string]string{
				"eventType":   "incident.up",
				"tenantId":    "DEFAULT",
				"serviceId":   "auth",
				"monitorId":   "public-http",
				"monitorName": "Homepage",
				"serviceName": "Auth API",
				"timestamp":   "2026-05-17T22:00:00Z",
				"message":     "✅ Incident Resolved: Homepage is UP",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := buildEscalationMessage(tc.transition, monitor, service, "INC_42", tc.result)
			var decoded map[string]interface{}
			if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
				t.Fatalf("payload is not valid JSON: %v\npayload=%s", err, payload)
			}
			for key, want := range tc.wantFields {
				got, ok := decoded[key]
				if !ok {
					t.Fatalf("payload missing key %q\npayload=%s", key, payload)
				}
				if str, ok := got.(string); !ok || !strings.HasPrefix(str, want) {
					t.Fatalf("payload[%q] = %v, want prefix %q", key, got, want)
				}
			}
			if _, ok := decoded["incidentId"]; !ok {
				t.Fatalf("payload missing incidentId\npayload=%s", payload)
			}
		})
	}
}

// TestSQSWorkerRequestShape pins the worker payload shape: the body that
// the scheduler Lambda places onto the worker SQS queue must keep the
// "monitor" and "trigger" fields so the worker's ExecuteHTTP can drive the
// HTTP request.
func TestSQSWorkerRequestShape(t *testing.T) {
	req := checkexecution.ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http",
			Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60,
			HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
		},
		RunID:   "RUN_ABC",
		Trigger: checkexecution.TriggerTypeRecurring,
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded checkexecution.ExecutionRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Trigger != checkexecution.TriggerTypeRecurring {
		t.Fatalf("trigger = %q", decoded.Trigger)
	}
	if decoded.RunID != "RUN_ABC" {
		t.Fatalf("runID = %q", decoded.RunID)
	}
	if decoded.Monitor.MonitorID != "public-http" {
		t.Fatalf("monitor.monitorId = %q", decoded.Monitor.MonitorID)
	}
}

// TestSQSEventDecodingRejectsGarbageBody ensures the worker fails closed when
// the scheduler publishes an unparseable body.
func TestSQSEventDecodingRejectsGarbageBody(t *testing.T) {
	handler := newTestRuntimeHandler(newFakeRuntimeRepository(), &fakeSQSClient{}, "", "", defaultTenantID, modeWorker)
	_, err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: "not-json"}}})
	if err == nil {
		t.Fatal("expected error for unparseable SQS body")
	}
}
