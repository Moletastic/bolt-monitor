package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandlerReturnsSuccessEnvelope(t *testing.T) {
	response, err := handler(events.APIGatewayV2HTTPRequest{})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	if response.Headers["Content-Type"] != "application/json" {
		t.Fatalf("content-type = %q, want %q", response.Headers["Content-Type"], "application/json")
	}

	var payload struct {
		Status  string            `json:"status"`
		Data    map[string]string `json:"data"`
		Reason  json.RawMessage   `json:"reason"`
		Message json.RawMessage   `json:"message"`
	}
	if err := json.Unmarshal([]byte(response.Body), &payload); err != nil {
		t.Fatalf("response body is not valid json: %v", err)
	}

	if payload.Status != "success" {
		t.Fatalf("payload.status = %q, want success", payload.Status)
	}
	if payload.Data["status"] != "ok" {
		t.Fatalf("payload.data.status = %q, want ok", payload.Data["status"])
	}
	if payload.Reason != nil || payload.Message != nil {
		t.Fatalf("error-only fields must be omitted: reason=%s message=%s", payload.Reason, payload.Message)
	}
}

func TestHandlerErrEnvelopeShape(t *testing.T) {
	env := errEnvelopeForTest()
	body, err := env.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var payload struct {
		Status string `json:"status"`
		Reason struct {
			Code string `json:"code"`
		} `json:"reason"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if payload.Status != "error" {
		t.Fatalf("payload.status = %q, want error", payload.Status)
	}
	if payload.Reason.Code != "HEALTH_UNAVAILABLE" {
		t.Fatalf("payload.reason.code = %q, want HEALTH_UNAVAILABLE", payload.Reason.Code)
	}
}
