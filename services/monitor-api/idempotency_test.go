package main

import (
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

func TestParseIdempotencyKeyRejectsEmpty(t *testing.T) {
	if _, err := parseIdempotencyKey(""); err == nil {
		t.Fatal("expected validation error for empty idempotency key")
	} else if typed, ok := sharederrors.As(err); !ok || typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("err = %v, want typed validation error", err)
	}
}

func TestParseIdempotencyKeyRejectsShortAndLong(t *testing.T) {
	if _, err := parseIdempotencyKey("short"); err == nil {
		t.Fatal("expected validation error for short key")
	}
	long := make([]byte, manualIdempotencyKeyMaxLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := parseIdempotencyKey(string(long)); err == nil {
		t.Fatal("expected validation error for long key")
	}
}

func TestParseIdempotencyKeyNormalizesWhitespace(t *testing.T) {
	if key, err := parseIdempotencyKey("  abc12345  "); err != nil || key != "abc12345" {
		t.Fatalf("key = %q, err = %v", key, err)
	}
}

func TestManualIdempotencyAddressIsStableAndScoped(t *testing.T) {
	first := manualIdempotencyAddress("DEFAULT", "AUTH", "public-http", "key-123")
	second := manualIdempotencyAddress(" default ", "auth", "Public-HTTP", " key-123 ")
	if first != second {
		t.Fatalf("address differs across normalization: %q vs %q", first, second)
	}
	if first == manualIdempotencyAddress("DEFAULT", "AUTH", "public-http", "key-124") {
		t.Fatal("address did not change with key")
	}
}