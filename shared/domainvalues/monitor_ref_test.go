package domainvalues

import (
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

func sampleRef() MonitorRef {
	return MustMonitorRef(DefaultTenantID, "auth", "public-http")
}

func TestNewMonitorRefRequiresAllFields(t *testing.T) {
	cases := []struct {
		name   string
		tenant TenantID
		svc    ServiceID
		mon    MonitorID
	}{
		{"missing tenant", "", "auth", "public-http"},
		{"missing service", DefaultTenantID, "", "public-http"},
		{"missing monitor", DefaultTenantID, "auth", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMonitorRef(tc.tenant, tc.svc, tc.mon)
			if err == nil {
				t.Fatal("expected validation error")
			}
			typed, ok := sharederrors.As(err)
			if !ok || typed.Code != sharederrors.CodeValidationFailed {
				t.Fatalf("expected VALIDATION_FAILED, got %v", err)
			}
		})
	}
}

func TestNewMonitorRefAcceptsValidTriple(t *testing.T) {
	ref, err := NewMonitorRef(DefaultTenantID, "auth", "public-http")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Tenant != "DEFAULT" {
		t.Fatalf("tenant = %q", ref.Tenant)
	}
	if ref.Service != "auth" {
		t.Fatalf("service = %q", ref.Service)
	}
	if ref.Monitor != "public-http" {
		t.Fatalf("monitor = %q", ref.Monitor)
	}
}

func TestMonitorRefStringFormat(t *testing.T) {
	ref := sampleRef()
	if ref.String() != "DEFAULT/auth/public-http" {
		t.Fatalf("String() = %q", ref.String())
	}
}

func TestMonitorRefStatusMapKeyMatchesLegacy(t *testing.T) {
	ref := sampleRef()
	want := "auth/public-http"
	if ref.StatusMapKey() != want {
		t.Fatalf("StatusMapKey() = %q, want %q", ref.StatusMapKey(), want)
	}
}

func TestMonitorRefPartitionKeys(t *testing.T) {
	ref := sampleRef()
	if got := ref.PartitionKey(); got != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("PartitionKey() = %q", got)
	}
	if got := ref.ServicePartitionKey(); got != "SERVICE#DEFAULT#AUTH" {
		t.Fatalf("ServicePartitionKey() = %q", got)
	}
	if got := ref.TenantPartitionKey(); got != "TENANT#DEFAULT" {
		t.Fatalf("TenantPartitionKey() = %q", got)
	}
	if got := ref.ServiceMonitorRefSK(); got != "MONITOR#PUBLIC-HTTP" {
		t.Fatalf("ServiceMonitorRefSK() = %q", got)
	}
	if got := ref.MetaSK(); got != "META" {
		t.Fatalf("MetaSK() = %q", got)
	}
	if got := ref.StatusSK(); got != "STATUS" {
		t.Fatalf("StatusSK() = %q", got)
	}
	if got := ref.RunPrefix(); got != "RUN#" {
		t.Fatalf("RunPrefix() = %q", got)
	}
	if got := ref.IncidentPrefix(); got != "INCIDENT#" {
		t.Fatalf("IncidentPrefix() = %q", got)
	}
}
