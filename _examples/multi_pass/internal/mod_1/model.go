package mod_1

type Example struct {
	StringField string        `unpack:"string_data;trim;not_empty"`
	Range       int           `unpack:"uncapped_int;clamp,min=1,max=10"`
	Nested      ExampleNested `unpack:"nested;dive"`
}

type ExampleNested struct {
	Data string `unpack:"data"`
}
