package monitorconfig

import (
	"strings"
	"testing"

	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
)

func TestCreateServiceRequestToService(t *testing.T) {
	service, err := (CreateServiceRequest{
		Name:            "Auth Service",
		ServiceCategory: ServiceCategoryFunction,
	}).ToService("default")
	if err != nil {
		t.Fatalf("ToService returned error: %v", err)
	}
	if service.Name != "Auth Service" {
		t.Fatalf("Name = %q, want Auth Service", service.Name)
	}
	if service.TenantID != "DEFAULT" {
		t.Fatalf("TenantID = %q, want DEFAULT", service.TenantID)
	}
	if service.ServiceCategory != ServiceCategoryFunction {
		t.Fatalf("ServiceCategory = %q, want function", service.ServiceCategory)
	}
}

func TestServiceBusinessHoursClone(t *testing.T) {
	service := Service{
		TenantID:           "DEFAULT",
		ServiceID:          "auth",
		Name:               "Auth",
		LifecycleState:     ServiceLifecycleDraft,
		EscalationPolicyID: "policy-1",
		BusinessHours:      &escalation.BusinessHoursConfig{Timezone: "America/New_York", StartHour: 9, EndHour: 17, DaysOfWeek: []int{5, 1, 3}},
	}

	clone := cloneBusinessHoursConfig(service.BusinessHours)
	if clone == nil {
		t.Fatal("cloneBusinessHoursConfig returned nil")
		return
	}
	if clone.Timezone != "America/New_York" {
		t.Fatalf("Timezone = %q, want America/New_York", clone.Timezone)
	}
	if clone.DaysOfWeek[0] != 1 || clone.DaysOfWeek[2] != 5 {
		t.Fatalf("DaysOfWeek = %v, want sorted [1 3 5]", clone.DaysOfWeek)
	}
}

func TestCreateMonitorRequestToMonitor(t *testing.T) {
	bodyContains := "ok"
	request := CreateMonitorRequest{
		Name:              "Homepage",
		Type:              MonitorTypeHTTP,
		IntervalSeconds:   60,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		HTTP: &HTTPConfiguration{
			Target:               "https://example.com",
			Method:               "get",
			TimeoutMs:            5000,
			ExpectedStatusCodes:  []int{200},
			ExpectedBodyContains: &bodyContains,
		},
	}

	monitor, err := request.ToMonitor("auth", "default", "public-http")
	if err != nil {
		t.Fatalf("ToMonitor returned error: %v", err)
	}

	if monitor.MonitorID != "public-http" {
		t.Fatalf("MonitorID = %q, want public-http", monitor.MonitorID)
	}
	if monitor.ServiceID != "auth" {
		t.Fatalf("ServiceID = %q, want auth", monitor.ServiceID)
	}
	if monitor.TenantID != "DEFAULT" {
		t.Fatalf("TenantID = %q, want DEFAULT", monitor.TenantID)
	}
	if monitor.HTTP == nil {
		t.Fatal("HTTP configuration is nil")
	}
	if monitor.HTTP.Method != "GET" {
		t.Fatalf("HTTP.Method = %q, want GET", monitor.HTTP.Method)
	}
}

func TestHTTPConfigurationValidateOutboundPolicy(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		timeout int
		field   string
	}{
		{name: "safe public", target: "https://status.example.com", timeout: 5000},
		{name: "unsupported scheme", target: "ftp://status.example.com", timeout: 5000, field: "http.target"},
		{name: "blocked literal", target: "http://127.0.0.1", timeout: 5000, field: "http.target"},
		{name: "userinfo redacted", target: "https://token:secret@status.example.com?key=private", timeout: 5000, field: "http.target"},
		{name: "timeout cap", target: "https://status.example.com", timeout: 30001, field: "http.timeoutMs"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := HTTPConfiguration{Target: test.target, Method: "GET", TimeoutMs: test.timeout}.Validate()
			if test.field == "" {
				if err != nil {
					t.Fatalf("Validate error = %v", err)
				}
				return
			}
			typed, ok := sharederrors.As(err)
			if !ok || typed.Code != sharederrors.CodeValidationFailed || typed.Details["field"] != test.field {
				t.Fatalf("Validate error = %#v", err)
			}
			if got := err.Error(); strings.Contains(got, "secret") || strings.Contains(got, "private") {
				t.Fatalf("error leaked target secret: %q", got)
			}
		})
	}
}

func TestHTTPConfigurationValidateOutboundPolicyMatrix(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		timeout int
		field   string
	}{
		{name: "safe public http", target: "http://status.example.com/health", timeout: 5000},
		{name: "safe public https", target: "https://status.example.com/health", timeout: 5000},
		{name: "non http scheme", target: "gopher://status.example.com", timeout: 5000, field: "http.target"},
		{name: "blocked literal", target: "http://10.0.0.1", timeout: 5000, field: "http.target"},
		{name: "blocked local alias", target: "https://LOCALHOST./health", timeout: 5000, field: "http.target"},
		{name: "excessive timeout", target: "https://status.example.com", timeout: 30001, field: "http.timeoutMs"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := HTTPConfiguration{Target: test.target, Method: "GET", TimeoutMs: test.timeout}.Validate()
			if test.field == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			typed, ok := sharederrors.As(err)
			if !ok || typed.Code != sharederrors.CodeValidationFailed || typed.Details["field"] != test.field {
				t.Fatalf("Validate() error = %#v", err)
			}
		})
	}
}

func TestMonitorValidateRejectsMissingServiceID(t *testing.T) {
	monitor := Monitor{
		MonitorID:         "public-http",
		TenantID:          "DEFAULT",
		Name:              "Homepage",
		Type:              MonitorTypeHTTP,
		IntervalSeconds:   60,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
	}

	if err := monitor.Validate(); err == nil {
		t.Fatal("Validate returned nil error, want failure for missing serviceId")
	}
}

func TestMonitorValidateAcceptsAllowedMinuteIntervals(t *testing.T) {
	for _, intervalSeconds := range AllowedIntervalSeconds() {
		monitor := validTestMonitor()
		monitor.IntervalSeconds = intervalSeconds
		if err := monitor.Validate(); err != nil {
			t.Fatalf("Validate(%d) returned error: %v", intervalSeconds, err)
		}
	}
}

func TestMonitorValidateRejectsUnsupportedInterval(t *testing.T) {
	monitor := validTestMonitor()
	monitor.IntervalSeconds = 90
	if err := monitor.Validate(); err == nil {
		t.Fatal("Validate returned nil error, want unsupported intervalSeconds failure")
	}
}

func TestMonitorToRecordRoundTrip(t *testing.T) {
	monitor := Monitor{
		MonitorID:         "public-http",
		ServiceID:         "auth",
		TenantID:          "DEFAULT",
		Name:              "Homepage",
		Type:              MonitorTypeHTTP,
		IntervalSeconds:   60,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		HTTP: &HTTPConfiguration{
			Target:              "https://example.com",
			Method:              "GET",
			TimeoutMs:           5000,
			ExpectedStatusCodes: []int{200},
		},
	}

	record, err := monitor.ToRecord()
	if err != nil {
		t.Fatalf("ToRecord returned error: %v", err)
	}
	if record.PK != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("PK = %q, want MONITOR#DEFAULT#AUTH#PUBLIC-HTTP", record.PK)
	}

	roundTrip, err := MonitorFromRecord(record)
	if err != nil {
		t.Fatalf("MonitorFromRecord returned error: %v", err)
	}
	if roundTrip.Name != monitor.Name {
		t.Fatalf("Name = %q, want %q", roundTrip.Name, monitor.Name)
	}
	if roundTrip.ServiceID != "auth" {
		t.Fatalf("ServiceID = %q, want auth", roundTrip.ServiceID)
	}
	if roundTrip.HTTP == nil || roundTrip.HTTP.Target != monitor.HTTP.Target {
		t.Fatal("HTTP target did not survive round trip")
	}
}

func TestServiceValidateRejectsUnsupportedServiceCategory(t *testing.T) {
	service := Service{TenantID: "DEFAULT", Name: "Auth", ServiceCategory: "bad-category"}
	if err := service.Validate(); err == nil {
		t.Fatal("Validate returned nil error, want unsupported serviceCategory failure")
	}
}

func TestServiceValidateAcceptsSupportedServiceCategories(t *testing.T) {
	for _, category := range SupportedServiceCategories() {
		service := Service{TenantID: "DEFAULT", Name: "Auth", ServiceCategory: category}
		if err := service.Validate(); err != nil {
			t.Fatalf("Validate(%q) returned error: %v", category, err)
		}
	}
}

func TestServiceValidateAllowsMissingServiceCategory(t *testing.T) {
	service := Service{TenantID: "DEFAULT", Name: "Auth"}
	if err := service.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestMonitorValidatePreservesFieldDetails(t *testing.T) {
	monitor := validTestMonitor()
	monitor.IntervalSeconds = 90

	typed, ok := sharederrors.As(monitor.Validate())
	if !ok {
		t.Fatal("Validate error is not typed")
	}
	if typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("code = %s, want %s", typed.Code, sharederrors.CodeValidationFailed)
	}
	if typed.Details["field"] != "intervalSeconds" {
		t.Fatalf("field = %v, want intervalSeconds", typed.Details["field"])
	}
	if typed.Details["reason"] != "must be one of: 60, 120, 180, 300, 600, 900, 1800, 3600" {
		t.Fatalf("reason = %v", typed.Details["reason"])
	}
}

func validTestMonitor() Monitor {
	return Monitor{
		MonitorID:         "public-http",
		ServiceID:         "auth",
		TenantID:          "DEFAULT",
		Name:              "Homepage",
		Type:              MonitorTypeHTTP,
		IntervalSeconds:   60,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		HTTP: &HTTPConfiguration{
			Target:    "https://example.com",
			Method:    "GET",
			TimeoutMs: 5000,
		},
	}
}
