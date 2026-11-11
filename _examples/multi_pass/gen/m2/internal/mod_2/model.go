package unpack_mod_2

import (
	target "github.com/indeedhat/unpack-example/internal/mod_2"

	"github.com/indeedhat/unpack"
)

func applyExample2Mutations(r *unpack.Register, data map[string]any) error {
	if v, err := r.ApplyMutator("trim", data["string_data"].(string), nil); err != nil {
		return err
	} else {
		data["string_data"] = v
	}
	if v, err := r.ApplyMutator("not_empty", data["string_data"].(string), nil); err != nil {
		return err
	} else {
		data["string_data"] = v
	}
	if v, err := r.ApplyMutator("clamp", data["uncapped_int"].(int), unpack.MutatorArgMap{
		"max": 10,
		"min": 1,
	}); err != nil {
		return err
	} else {
		data["uncapped_int"] = v
	}
	return nil
}

func UnpackExample2(r *unpack.Register, t *target.Example2, data map[string]any) error {
	if t == nil {
		t = &target.Example2{}
	}
	if err := applyExample2Mutations(r, data); err != nil {
		return err
	}
	if v, found := data["string_data"]; found {
		t.StringField = v.(string)
	}
	if v, found := data["uncapped_int"]; found {
		t.Range = v.(int)
	}
	if v, found := data["nested"]; found {
		vv, ok := v.(map[string]any)
		if ok {
			if err := UnpackExampleNested2(r, &t.Nested, vv); err != nil {
				return err
			}
		}
	}
	return nil
}
func applyExampleNested2Mutations(r *unpack.Register, data map[string]any) error {
	return nil
}

func UnpackExampleNested2(r *unpack.Register, t *target.ExampleNested2, data map[string]any) error {
	if t == nil {
		t = &target.ExampleNested2{}
	}
	if err := applyExampleNested2Mutations(r, data); err != nil {
		return err
	}
	if v, found := data["data"]; found {
		t.Data = v.(string)
	}
	return nil
}
