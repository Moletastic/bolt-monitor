package checkexecution

import (
	"context"
	"strings"
	"testing"

	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/outboundhttp"
)

type fakeHTTPExecutor struct {
	response outboundhttp.Response
	err      error
}

func (e fakeHTTPExecutor) Execute(_ context.Context, _ outboundhttp.Request) (outboundhttp.Response, error) {
	return e.response, e.err
}

func TestBuildExecutionRequestsSkipsDisabledMonitors(t *testing.T) {
	monitors := []monitorconfig.Monitor{
		{
			ServiceID:         "auth",
			MonitorID:         "public-http",
			TenantID:          "DEFAULT",
			Name:              "Enabled",
			Type:              monitorconfig.MonitorTypeHTTP,
			IntervalSeconds:   60,
			Enabled:           true,
			FailureThreshold:  1,
			RecoveryThreshold: 1,
			HTTP:              &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
		},
		{
			ServiceID:         "auth",
			MonitorID:         "deep-ready",
			TenantID:          "DEFAULT",
			Name:              "Disabled",
			Type:              monitorconfig.MonitorTypeHTTP,
			IntervalSeconds:   60,
			Enabled:           false,
			FailureThreshold:  1,
			RecoveryThreshold: 1,
			HTTP:              &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
		},
	}

	requests, err := BuildExecutionRequests(monitors, TriggerTypeRecurring)
	if err != nil {
		t.Fatalf("BuildExecutionRequests returned error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	if requests[0].Monitor.MonitorID != "public-http" {
		t.Fatalf("monitorID = %q, want public-http", requests[0].Monitor.MonitorID)
	}
}

func TestSchedulerConfigRequiresStopControl(t *testing.T) {
	config := SchedulerConfig{RecurringEnabled: true}
	err := config.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error, want missing stop control failure")
	}
	typed, ok := sharederrors.As(err)
	if !ok {
		t.Fatalf("Validate error = %T, want typed error", err)
	}
	if typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("Code = %s, want %s", typed.Code, sharederrors.CodeValidationFailed)
	}
	if typed.Details["field"] != "stopControlMode" {
		t.Fatalf("field = %v, want stopControlMode", typed.Details["field"])
	}
}

func TestExecuteHTTPReturnsNormalizedResult(t *testing.T) {
	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:              "https://status.example.com",
				Method:              "GET",
				TimeoutMs:           5000,
				ExpectedStatusCodes: []int{200},
			},
		},
		Trigger: TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), fakeHTTPExecutor{response: outboundhttp.Response{StatusCode: 200, Body: []byte("ok")}}, request)
	if result.Outcome != OutcomeSuccess {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeSuccess)
	}
	if result.StatusCode == nil || *result.StatusCode != 200 {
		t.Fatal("status code missing or incorrect")
	}
}

func TestExecuteHTTPSucceedsWhenExpectedBodyMatches(t *testing.T) {
	expected := "healthy"
	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:               "https://status.example.com",
				Method:               "GET",
				TimeoutMs:            5000,
				ExpectedStatusCodes:  []int{200},
				ExpectedBodyContains: &expected,
			},
		},
		Trigger: TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), fakeHTTPExecutor{response: outboundhttp.Response{StatusCode: 200, Body: []byte("service healthy")}}, request)
	if result.Outcome != OutcomeSuccess {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeSuccess)
	}
	if result.Error != "" {
		t.Fatalf("Error = %q, want empty", result.Error)
	}
}

func TestExecuteHTTPFailsWhenExpectedBodyMissing(t *testing.T) {
	expected := "healthy"
	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:               "https://status.example.com",
				Method:               "GET",
				TimeoutMs:            5000,
				ExpectedStatusCodes:  []int{200},
				ExpectedBodyContains: &expected,
			},
		},
		Trigger: TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), fakeHTTPExecutor{response: outboundhttp.Response{StatusCode: 200, Body: []byte("service degraded")}}, request)
	if result.Outcome != OutcomeFailure {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeFailure)
	}
	if result.Error == "" {
		t.Fatal("Error empty, want body assertion failure")
	}
}

func TestExecuteHTTPFailsStatusBeforeBodyMatch(t *testing.T) {
	expected := "healthy"
	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:               "https://status.example.com",
				Method:               "GET",
				TimeoutMs:            5000,
				ExpectedStatusCodes:  []int{200},
				ExpectedBodyContains: &expected,
			},
		},
		Trigger: TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), fakeHTTPExecutor{response: outboundhttp.Response{StatusCode: 502, Body: []byte("service healthy")}}, request)
	if result.Outcome != OutcomeFailure {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeFailure)
	}
	if result.Error != "unexpected status code 502" {
		t.Fatalf("Error = %q, want unexpected status code message", result.Error)
	}
}

func TestExecuteHTTPMapsOutboundFailureWithoutSecrets(t *testing.T) {
	request := ExecutionRequest{Monitor: monitorconfig.Monitor{
		ServiceID: "auth", MonitorID: "public-http", TenantID: "DEFAULT", Name: "Homepage",
		Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, Enabled: true,
		HTTP: &monitorconfig.HTTPConfiguration{Target: "https://token:secret@status.example.com", Method: "GET", TimeoutMs: 5000},
	}}
	result := ExecuteHTTP(context.Background(), fakeHTTPExecutor{err: &outboundhttp.Error{Kind: outboundhttp.KindAddressBlocked}}, request)
	if result.Outcome != OutcomeError || result.FailureCode != string(outboundhttp.KindAddressBlocked) {
		t.Fatalf("result = %#v", result)
	}
	if strings.Contains(result.Error, "secret") || result.Error != outboundhttp.SafeMessage(&outboundhttp.Error{Kind: outboundhttp.KindAddressBlocked}) {
		t.Fatalf("unsafe error = %q", result.Error)
	}
}
