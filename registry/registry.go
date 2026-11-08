package registry

import "errors"

type Mutator[T any] func(v T, args map[string]string) (T, error)

var reg = make(map[string]Mutator[any])

var (
	ErrInvalidType          = errors.New("invalid type passed to mutator")
	ErrNotRegistered        = errors.New("mutator not registered")
	ErrUnexpectedReturnType = errors.New("mutator returned unexpected type")
)

func Register[T any](key string, mu Mutator[T]) {
	reg[key] = func(v any, args map[string]string) (any, error) {
		typed, ok := v.(T)
		if !ok {
			var zero T
			return zero, ErrInvalidType
		}

		return mu(typed, args)
	}
}

func Apply[T any](key string, v T, args map[string]string) (T, error) {
	var zero T
	mu, ok := reg[key]
	if !ok {
		return zero, ErrNotRegistered
	}

	result, err := mu(v, args)
	if err != nil {
		return zero, err
	}

	typed, ok := result.(T)
	if !ok {
		return zero, ErrUnexpectedReturnType
	}

	return typed, nil
}
