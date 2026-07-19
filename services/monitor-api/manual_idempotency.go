package main

import (
	"fmt"
	"time"
)

type manualIdempotencyOutcome string

const (
	manualIdempotencyOutcomeReserved manualIdempotencyOutcome = "reserved"
	manualIdempotencyOutcomeCompleted manualIdempotencyOutcome = "completed"
)

// manualIdempotencyRecord represents one scoped key. Stored at the deterministic
// idempotency address derived from tenant/service/monitor/key. TTL bounds the
// replay window; manual runs always carry a fresh runID.
type manualIdempotencyRecord struct {
	TenantID         string
	ServiceID        string
	MonitorID        string
	Key              string
	Fingerprint      string
	Outcome          manualIdempotencyOutcome
	RunID            string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	TTL              int64
}

func newManualIdempotencyRecord(tenantID, serviceID, monitorID, key, fingerprint, runID string, now time.Time, ttl int64) manualIdempotencyRecord {
	return manualIdempotencyRecord{
		TenantID:    tenantID,
		ServiceID:   serviceID,
		MonitorID:   monitorID,
		Key:         key,
		Fingerprint: fingerprint,
		Outcome:     manualIdempotencyOutcomeReserved,
		RunID:       runID,
		CreatedAt:   now.UTC(),
		ExpiresAt:   now.UTC().Add(time.Duration(ttl) * time.Second),
		TTL:         ttl,
	}
}

func (r manualIdempotencyRecord) validate() error {
	if r.TenantID == "" {
		return fmt.Errorf("idempotency record missing tenant")
	}
	if r.Key == "" {
		return fmt.Errorf("idempotency record missing key")
	}
	if r.RunID == "" {
		return fmt.Errorf("idempotency record missing runID")
	}
	return nil
}