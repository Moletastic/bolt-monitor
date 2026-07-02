package main

import (
	"reflect"
	"strconv"
	"strings"

	sharederrors "bolt-monitor/shared/errors"
	"github.com/go-playground/validator/v10"
)

var inputValidator = newInputValidator()

func newInputValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return false
			}
			field = field.Elem()
		}
		if field.Kind() != reflect.String {
			return false
		}
		return strings.TrimSpace(field.String()) != ""
	})
	return v
}

func validateInput(value any) error {
	if err := inputValidator.Struct(value); err != nil {
		return validationErrorFromValidator(err)
	}
	return nil
}

func validationErrorFromValidator(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, err, nil)
	}
	failure := validationErrors[0]
	return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{
		"field":  fieldPath(failure),
		"reason": reasonForValidationFailure(failure),
	})
}

func fieldPath(failure validator.FieldError) string {
	path := failure.Namespace()
	if root, rest, ok := strings.Cut(path, "."); ok && root != "" {
		return rest
	}
	return failure.Field()
}

func reasonForValidationFailure(failure validator.FieldError) string {
	switch failure.Tag() {
	case "required", "notblank":
		return "required"
	case "max":
		return "must be " + failure.Param() + " characters or less"
	case "min":
		if failure.Kind() == reflect.Slice || failure.Kind() == reflect.Array {
			return "must have at least " + itemCount(failure.Param())
		}
		return "must be at least " + failure.Param()
	case "oneof":
		return "must be one of: " + failure.Param()
	default:
		return "failed " + failure.Tag() + " validation"
	}
}

func itemCount(param string) string {
	count, err := strconv.Atoi(param)
	if err != nil || count != 1 {
		return param + " items"
	}
	return "one item"
}
