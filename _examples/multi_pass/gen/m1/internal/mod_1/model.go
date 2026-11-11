package unpack_mod_1

import (
	target "github.com/indeedhat/unpack-example/internal/mod_1"

	"github.com/indeedhat/unpack"
)

func applyExampleMutations(r *unpack.Register, data map[string]any) error {
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

func UnpackExample(r *unpack.Register, t *target.Example, data map[string]any) error {
	if t == nil {
		t = &target.Example{}
	}
	if err := applyExampleMutations(r, data); err != nil {
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
			if err := UnpackExampleNested(r, &t.Nested, vv); err != nil {
				return err
			}
		}
	}
	return nil
}
func applyExampleNestedMutations(r *unpack.Register, data map[string]any) error {
	return nil
}

func UnpackExampleNested(r *unpack.Register, t *target.ExampleNested, data map[string]any) error {
	if t == nil {
		t = &target.ExampleNested{}
	}
	if err := applyExampleNestedMutations(r, data); err != nil {
		return err
	}
	if v, found := data["data"]; found {
		t.Data = v.(string)
	}
	return nil
}
