package unpack

import (
	"errors"
	"fmt"
	"strings"
)

type Register struct {
	handler  Handler
	mutators map[string]Mutator[any]
}

func NewRegister(handler Handler) *Register {
	return &Register{
		handler:  handler,
		mutators: make(map[string]Mutator[any]),
	}
}

func (r *Register) AddMutator[T any](key string, mu Mutator[T]) {
	r.mutators[key] = func(v any, args MutatorArgs) (any, error) {
		typed, ok := v.(T)
		if !ok {
			var zero T
			return zero, ErrInvalidType
		}

		return mu(typed, args)
	}

}

func (r *Register) ApplyMutator[T any](key string, v T, args MutatorArgs) (T, error) {
	var zero T
	mu, ok := r.mutators[key]
	if !ok {
		return zero, ErrNotRegistered
	}

	result, err := mu(v, args)
	if err != nil {
		return zero, fmt.Errorf("unpack: mutator '%s' returned error: %w", key, err)
	}

	typed, ok := result.(T)
	if !ok {
		return zero, ErrUnexpectedReturnType
	}

	return typed, nil
}

func (r *Register) Unpack(v any, data map[string]any) error {
	return r.handler(r, v, data)
}

type MutatorArgMap map[string]any

// Bool implements [MutatorArgs].
func (m MutatorArgMap) Bool(key string, fallback ...bool) bool {
	v := m.Get(key, nil)
	switch vv := v.(type) {
	case bool:
		return vv
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return false
}

// Float implements [MutatorArgs].
func (m MutatorArgMap) Float(key string, fallback ...float64) float64 {
	v := m.Get(key, nil)
	switch vv := v.(type) {
	case float64:
		return vv
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return 0
}

// Get implements [MutatorArgs].
func (m MutatorArgMap) Get(key string, fallback ...any) any {
	v, has := m[key]
	if !has && len(fallback) > 0 {
		return fallback[0]
	}

	return v
}

// Has implements [MutatorArgs].
func (m MutatorArgMap) Has(keys ...string) bool {
	for _, k := range keys {
		if _, has := m[k]; !has {
			return false
		}
	}

	return true
}

// Int implements [MutatorArgs].
func (m MutatorArgMap) Int(key string, fallback ...int) int {
	v := m.Get(key, nil)
	switch vv := v.(type) {
	case int:
		return vv
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return 0
}

// MustHave implements [MutatorArgs].
func (m MutatorArgMap) MustHave(keys ...string) {
	if err := m.Require(keys...); err != nil {
		panic(err.Error())
	}
}

func (m MutatorArgMap) Require(keys ...string) error {
	var missing []string
	for _, k := range keys {
		if _, has := m[k]; !has {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return errors.New("unpack: mutator is missing the following args: " + strings.Join(missing, ", "))
	}

	return nil
}

// String implements [MutatorArgs].
func (m MutatorArgMap) String(key string, fallback ...string) string {
	v := m.Get(key, nil)
	switch vv := v.(type) {
	case string:
		return vv
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return ""
}

var _ MutatorArgs = (*MutatorArgMap)(nil)
