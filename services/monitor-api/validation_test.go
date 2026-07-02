package main

import (
	"testing"

	sharederrors "bolt-monitor/shared/errors"
)

func TestValidateInputMapsRequiredField(t *testing.T) {
	type request struct {
		Name string `json:"name" validate:"notblank"`
	}

	err := validateInput(request{})
	assertValidationFailure(t, err, "name", "required")
}

func TestValidateInputMapsMaxLength(t *testing.T) {
	type request struct {
		Name string `json:"name" validate:"max=3"`
	}

	err := validateInput(request{Name: "toolong"})
	assertValidationFailure(t, err, "name", "must be 3 characters or less")
}

func TestValidateInputMapsMinimumSliceLength(t *testing.T) {
	type request struct {
		Items []string `json:"items" validate:"min=1"`
	}

	err := validateInput(request{})
	assertValidationFailure(t, err, "items", "must have at least one item")
}

func TestValidateInputMapsNestedJSONPath(t *testing.T) {
	type step struct {
		ChannelID string `json:"channelId" validate:"notblank"`
	}
	type path struct {
		Steps []step `json:"steps" validate:"min=1,dive"`
	}
	type request struct {
		BusinessHoursPath path `json:"businessHoursPath" validate:"required"`
	}

	err := validateInput(request{BusinessHoursPath: path{Steps: []step{{}}}})
	assertValidationFailure(t, err, "businessHoursPath.steps[0].channelId", "required")
}

func TestValidateInputReturnsDeterministicFirstFailure(t *testing.T) {
	type request struct {
		Name   string `json:"name" validate:"notblank"`
		Target string `json:"target" validate:"notblank"`
	}

	err := validateInput(request{})
	assertValidationFailure(t, err, "name", "required")
}

func assertValidationFailure(t *testing.T, err error, field, reason string) {
	t.Helper()
	typed, ok := sharederrors.As(err)
	if !ok {
		t.Fatalf("error = %v, want typed error", err)
	}
	if typed.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("code = %s, want %s", typed.Code, sharederrors.CodeValidationFailed)
	}
	if got := typed.Details["field"]; got != field {
		t.Fatalf("field = %v, want %s", got, field)
	}
	if got := typed.Details["reason"]; got != reason {
		t.Fatalf("reason = %v, want %s", got, reason)
	}
}
