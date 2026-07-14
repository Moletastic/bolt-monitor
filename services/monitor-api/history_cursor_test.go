package main

import (
	"testing"

	sharedaws "bolt-monitor/shared/aws"
)

func TestHistoryCursorRoundTrip(t *testing.T) {
	key := map[string]sharedaws.AttributeValue{
		"PK": &sharedaws.AttributeValueMemberS{Value: "MONITOR#DEFAULT#AUTH#PUBLIC"},
		"SK": &sharedaws.AttributeValueMemberS{Value: "RUN#2026-05-16T10:00:00Z#RUN_123"},
	}
	raw, err := encodeHistoryCursor("MONITOR#DEFAULT#AUTH#PUBLIC", key)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeHistoryCursor(raw, "MONITOR#DEFAULT#AUTH#PUBLIC", "PK")
	if err != nil {
		t.Fatal(err)
	}
	if decoded["SK"].(*sharedaws.AttributeValueMemberS).Value != "RUN#2026-05-16T10:00:00Z#RUN_123" {
		t.Fatalf("decoded SK = %#v", decoded["SK"])
	}
}

func TestHistoryCursorRejectsResourceMismatch(t *testing.T) {
	raw, err := encodeHistoryCursor("MONITOR#DEFAULT#AUTH#PUBLIC", map[string]sharedaws.AttributeValue{
		"PK": &sharedaws.AttributeValueMemberS{Value: "MONITOR#DEFAULT#AUTH#PUBLIC"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeHistoryCursor(raw, "MONITOR#DEFAULT#PAYMENTS#PUBLIC", "PK"); err == nil {
		t.Fatal("expected cursor mismatch error")
	}
}
