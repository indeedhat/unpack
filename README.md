# Unpack

## Schema
```go
type Example struct {
    Field string `unpack:"data_name;calllback,arg1=value,ag2=value;callback2"`
    Range int `unpack:"uncapped_int;between,min=1,max=10"`
}
```

## Todo
- [ ] better errors
- [ ] logging
- [ ] add helpers for mutators to more easily handle required/typed args
- [x] sort out package structure
- [x] sort out central unpacker to avoid cyclic imports
- [ ] profile
- [ ] write docs
