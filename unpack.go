package unpack

import "errors"

var defaultRegister = NewRegister(nil)

var (
	ErrNoHandlerRegistered = errors.New("unpack: no handler not registered")
	ErrNotFound            = errors.New("unpack: no unpacker for type found")

	ErrInvalidType          = errors.New("unpack: invalid type passed to mutator")
	ErrNotRegistered        = errors.New("unpack: mutator not registered")
	ErrUnexpectedReturnType = errors.New("unpack: mutator returned unexpected type")
)

func RegisterHandler(h Handler) {
	defaultRegister.handler = h
}

func Struct(v any, data map[string]any) error {
	if defaultRegister.handler == nil {
		return ErrNoHandlerRegistered
	}

	return defaultRegister.Unpack(v, data)
}

func AddMutator[T any](key string, mu Mutator[any]) {
	defaultRegister.AddMutator(key, mu)
}

func ApplyMutator[T any](key string, v T, args MutatorArgs) (T, error) {
	o, err := defaultRegister.ApplyMutator(key, v, args)
	return o, err
}
