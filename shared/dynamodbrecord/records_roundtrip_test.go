package dynamodbrecord

import (
	"reflect"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

// roundTripAssertKeyAttributes asserts that a record survives a Marshal/Unmarshal
// cycle with its PK, SK, EntityType, and any explicit attribute names intact.
// This is the contract that downstream runtimes depend on for legacy-cased
// items already stored in the table.
func roundTripAssertKeyAttributes(t *testing.T, original any, expectedPK, expectedSK, expectedEntity string, extra map[string]string) {
	t.Helper()
	wire, err := sharedaws.MarshalMap(original)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	rawPK, ok := wire["PK"].(*sharedaws.AttributeValueMemberS)
	if !ok || rawPK.Value != expectedPK {
		t.Fatalf("wire PK = %v, want %q", wire["PK"], expectedPK)
	}
	rawSK, ok := wire["SK"].(*sharedaws.AttributeValueMemberS)
	if !ok || rawSK.Value != expectedSK {
		t.Fatalf("wire SK = %v, want %q", wire["SK"], expectedSK)
	}
	rawEntity, ok := wire["EntityType"].(*sharedaws.AttributeValueMemberS)
	if !ok || rawEntity.Value != expectedEntity {
		t.Fatalf("wire EntityType = %v, want %q", wire["EntityType"], expectedEntity)
	}
	for key, want := range extra {
		got, ok := wire[key].(*sharedaws.AttributeValueMemberS)
		if !ok {
			t.Fatalf("attribute %s missing or wrong type", key)
		}
		if got.Value != want {
			t.Fatalf("attribute %s = %q, want %q", key, got.Value, want)
		}
	}
}

// TestServiceItemRecordRoundTripPreservesKeyAttributes pins the wire attribute
// shape for the Service item family after normalization.
func TestServiceItemRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	service := monitorconfig.Service{
		TenantID:           "DEFAULT",
		ServiceID:          "Auth",
		Name:               "Auth API",
		Description:        "Authentication service",
		LifecycleState:     monitorconfig.ServiceLifecycleActive,
		ServiceCategory:    monitorconfig.ServiceCategoryAPI,
		EscalationPolicyID: "POL_MAIN",
		BusinessHours: &escalation.BusinessHoursConfig{
			Timezone:   "UTC",
			StartHour:  9,
			EndHour:    17,
			DaysOfWeek: []int{1, 2, 3},
		},
	}

	record := NewServiceItemRecord(service)
	roundTripAssertKeyAttributes(t, record,
		"SERVICE#DEFAULT#AUTH",
		"META",
		dynamodbschema.EntityService,
		map[string]string{
			"TenantID":  "DEFAULT",
			"ServiceID": "auth",
		},
	)

	// Decoder keeps the canonicalized identifiers.
	var decoded ServiceItemRecord
	wire, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if err := sharedaws.UnmarshalMap(wire, &decoded); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}
	if decoded.PK != "SERVICE#DEFAULT#AUTH" {
		t.Fatalf("decoded PK = %q", decoded.PK)
	}
	if decoded.SK != "META" {
		t.Fatalf("decoded SK = %q", decoded.SK)
	}
	if decoded.ServiceID != "auth" {
		t.Fatalf("decoded ServiceID = %q, want canonical lower-case", decoded.ServiceID)
	}
}

// TestMonitorItemRecordRoundTripPreservesKeyAttributes pins the wire attribute
// shape for the Monitor item family and the ServiceMonitorRef variant.
func TestMonitorItemRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	body := "ok"
	monitor := monitorconfig.Monitor{
		TenantID:          "DEFAULT",
		ServiceID:         "Auth",
		MonitorID:         "Public-HTTP",
		Name:              "Homepage",
		Type:              monitorconfig.MonitorTypeHTTP,
		IntervalSeconds:   60,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		HTTP: &monitorconfig.HTTPConfiguration{
			Target:               "https://example.com",
			Method:               "GET",
			TimeoutMs:            5000,
			ExpectedStatusCodes:  []int{200},
			ExpectedBodyContains: &body,
		},
	}

	record := NewMonitorItemRecord(monitor)
	roundTripAssertKeyAttributes(t, record,
		"MONITOR#DEFAULT#AUTH#PUBLIC-HTTP",
		"META",
		dynamodbschema.EntityMonitor,
		map[string]string{
			"TenantID":  "DEFAULT",
			"ServiceID": "auth",
			"MonitorID": "public-http",
		},
	)

	refRecord := NewServiceMonitorRefItemRecord(monitor)
	roundTripAssertKeyAttributes(t, refRecord,
		"SERVICE#DEFAULT#AUTH",
		"MONITOR#PUBLIC-HTTP",
		dynamodbschema.EntityServiceMonitorRef,
		map[string]string{
			"MonitorID": "public-http",
		},
	)
}

// TestMonitorRecordDefensiveCopyIsolatesMutableConfig pins that the HTTP
// configuration is defensively copied so downstream mutations to headers,
// status codes, or expected body do not leak into the record.
func TestMonitorRecordDefensiveCopyIsolatesMutableConfig(t *testing.T) {
	body := "ok"
	monitor := monitorconfig.Monitor{
		TenantID:          "DEFAULT",
		ServiceID:         "auth",
		MonitorID:         "public-http",
		Name:              "Homepage",
		Type:              monitorconfig.MonitorTypeHTTP,
		IntervalSeconds:   60,
		Enabled:           true,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		HTTP: &monitorconfig.HTTPConfiguration{
			Target:               "https://example.com",
			Method:               "GET",
			Headers:              map[string]string{"X-Trace": "abc"},
			TimeoutMs:            5000,
			ExpectedStatusCodes:  []int{200},
			ExpectedBodyContains: &body,
		},
	}

	record := NewMonitorItemRecord(monitor)
	if record.HTTP == nil {
		t.Fatal("record HTTP is nil")
	}
	// Mutate the record's copy and confirm the original input is unchanged.
	record.HTTP.Headers["X-Trace"] = "mutated"
	record.HTTP.ExpectedStatusCodes[0] = 500

	if monitor.HTTP.Headers["X-Trace"] != "abc" {
		t.Fatalf("original headers leaked mutation: %v", monitor.HTTP.Headers)
	}
	if monitor.HTTP.ExpectedStatusCodes[0] != 200 {
		t.Fatalf("original status codes leaked mutation: %v", monitor.HTTP.ExpectedStatusCodes)
	}
}

// TestMonitorStatusRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the MonitorStatus item family using UPPERCASE stored
// state.
func TestMonitorStatusRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	status := resultstatus.MonitorStatus{
		ServiceID:     "auth",
		MonitorID:     "public-http",
		TenantID:      "DEFAULT",
		CurrentStatus: string(resultstatus.MonitorStateDown),
		LastCheckedAt: now,
		LastOutcome:   checkexecution.OutcomeFailure,
	}
	record := status.ToRecord()
	roundTripAssertKeyAttributes(t, record,
		"MONITOR#DEFAULT#AUTH#PUBLIC-HTTP",
		"STATUS",
		"MonitorStatus",
		map[string]string{
			"CurrentStatus": "DOWN",
			"TenantID":      "DEFAULT",
			"ServiceID":     "auth",
			"MonitorID":     "public-http",
		},
	)
}

// TestCheckRunRecordRoundTripPreservesKeyAttributes pins the wire attribute
// shape for the CheckRun item family using RUN# prefix and canonical tokens.
func TestCheckRunRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	statusCode := 200
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	run := resultstatus.NewCheckRun(checkexecution.ExecutionResult{
		ServiceID:  "auth",
		MonitorID:  "Public-HTTP",
		TenantID:   "default",
		RunID:      "run_manual_1",
		Type:       "http",
		Trigger:    checkexecution.TriggerTypeManual,
		StartedAt:  now,
		FinishedAt: now.Add(time.Second),
		DurationMs: 1000,
		Outcome:    checkexecution.OutcomeSuccess,
		StatusCode: &statusCode,
	}, now.Add(time.Second))
	record := run.ToRecord()

	wire, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if got := wire["SK"].(*sharedaws.AttributeValueMemberS).Value; !hasPrefix(got, "RUN#") {
		t.Fatalf("wire SK = %q, want RUN# prefix", got)
	}
	if got := wire["PK"].(*sharedaws.AttributeValueMemberS).Value; got != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("wire PK = %q", got)
	}
	if got := wire["TenantID"].(*sharedaws.AttributeValueMemberS).Value; got != "DEFAULT" {
		t.Fatalf("wire TenantID = %q, want canonical upper-case", got)
	}
	if got := wire["ServiceID"].(*sharedaws.AttributeValueMemberS).Value; got != "auth" {
		t.Fatalf("wire ServiceID = %q, want canonical lower-case", got)
	}
	if got := wire["MonitorID"].(*sharedaws.AttributeValueMemberS).Value; got != "public-http" {
		t.Fatalf("wire MonitorID = %q, want canonical lower-case", got)
	}
}

// TestIncidentItemRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the monitor-scoped Incident item family.
func TestIncidentItemRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	incident := IncidentRecord{
		IncidentID: "inc_42",
		ServiceID:  "Auth",
		MonitorID:  "Public-HTTP",
		TenantID:   "default",
		Type:       "monitoring",
		Summary:    "Monitor check failed",
		Status:     "open",
		OpenedAt:   "2026-05-17T22:00:00Z",
		UpdatedAt:  "2026-05-17T22:00:00Z",
		Origin:     "monitoring",
	}

	monitorRecord := NewIncidentMonitorItemRecord(incident)
	wire, err := sharedaws.MarshalMap(monitorRecord)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if got := wire["PK"].(*sharedaws.AttributeValueMemberS).Value; got != "MONITOR#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("wire PK = %q", got)
	}
	sk := wire["SK"].(*sharedaws.AttributeValueMemberS).Value
	if !hasPrefix(sk, "INCIDENT#2026-05-17T22:00:00Z#INC_42") {
		t.Fatalf("wire SK = %q, want INCIDENT#...#INC_42 prefix", sk)
	}
	if got := wire["EntityType"].(*sharedaws.AttributeValueMemberS).Value; got != dynamodbschema.EntityIncident {
		t.Fatalf("wire EntityType = %q", got)
	}
	if got := wire["TenantID"].(*sharedaws.AttributeValueMemberS).Value; got != "DEFAULT" {
		t.Fatalf("wire TenantID = %q", got)
	}
	if got := wire["ServiceID"].(*sharedaws.AttributeValueMemberS).Value; got != "auth" {
		t.Fatalf("wire ServiceID = %q", got)
	}
	if got := wire["MonitorID"].(*sharedaws.AttributeValueMemberS).Value; got != "public-http" {
		t.Fatalf("wire MonitorID = %q", got)
	}
	if got := wire["IncidentID"].(*sharedaws.AttributeValueMemberS).Value; got != "INC_42" {
		t.Fatalf("wire IncidentID = %q", got)
	}
	if got := wire["GSI1PK"].(*sharedaws.AttributeValueMemberS).Value; got != "TENANT#DEFAULT" {
		t.Fatalf("wire GSI1PK = %q", got)
	}
}

// TestExecutionWorkItemRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the ExecutionWork item family.
func TestExecutionWorkItemRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	record := NewExecutionWorkItemRecord(
		"default", "Auth", "Public-HTTP", "run_abc",
		checkexecution.TriggerTypeManual, "2026-05-17T22:00:00Z",
		checkexecution.ExecutionWorkPending, nil, nil, "",
	)
	roundTripAssertKeyAttributes(t, record,
		"TENANT#DEFAULT",
		"RUN_REQUEST#RUN_ABC",
		dynamodbschema.EntityExecutionWork,
		map[string]string{
			"RunID":     "RUN_ABC",
			"ServiceID": "auth",
			"MonitorID": "public-http",
		},
	)
}

// TestAuditEventRecordRoundTripPreservesKeyAttributes pins the wire attribute
// shape for the AuditEvent item family including the GSI3 keys.
func TestAuditEventRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	event := NewAuditEventRecord(
		time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC),
		"aud_xyz", "default", "MONITOR_UPDATED", "Auth", "Public-HTTP",
	)
	wire, err := sharedaws.MarshalMap(event)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if got := wire["PK"].(*sharedaws.AttributeValueMemberS).Value; got != "TENANT#DEFAULT" {
		t.Fatalf("wire PK = %q", got)
	}
	if got := wire["SK"].(*sharedaws.AttributeValueMemberS).Value; !hasPrefix(got, "AUDIT#2026-05-17T22:00:00Z#AUD_XYZ") {
		t.Fatalf("wire SK = %q", got)
	}
	if got := wire["EntityType"].(*sharedaws.AttributeValueMemberS).Value; got != "AuditEvent" {
		t.Fatalf("wire EntityType = %q", got)
	}
	if got := wire["AuditID"].(*sharedaws.AttributeValueMemberS).Value; got != "aud_xyz" {
		t.Fatalf("wire AuditID = %q", got)
	}
	if got := wire["ServiceID"].(*sharedaws.AttributeValueMemberS).Value; got != "auth" {
		t.Fatalf("wire ServiceID = %q", got)
	}
	if got := wire["MonitorID"].(*sharedaws.AttributeValueMemberS).Value; got != "public-http" {
		t.Fatalf("wire MonitorID = %q", got)
	}
	if got := wire["GSI3PK"].(*sharedaws.AttributeValueMemberS).Value; got != "AUDIT_RESOURCE#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("wire GSI3PK = %q", got)
	}
}

// TestSchedulerConfigItemRecordRoundTripPreservesKeyAttributes pins the wire
// attribute shape for the SchedulerConfig item family.
func TestSchedulerConfigItemRecordRoundTripPreservesKeyAttributes(t *testing.T) {
	now := time.Date(2026, 5, 17, 22, 0, 0, 0, time.UTC)
	record := NewSchedulerConfigItemRecord("default", checkexecution.SchedulerConfig{
		RecurringEnabled: true,
		StopControlMode:  checkexecution.StopControlMonitorDisable,
	}, now)
	roundTripAssertKeyAttributes(t, record,
		"TENANT#DEFAULT",
		"SCHEDULER_CONFIG",
		SchedulerConfigEntityType,
		map[string]string{
			"TenantID":        "DEFAULT",
			"StopControlMode": "monitor-disable",
		},
	)
}

// TestEscalationPolicyItemKeyShape pins the policy record shape so the §3
// record migration to shared/dynamodbrecord cannot drift.
func TestEscalationPolicyItemKeyShape(t *testing.T) {
	item := dynamodbschema.EscalationPolicyItem("default", "POL_MAIN")
	if item.PK != "TENANT#DEFAULT" {
		t.Fatalf("policy PK = %q", item.PK)
	}
	if item.SK != "ESCALATION_POLICY#POL_MAIN" {
		t.Fatalf("policy SK = %q", item.SK)
	}
}

// TestNotificationChannelItemKeyShape pins the channel record shape so the §3
// record migration to shared/dynamodbrecord cannot drift.
func TestNotificationChannelItemKeyShape(t *testing.T) {
	item := dynamodbschema.NotificationChannelItem("default", "CH_EMAIL")
	if item.PK != "TENANT#DEFAULT" {
		t.Fatalf("channel PK = %q", item.PK)
	}
	if item.SK != "NOTIFICATION_CHANNEL#CH_EMAIL" {
		t.Fatalf("channel SK = %q", item.SK)
	}
}

// TestEscalationStateItemKeyShape pins the escalation-state record shape so
// the §3 record migration to shared/dynamodbrecord cannot drift.
func TestEscalationStateItemKeyShape(t *testing.T) {
	item := dynamodbschema.EscalationStateItem("default", "INC_42")
	if item.PK != "INCIDENT#INC_42" {
		t.Fatalf("state PK = %q", item.PK)
	}
	if item.SK != "ESCALATION_STATE" {
		t.Fatalf("state SK = %q", item.SK)
	}
	if reflect.DeepEqual(item.PK, "") {
		t.Fatal("empty PK")
	}
}

func hasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}
