package unpack_gen

import (
	"github.com/indeedhat/unpack"
	t0 "github.com/indeedhat/unpack-example/internal"
	unpack_t0 "github.com/indeedhat/unpack-example/unpack_gen/internal"
)

func init() {
	unpack.RegisterHandler(handler)
}

func handler(r *unpack.Register, v any, data map[string]any) error {
	switch vv := v.(type) {
	case *t0.Example:
		return unpack_t0.UnpackExample(r, vv, data)
	case *t0.ExampleNested:
		return unpack_t0.UnpackExampleNested(r, vv, data)
	default:
		return unpack.ErrNotFound
	}
}
