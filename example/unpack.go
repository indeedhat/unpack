package example

import "github.com/indeedhat/unpack/registry"

type Example struct {
	StringField string        `unpack:"string_data;trim;not_empty"`
	Range       int           `unpack:"uncapped_int;between,min=1,max=10"`
	Nested      ExampleNested `unpack:"nested;dive"`
}

type ExampleNested struct {
	Data string `unpack:"data"`
}

type DataProvider interface {
	Acquire() map[string]any
}

type MutationRegistry[T any] interface {
	Get(key string) func(T any) (T, error)
}

var reg MutationRegistry[any]

func Unpack[T any](target any, provider DataProvider) error {
	data := provider.Acquire()
	switch t := target.(type) {
	case *Example:
		if err := applyExampleMutations(data); err != nil {
			return err
		}
		if err := unpackExample(t, data); err != nil {
			return err
		}
	}

	return nil
}

func applyExampleMutations(data map[string]any) error {
	v, err := registry.Apply("trim", data["string_data"].(string), map[string]string{
		"min": "1",
		"max": "10",
	})
	if err != nil {
		return err
	} else {
		data["string_data"] = v
	}
	if v, err := reg.Get("Example.not_empty")(data["string_data"].(string)); err != nil {
		return err
	} else {
		data["string_data"] = v
	}
	if v, err := reg.Get("Example.between")(data["uncapped_int"].(int)); err != nil {
		return err
	} else {
		data["string_data"] = v
	}

	return nil
}

func unpackExample(target *Example, data map[string]any) error {
	if v, found := data["string_data"]; found {
		target.StringField = v.(string)
	}
	if v, found := data["uncapped_int"]; found {
		target.Range = v.(int)
	}
	if v, found := data["nested"]; found {
		if err := applyExampleNestedMutations(v.(map[string]any)); err != nil {
			return err
		}
		if err := unpackExampleNested(&ExampleNested{}, v.(map[string]any)); err != nil {
			return err
		}
	}
	return nil
}

func applyExampleNestedMutations(_data map[string]any) error {
	return nil
}

func unpackExampleNested(target *ExampleNested, data map[string]any) error {
	if v, found := data["data"]; found {
		target.Data = v.(string)
	}
	return nil
}
