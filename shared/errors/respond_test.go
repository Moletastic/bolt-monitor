package errors

import (
	stdlibErrors "errors"
	"net/http"
	"testing"

	"bolt-monitor/shared/api/response"
)

func TestRespondTypedPassesThroughDetails(t *testing.T) {
	te := New(CodeValidationFailed, map[string]any{"field": "name", "reason": "required"})
	status, env := Respond(te)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
	if env.Reason == nil || env.Reason.Code != "VALIDATION_FAILED" {
		t.Fatalf("env.Reason = %+v", env.Reason)
	}
	if env.Reason.Details["field"] != "name" {
		t.Fatalf("Details[field] = %v", env.Reason.Details["field"])
	}
}

func TestRespondNonTypedYieldsInternalWithNilDetails(t *testing.T) {
	status, env := Respond(stdlibErrors.New("boom"))
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if env.Reason == nil || env.Reason.Code != "INTERNAL" {
		t.Fatalf("env.Reason = %+v", env.Reason)
	}
	if env.Reason.Details != nil {
		t.Fatalf("Details = %v, want nil", env.Reason.Details)
	}
}

func TestRespondInternalStrippedEvenWithCause(t *testing.T) {
	status, env := Respond(stdlibErrors.New("something exploded"))
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if env.Reason.Code != "INTERNAL" {
		t.Fatalf("Code = %s", env.Reason.Code)
	}
	if env.Reason.Details != nil {
		t.Fatalf("INTERNAL details leaked: %v", env.Reason.Details)
	}
}

func TestRespondRegisteredStatusMapping(t *testing.T) {
	cases := map[Code]int{
		CodeNotFound:              http.StatusNotFound,
		CodeInvalidJSON:           http.StatusBadRequest,
		CodeValidationFailed:      http.StatusBadRequest,
		CodeImmutableField:        http.StatusBadRequest,
		CodeInlineChannelConfig:   http.StatusBadRequest,
		CodeServiceNotFound:       http.StatusNotFound,
		CodeServiceAlreadyExists:  http.StatusConflict,
		CodeServiceActive:         http.StatusConflict,
		CodeServiceNotArchived:    http.StatusConflict,
		CodeServiceHasNoPolicy:    http.StatusNotFound,
		CodeMonitorNotFound:       http.StatusNotFound,
		CodeMonitorAlreadyExists:  http.StatusConflict,
		CodeMonitorDisabled:       http.StatusConflict,
		CodeMonitorStatusNotFound: http.StatusNotFound,
		CodeLastMonitor:           http.StatusConflict,
		CodeIncidentNotFound:      http.StatusNotFound,
		CodeIncidentNotActionable: http.StatusConflict,
		CodePolicyNotFound:        http.StatusNotFound,
		CodePolicyReferenced:      http.StatusConflict,
		CodeChannelNotFound:       http.StatusNotFound,
	}
	for code, want := range cases {
		status, _ := Respond(New(code, nil))
		if status != want {
			t.Fatalf("Respond(%s) status = %d, want %d", code, status, want)
		}
	}
}

func TestRespondEnvelopeStatus(t *testing.T) {
	te := New(CodeServiceNotFound, nil)
	_, env := Respond(te)
	if env.Status != response.StatusError {
		t.Fatalf("Status = %s, want error", env.Status)
	}
	if env.Data != nil {
		t.Fatalf("Data should be nil on error envelope")
	}
}
