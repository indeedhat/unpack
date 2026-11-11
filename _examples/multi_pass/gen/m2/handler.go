package unpack_gen

import (
	"github.com/indeedhat/unpack"
	unpack_t0 "github.com/indeedhat/unpack-example/gen/m2/internal/mod_2"
	t0 "github.com/indeedhat/unpack-example/internal/mod_2"
)

var Register = unpack.NewRegister(handler)

func handler(r *unpack.Register, v any, data map[string]any) error {
	switch vv := v.(type) {
	case *t0.Example2:
		return unpack_t0.UnpackExample2(r, vv, data)
	case *t0.ExampleNested2:
		return unpack_t0.UnpackExampleNested2(r, vv, data)
	default:
		return unpack.ErrNotFound
	}
}
