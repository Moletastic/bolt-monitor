package errors

import "fmt"

type TypedError struct {
	Code    Code
	Details map[string]any
	Cause   error
	Message string
}

func (e *TypedError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Cause.Error())
	}
	return string(e.Code)
}

func (e *TypedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func New(code Code, details map[string]any) *TypedError {
	return &TypedError{Code: code, Details: details}
}

func Wrap(code Code, cause error, details map[string]any) *TypedError {
	return &TypedError{Code: code, Cause: cause, Details: details}
}

func WithField(err *TypedError, field string) *TypedError {
	cloned := &TypedError{
		Code:    err.Code,
		Cause:   err.Cause,
		Message: err.Message,
	}
	if err.Details != nil {
		cloned.Details = make(map[string]any, len(err.Details)+1)
		for k, v := range err.Details {
			cloned.Details[k] = v
		}
	} else {
		cloned.Details = map[string]any{}
	}
	cloned.Details["field"] = field
	return cloned
}

func As(err error) (*TypedError, bool) {
	var te *TypedError
	if err == nil {
		return nil, false
	}
	if !asError(err, &te) {
		return nil, false
	}
	return te, true
}

func asError(err error, target **TypedError) bool {
	for cur := err; cur != nil; {
		if typed, ok := cur.(*TypedError); ok {
			*target = typed
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := cur.(unwrapper)
		if !ok {
			return false
		}
		cur = u.Unwrap()
	}
	return false
}
