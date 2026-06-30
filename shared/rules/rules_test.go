package rules

import (
	"errors"
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

func TestAllShortCircuitsOnFirstFailure(t *testing.T) {
	called := false
	rule := All[int](
		func(int) error { return errors.New("first") },
		func(int) error {
			called = true
			return nil
		},
	)

	if err := rule(1); err == nil || err.Error() != "first" {
		t.Fatalf("All error = %v, want first", err)
	}
	if called {
		t.Fatal("All called rule after first failure")
	}
}

func TestAnyShortCircuitsOnFirstSuccess(t *testing.T) {
	called := false
	rule := Any[int](
		func(int) error { return errors.New("first") },
		func(int) error { return nil },
		func(int) error {
			called = true
			return nil
		},
	)

	if err := rule(1); err != nil {
		t.Fatalf("Any returned error: %v", err)
	}
	if called {
		t.Fatal("Any called rule after first success")
	}
}

func TestBuilderAggregatesFailures(t *testing.T) {
	var builder Builder[int]
	first := errors.New("first")
	second := errors.New("second")
	builder.Add(func(int) error { return first })
	builder.Add(func(int) error { return second })

	err := builder.Build()(1)
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("Build error = %v, want both failures", err)
	}
}

func TestFieldScopesTypedError(t *testing.T) {
	rule := Field[int]("intervalSeconds", func(int) error {
		return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"reason": "bad"})
	})

	typed, ok := sharederrors.As(rule(1))
	if !ok {
		t.Fatal("Field error is not typed")
	}
	if typed.Details["field"] != "intervalSeconds" {
		t.Fatalf("field = %v, want intervalSeconds", typed.Details["field"])
	}
	if typed.Details["reason"] != "bad" {
		t.Fatalf("reason = %v, want bad", typed.Details["reason"])
	}
}
