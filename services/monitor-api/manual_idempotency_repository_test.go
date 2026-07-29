package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
)

func TestReserveManualIdempotencyPersistsExpiryAsDynamoDBTTL(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoMonitorRepository(client, "table-name")
	record := newManualIdempotencyRecord("DEFAULT", "auth", "public-http", "key-1", "fingerprint", "RUN_1", now, manualIdempotencyRetentionSeconds)

	if _, err := repo.ReserveManualIdempotency(context.Background(), record); err != nil {
		t.Fatalf("ReserveManualIdempotency: %v", err)
	}

	if got, want := record.ExpiresAt, now.Add(time.Duration(manualIdempotencyRetentionSeconds)*time.Second); !got.Equal(want) {
		t.Fatalf("ExpiresAt = %s, want %s", got, want)
	}
	if got, want := record.TTL, record.ExpiresAt.Unix(); got != want {
		t.Fatalf("TTL = %d, want expiry epoch %d", got, want)
	}
	ttl, ok := client.putInput.Item["TTL"].(*sharedaws.AttributeValueMemberN)
	if !ok || ttl.Value != strconv.FormatInt(record.ExpiresAt.Unix(), 10) {
		t.Fatalf("persisted TTL = %#v, want Unix expiry %d", client.putInput.Item["TTL"], record.ExpiresAt.Unix())
	}
}

func TestReserveManualIdempotencyAllowsReplacingExpiredRecordDuringTTLCleanupLag(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoMonitorRepository(client, "table-name")
	record := newManualIdempotencyRecord("DEFAULT", "auth", "public-http", "key-1", "fingerprint", "RUN_2", now, manualIdempotencyRetentionSeconds)

	if _, err := repo.ReserveManualIdempotency(context.Background(), record); err != nil {
		t.Fatalf("ReserveManualIdempotency: %v", err)
	}

	if got, want := sharedaws.ToString(client.putInput.ConditionExpression), "attribute_not_exists(PK) OR TTL <= :now"; got != want {
		t.Fatalf("condition = %q, want %q", got, want)
	}
	conditionNow, ok := client.putInput.ExpressionAttributeValues[":now"].(*sharedaws.AttributeValueMemberN)
	if !ok || conditionNow.Value != strconv.FormatInt(now.Unix(), 10) {
		t.Fatalf("condition expiry boundary = %#v, want Unix time %d", client.putInput.ExpressionAttributeValues[":now"], now.Unix())
	}
}
