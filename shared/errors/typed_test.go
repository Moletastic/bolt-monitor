package errors

import (
	stdlibErrors "errors"
	"fmt"
	"testing"
)

func TestTypedErrorStringNoCause(t *testing.T) {
	te := New(CodeValidationFailed, nil)
	if te.Error() != "VALIDATION_FAILED" {
		t.Fatalf("Error() = %q, want VALIDATION_FAILED", te.Error())
	}
}

func TestTypedErrorStringWithCause(t *testing.T) {
	cause := stdlibErrors.New("boom")
	te := Wrap(CodeInternal, cause, nil)
	want := "INTERNAL: boom"
	if te.Error() != want {
		t.Fatalf("Error() = %q, want %q", te.Error(), want)
	}
}

func TestTypedErrorUnwrapReturnsCause(t *testing.T) {
	cause := stdlibErrors.New("boom")
	te := Wrap(CodeInternal, cause, nil)
	if te.Unwrap() != cause {
		t.Fatalf("Unwrap did not return cause")
	}
}

func TestTypedErrorUnwrapNilCause(t *testing.T) {
	te := New(CodeValidationFailed, nil)
	if te.Unwrap() != nil {
		t.Fatalf("Unwrap = %v, want nil", te.Unwrap())
	}
}

func TestTypedErrorAsFindsWrapped(t *testing.T) {
	te := New(CodeValidationFailed, map[string]any{"field": "name"})
	wrapped := fmt.Errorf("wrap: %w", te)
	found, ok := As(wrapped)
	if !ok {
		t.Fatal("As did not find TypedError under fmt.Errorf wrap")
	}
	if found != te {
		t.Fatalf("As returned different pointer")
	}
	if found.Code != CodeValidationFailed {
		t.Fatalf("Code = %s, want VALIDATION_FAILED", found.Code)
	}
}

func TestTypedErrorAsMissesUnrelatedError(t *testing.T) {
	_, ok := As(fmt.Errorf("plain"))
	if ok {
		t.Fatal("As matched an unrelated error")
	}
}

func TestTypedErrorWithFieldImmutability(t *testing.T) {
	te := New(CodeValidationFailed, map[string]any{"field": "name", "reason": "required"})
	enriched := WithField(te, "tenantId")
	if enriched.Details["field"] != "tenantId" {
		t.Fatalf("WithField did not overwrite field")
	}
	if te.Details["field"] != "name" {
		t.Fatalf("WithField mutated original; original field = %v", te.Details["field"])
	}
	if enriched == te {
		t.Fatal("WithField returned same pointer")
	}
}

func TestTypedErrorWithFieldEmptyDetails(t *testing.T) {
	te := New(CodeValidationFailed, nil)
	enriched := WithField(te, "tenantId")
	if enriched.Details["field"] != "tenantId" {
		t.Fatalf("WithField did not set field on empty details")
	}
	if te.Details != nil {
		t.Fatalf("WithField created a Details map on the receiver: %v", te.Details)
	}
}

func TestTypedErrorWithFieldOverwritesPriorField(t *testing.T) {
	te := New(CodeValidationFailed, map[string]any{"field": "old", "reason": "required"})
	enriched := WithField(te, "new")
	if enriched.Details["field"] != "new" {
		t.Fatalf("WithField did not overwrite prior field; got %v", enriched.Details["field"])
	}
	if enriched.Details["reason"] != "required" {
		t.Fatalf("WithField dropped reason: %v", enriched.Details)
	}
}
