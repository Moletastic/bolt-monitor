package dynamodbrecord

import (
	"testing"
	"time"
)

func TestNewAuditEventRecordIncludesResourceIndexKeys(t *testing.T) {
	event := NewAuditEventRecord(time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC), "aud_123", "default", "MONITOR_UPDATED", "auth", "public-http")
	if event.GSI3PK != "AUDIT_RESOURCE#DEFAULT#AUTH#PUBLIC-HTTP" {
		t.Fatalf("GSI3PK = %q", event.GSI3PK)
	}
	if event.GSI3SK != "AUDIT#2026-05-16T10:00:00Z#AUD_123" {
		t.Fatalf("GSI3SK = %q", event.GSI3SK)
	}
}
