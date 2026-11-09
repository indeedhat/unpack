package unpack

import "errors"

type Handler func(v any, data map[string]any) error

var handler Handler

var (
	ErrNoHandlerRegistered = errors.New("unpack handler not registered")
	ErrNotFound            = errors.New("no unpacker for type found")
)

func RegisterHandler(h Handler) {
	handler = h
}

func Struct(v any, data map[string]any) error {
	if handler == nil {
		return ErrNoHandlerRegistered
	}

	return handler(v, data)
}
