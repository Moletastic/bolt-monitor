package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// manualRequestFingerprint canonicalizes the scoped command into one stable
// hash. The run command currently has no body parameters; any future body
// fields can be appended without changing the scoped key derivation.
func manualRequestFingerprint(tenantID, serviceID, monitorID, idempotencyKey string) string {
	payload := strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(tenantID)),
		strings.ToLower(strings.TrimSpace(serviceID)),
		strings.ToLower(strings.TrimSpace(monitorID)),
		"v1",
		strings.TrimSpace(idempotencyKey),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return "FP_" + strings.ToUpper(hex.EncodeToString(sum[:]))
}

// manualRunResponseFromRecord rebuilds the public response envelope from a
// stored idempotency record. The current design returns the reserved run
// identity plus the bounded TTL and replay state; full canonical result
// replay is a follow-up and depends on durable CheckRun storage.
func manualRunResponseFromRecord(record manualIdempotencyRecord) manualRunResponse {
	if record.Response != "" {
		var response manualRunResponse
		if json.Unmarshal([]byte(record.Response), &response) == nil {
			return response
		}
	}
	return manualRunResponse{
		RunID:     record.RunID,
		ServiceID: record.ServiceID,
		MonitorID: record.MonitorID,
		Trigger:   "manual",
	}
}
