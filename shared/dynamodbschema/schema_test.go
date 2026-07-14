package dynamodbschema

import "testing"

func TestServiceStatusItemIncludesDashboardGSI(t *testing.T) {
	item := ServiceStatusItem("default", "auth", "degraded", "2026-05-16T10:00:00Z")
	if item.PK != "SERVICE#DEFAULT#AUTH" {
		t.Fatalf("PK = %q, want %q", item.PK, "SERVICE#DEFAULT#AUTH")
	}
	if item.GSI2PK != "TENANT#DEFAULT" {
		t.Fatalf("GSI2PK = %q, want %q", item.GSI2PK, "TENANT#DEFAULT")
	}
	if item.GSI2SK == "" {
		t.Fatal("GSI2SK is empty")
	}
}

func TestIncidentItemIncludesOpenIncidentGSI(t *testing.T) {
	item := IncidentItem("default", "auth", "public-http", "inc_123", "2026-05-16T10:00:00Z", "INCIDENT_OPEN")
	if item.GSI1PK != "TENANT#DEFAULT" {
		t.Fatalf("GSI1PK = %q, want %q", item.GSI1PK, "TENANT#DEFAULT")
	}
	if item.GSI1SK != "INCIDENT_OPEN#2026-05-16T10:00:00Z#INC_123" {
		t.Fatalf("GSI1SK = %q, want exact open incident key", item.GSI1SK)
	}
}

func TestAuditResourceItemIncludesMonitorScopeAndTimeOrder(t *testing.T) {
	item := AuditResourceItem("default", "auth", "public-http", "aud_123", "2026-05-16T10:00:00Z")
	if item.GSI3PK != "AUDIT_RESOURCE#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("GSI3PK = %q", item.GSI3PK)
	}
	if item.GSI3SK != "AUDIT#2026-05-16T10:00:00Z#AUD_123" {
		t.Fatalf("GSI3SK = %q", item.GSI3SK)
	}
}

func TestAuditResourceItemKeepsServiceScopeDistinct(t *testing.T) {
	item := AuditResourceItem("default", "auth", "", "aud_123", "2026-05-16T10:00:00Z")
	if item.GSI3PK != "AUDIT_RESOURCE#DEFAULT#AUTH#" {
		t.Fatalf("GSI3PK = %q", item.GSI3PK)
	}
}

func TestCheckRunItemCarriesTTL(t *testing.T) {
	item := CheckRunItem("default", "auth", "public-http", "2026-05-16T10:00:00Z", "run_123", 1780000000)
	if item.TTL != 1780000000 {
		t.Fatalf("TTL = %d, want %d", item.TTL, 1780000000)
	}
	if item.PK != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("PK = %q, want monitor composite key", item.PK)
	}
}

func TestExecutionWorkItemCarriesTTL(t *testing.T) {
	item := ExecutionWorkItem("default", "2026-05-16T10:00:00Z", "run_123", "iad", 1779559200)
	if item.TTL != 1779559200 {
		t.Fatalf("TTL = %d, want %d", item.TTL, 1779559200)
	}
	if item.PK != "TENANT#DEFAULT" {
		t.Fatalf("PK = %q, want tenant key", item.PK)
	}
}

func TestAccessPatternsCoversCoreReads(t *testing.T) {
	patterns := AccessPatterns()
	if len(patterns) < 10 {
		t.Fatalf("patterns = %d, want at least 10", len(patterns))
	}
	if patterns[0].Name != "list-services-for-tenant" {
		t.Fatalf("first pattern = %q, want %q", patterns[0].Name, "list-services-for-tenant")
	}
}

func TestEscalationItemsUseExpectedKeys(t *testing.T) {
	policy := EscalationPolicyItem("default", "primary")
	if policy.PK != "TENANT#DEFAULT" {
		t.Fatalf("policy PK = %q, want TENANT#DEFAULT", policy.PK)
	}
	if policy.SK != "ESCALATION_POLICY#PRIMARY" {
		t.Fatalf("policy SK = %q, want ESCALATION_POLICY#PRIMARY", policy.SK)
	}

	state := EscalationStateItem("default", "inc_123")
	if state.PK != "INCIDENT#INC_123" {
		t.Fatalf("state PK = %q, want INCIDENT#INC_123", state.PK)
	}
	if state.SK != "ESCALATION_STATE" {
		t.Fatalf("state SK = %q, want ESCALATION_STATE", state.SK)
	}
}
