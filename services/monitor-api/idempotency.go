package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	sharederrors "bolt-monitor/shared/errors"
)

const (
	manualIdempotencyKeyMinLen     = 8
	manualIdempotencyKeyMaxLen     = 128
	manualIdempotencyRetentionDays = 30
)

// parseIdempotencyKey validates the inbound Idempotency-Key header. Empty values
// become a typed validation error so handlers can return immediately without
// performing durable side effects.
func parseIdempotencyKey(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "required"})
	}
	if len(trimmed) < manualIdempotencyKeyMinLen || len(trimmed) > manualIdempotencyKeyMaxLen {
		return "", sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "length must be between 8 and 128"})
	}
	return trimmed, nil
}

// manualIdempotencyAddress hashes the scoped key into one DynamoDB address
// shared by all manual-run retries. Inputs are normalized before hashing.
func manualIdempotencyAddress(tenantID, serviceID, monitorID, key string) string {
	payload := strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(tenantID)),
		strings.ToLower(strings.TrimSpace(serviceID)),
		strings.ToLower(strings.TrimSpace(monitorID)),
		strings.TrimSpace(key),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return "IDEMPOTENCY#" + strings.ToUpper(hex.EncodeToString(sum[:]))
}
