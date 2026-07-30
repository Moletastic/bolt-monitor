package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type commandIdempotencyState string

const (
	commandIdempotencyPending   commandIdempotencyState = "pending"
	commandIdempotencyCompleted commandIdempotencyState = "completed"
)

// commandIdempotencyRecord is a bounded, operation-scoped public result.
// Response holds only JSON already safe to return to an operator.
type commandIdempotencyRecord struct {
	TenantID, Operation, ResourceID, Key, Fingerprint, Response, ReservationToken string
	State                                                                         commandIdempotencyState
	CreatedAt                                                                     time.Time
	TTL                                                                           int64
}

func newCommandIdempotencyRecord(tenantID, operation, resourceID, key, fingerprint string, now time.Time, retention time.Duration) commandIdempotencyRecord {
	var token [16]byte
	_, _ = rand.Read(token[:])
	return commandIdempotencyRecord{TenantID: tenantID, Operation: operation, ResourceID: resourceID, Key: key, Fingerprint: fingerprint, ReservationToken: hex.EncodeToString(token[:]), State: commandIdempotencyPending, CreatedAt: now.UTC(), TTL: now.UTC().Add(retention).Unix()}
}

func commandIdempotencyAddress(tenantID, operation, resourceID, key string) string {
	payload := strings.Join([]string{strings.ToUpper(strings.TrimSpace(tenantID)), strings.ToLower(strings.TrimSpace(operation)), strings.ToLower(strings.TrimSpace(resourceID)), strings.TrimSpace(key)}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return "COMMAND_IDEMPOTENCY#" + strings.ToUpper(hex.EncodeToString(sum[:]))
}

func commandRequestFingerprint(tenantID, operation, resourceID, body string) string {
	payload := strings.Join([]string{strings.ToUpper(strings.TrimSpace(tenantID)), strings.ToLower(strings.TrimSpace(operation)), strings.ToLower(strings.TrimSpace(resourceID)), strings.TrimSpace(body)}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return "FP_" + strings.ToUpper(hex.EncodeToString(sum[:]))
}
