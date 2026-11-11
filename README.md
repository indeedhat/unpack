# Unpack
Use go generate to generate efficient unpack logic for any struct

## Docs
I haven't gotten that far yet, see [_examples](_examples)

## Schema
```go
type Example struct {
    Field string `unpack:"data_name;calllback,arg1=value,ag2=value;callback2"`
    Range int `unpack:"uncapped_int;between,min=1,max=10"`
}
```


## Todo
- [ ] write tests
- [ ] profile
- [ ] write docs
- [ ] better errors
- [ ] don't generate mutator if not needed
- [x] logging
- [x] add helpers for mutators to more easily handle required/typed args
- [x] sort out package structure
- [x] sort out central unpacker to avoid cyclic imports
