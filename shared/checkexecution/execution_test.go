package checkexecution

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/probelocationcatalog"
)

func catalog() probelocationcatalog.Catalog {
	return probelocationcatalog.Catalog{Locations: []probelocationcatalog.Location{{
		LocationID:      "iad",
		DisplayName:     "US East",
		ExecutionTarget: "worker-us-east",
		Enabled:         true,
	}}}
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
			ProbeLocations:    []string{"iad"},
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
			ProbeLocations:    []string{"iad"},
			Enabled:           false,
			FailureThreshold:  1,
			RecoveryThreshold: 1,
			HTTP:              &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000},
		},
	}

	requests, err := BuildExecutionRequests(monitors, catalog(), TriggerTypeRecurring)
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
	if err := config.Validate(); err == nil {
		t.Fatal("Validate returned nil error, want missing stop control failure")
	}
}

func TestExecuteHTTPReturnsNormalizedResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			ProbeLocations:  []string{"iad"},
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:              server.URL,
				Method:              "GET",
				TimeoutMs:           5000,
				ExpectedStatusCodes: []int{200},
			},
		},
		ProbeLocation: catalog().Locations[0],
		Trigger:       TriggerTypeManual,
	}

	client := &http.Client{Timeout: 5 * time.Second}
	result := ExecuteHTTP(context.Background(), client, request)
	if result.Outcome != OutcomeSuccess {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeSuccess)
	}
	if result.StatusCode == nil || *result.StatusCode != 200 {
		t.Fatal("status code missing or incorrect")
	}
	if result.ProbeLocationID != "iad" {
		t.Fatalf("ProbeLocationID = %q, want iad", result.ProbeLocationID)
	}
}

func TestExecuteHTTPSucceedsWhenExpectedBodyMatches(t *testing.T) {
	expected := "healthy"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("service healthy"))
	}))
	defer server.Close()

	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			ProbeLocations:  []string{"iad"},
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:               server.URL,
				Method:               "GET",
				TimeoutMs:            5000,
				ExpectedStatusCodes:  []int{200},
				ExpectedBodyContains: &expected,
			},
		},
		ProbeLocation: catalog().Locations[0],
		Trigger:       TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), &http.Client{Timeout: 5 * time.Second}, request)
	if result.Outcome != OutcomeSuccess {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeSuccess)
	}
	if result.Error != "" {
		t.Fatalf("Error = %q, want empty", result.Error)
	}
}

func TestExecuteHTTPFailsWhenExpectedBodyMissing(t *testing.T) {
	expected := "healthy"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("service degraded"))
	}))
	defer server.Close()

	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			ProbeLocations:  []string{"iad"},
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:               server.URL,
				Method:               "GET",
				TimeoutMs:            5000,
				ExpectedStatusCodes:  []int{200},
				ExpectedBodyContains: &expected,
			},
		},
		ProbeLocation: catalog().Locations[0],
		Trigger:       TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), &http.Client{Timeout: 5 * time.Second}, request)
	if result.Outcome != OutcomeFailure {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeFailure)
	}
	if result.Error == "" {
		t.Fatal("Error empty, want body assertion failure")
	}
}

func TestExecuteHTTPFailsStatusBeforeBodyMatch(t *testing.T) {
	expected := "healthy"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("service healthy"))
	}))
	defer server.Close()

	request := ExecutionRequest{
		Monitor: monitorconfig.Monitor{
			ServiceID:       "auth",
			MonitorID:       "public-http",
			TenantID:        "DEFAULT",
			Name:            "Homepage",
			Type:            monitorconfig.MonitorTypeHTTP,
			IntervalSeconds: 60,
			ProbeLocations:  []string{"iad"},
			Enabled:         true,
			HTTP: &monitorconfig.HTTPConfiguration{
				Target:               server.URL,
				Method:               "GET",
				TimeoutMs:            5000,
				ExpectedStatusCodes:  []int{200},
				ExpectedBodyContains: &expected,
			},
		},
		ProbeLocation: catalog().Locations[0],
		Trigger:       TriggerTypeManual,
	}

	result := ExecuteHTTP(context.Background(), &http.Client{Timeout: 5 * time.Second}, request)
	if result.Outcome != OutcomeFailure {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeFailure)
	}
	if result.Error != "unexpected status code 502" {
		t.Fatalf("Error = %q, want unexpected status code message", result.Error)
	}
}
