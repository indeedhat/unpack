package unpack

type Handler func(r *Register, v any, data map[string]any) error

type Mutator[T any] func(v T, args map[string]string) (T, error)

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
	r.mutators[key] = func(v any, args map[string]string) (any, error) {
		typed, ok := v.(T)
		if !ok {
			var zero T
			return zero, ErrInvalidType
		}

		return mu(typed, args)
	}

}

func (r *Register) ApplyMutator[T any](key string, v T, args map[string]string) (T, error) {
	var zero T
	mu, ok := r.mutators[key]
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

func (r *Register) Unpack(v any, data map[string]any) error {
	return r.handler(r, v, data)
}
