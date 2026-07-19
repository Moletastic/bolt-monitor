package dynamodbschema

import "testing"

func TestExecutionWorkItemIsAddressableByRunID(t *testing.T) {
	first := ExecutionWorkItem("DEFAULT", "2026-07-18T10:00:00Z", "run_123", 1)
	second := ExecutionWorkItem("DEFAULT", "2026-07-18T10:01:00Z", "run_123", 1)
	if first.PK != second.PK || first.SK != "RUN_REQUEST#RUN_123" || first.SK != second.SK {
		t.Fatalf("work keys = %#v, %#v", first, second)
	}
}
