package unpack_example

import (
	target "github.com/indeedhat/unpack/example"

	"github.com/indeedhat/unpack/registry"
)

func UnpackExample(t *target.Example, data map[string]any) error {
	if t == nil {
		*t = target.Example{}
	}
	if err := applyExampleMutations(data); err != nil {
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
			if err := UnpackExampleNested(&t.Nested, vv); err != nil {
				return err
			}
		}
	}
	return nil
}
func applyExampleMutations(data map[string]any) error {
	if v, err := registry.Apply("trim", data["string_data"].(string), nil); err != nil {
		return err
	} else {
		data["string_data"] = v
	}
	if v, err := registry.Apply("not_empty", data["string_data"].(string), nil); err != nil {
		return err
	} else {
		data["string_data"] = v
	}
	if v, err := registry.Apply("clamp", data["uncapped_int"].(int), map[string]string{
		"min": "1", "max": "10",
	}); err != nil {
		return err
	} else {
		data["uncapped_int"] = v
	}
	return nil
}
func UnpackExampleNested(t *target.ExampleNested, data map[string]any) error {
	if t == nil {
		*t = target.ExampleNested{}
	}
	if err := applyExampleNestedMutations(data); err != nil {
		return err
	}
	if v, found := data["data"]; found {
		t.Data = v.(string)
	}
	return nil
}
func applyExampleNestedMutations(data map[string]any) error {
	return nil
}
