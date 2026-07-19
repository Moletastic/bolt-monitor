package dynamodbrecord

import (
	"strings"
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
)

func TestExecutionMarkerUsesBoundedRecoveryPartition(t *testing.T) {
	work := checkexecution.ExecutionWork{TenantID: "DEFAULT", RunID: "RUN_1", AcceptedAt: time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)}
	marker := NewExecutionMarkerRecord(work, ExecutionMarkerPublication, "2026071810", "03")
	if !strings.HasPrefix(marker.PK, "RECOVERY#DEFAULT#PUBLICATION#2026071810#03") || marker.SK != "MARKER#RUN_1" {
		t.Fatalf("marker = %#v", marker)
	}
}
