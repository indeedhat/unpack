package unpack_gen

import (
	"github.com/indeedhat/unpack"
	t0 "github.com/indeedhat/unpack-example/internal"
	unpack_t0 "github.com/indeedhat/unpack-example/unpack_gen/internal"
)

func init() {
	unpack.RegisterHandler(handler)
}

func handler(v any, data map[string]any) error {
	switch vv := v.(type) {
	case *t0.Example:
		return unpack_t0.UnpackExample(vv, data)
	case *t0.ExampleNested:
		return unpack_t0.UnpackExampleNested(vv, data)
	default:
		return unpack.ErrNotFound
	}
}
