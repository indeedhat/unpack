package unpack_gen

import (
	"github.com/indeedhat/unpack"
	unpack_t0 "github.com/indeedhat/unpack-example/gen/m1/internal/mod_1"
	t0 "github.com/indeedhat/unpack-example/internal/mod_1"
)

var Register = unpack.NewRegister(handler)

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
