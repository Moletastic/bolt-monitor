package main

import (
	"context"
	"sync"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
)

func TestReserveCommandIdempotencyConcurrentEquivalentRequestsConverge(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	repo := newDynamoMonitorRepository(&fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}, "table-name")
	record := newCommandIdempotencyRecord(defaultTenantID, "notification-channel-test", "CH_1", "test-key-1", "fingerprint", now, time.Hour)
	var first, second commandIdempotencyRecord
	var firstErr, secondErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		first, firstErr = repo.ReserveCommandIdempotency(context.Background(), record)
	}()
	go func() {
		defer wg.Done()
		second, secondErr = repo.ReserveCommandIdempotency(context.Background(), record)
	}()
	wg.Wait()
	if firstErr != nil || secondErr != nil || first.ReservationToken != second.ReservationToken {
		t.Fatalf("first=%+v err=%v second=%+v err=%v", first, firstErr, second, secondErr)
	}
}
