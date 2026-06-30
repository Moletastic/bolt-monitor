package rules

import (
	"errors"

	sharederrors "bolt-monitor/shared/errors"
)

type Rule[T any] func(T) error

func All[T any](rules ...Rule[T]) Rule[T] {
	return func(value T) error {
		for _, rule := range rules {
			if rule == nil {
				continue
			}
			if err := rule(value); err != nil {
				return err
			}
		}
		return nil
	}
}

func Any[T any](rules ...Rule[T]) Rule[T] {
	return func(value T) error {
		var failures []error
		for _, rule := range rules {
			if rule == nil {
				continue
			}
			if err := rule(value); err != nil {
				failures = append(failures, err)
				continue
			}
			return nil
		}
		return errors.Join(failures...)
	}
}

func Not[T any](rule Rule[T]) Rule[T] {
	return func(value T) error {
		if rule == nil {
			return nil
		}
		if err := rule(value); err == nil {
			return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"reason": "must not match"})
		}
		return nil
	}
}

func When[T any](pred func(T) bool, rule Rule[T]) Rule[T] {
	return func(value T) error {
		if pred == nil || rule == nil || !pred(value) {
			return nil
		}
		return rule(value)
	}
}

type Builder[T any] struct {
	rules []Rule[T]
}

func (b *Builder[T]) Add(rule Rule[T]) {
	if rule != nil {
		b.rules = append(b.rules, rule)
	}
}

func (b *Builder[T]) Build() Rule[T] {
	rules := append([]Rule[T](nil), b.rules...)
	return func(value T) error {
		failures := make([]error, 0, len(rules))
		for _, rule := range rules {
			if err := rule(value); err != nil {
				failures = append(failures, err)
			}
		}
		return errors.Join(failures...)
	}
}

func Field[T any](field string, rule Rule[T]) Rule[T] {
	return func(value T) error {
		if rule == nil {
			return nil
		}
		err := rule(value)
		if err == nil {
			return nil
		}
		if typed, ok := sharederrors.As(err); ok {
			return sharederrors.WithField(typed, field)
		}
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, err, map[string]any{"field": field})
	}
}
