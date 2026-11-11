package mod_2

type Example2 struct {
	StringField string         `unpack:"string_data;trim;not_empty"`
	Range       int            `unpack:"uncapped_int;clamp,min=1,max=10"`
	Nested      ExampleNested2 `unpack:"nested;dive"`
}

type ExampleNested2 struct {
	Data string `unpack:"data"`
}
