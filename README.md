# Unpack
Use go generate to generate efficient unpack logic for any struct

## Docs
I haven't gotten that far yet, see [_examples](_examples)

## Usage
```console
Usage of ./unpack-gen:
  -export
    	Export a new 'Register' from the handler.go file rather than applying the handler to the defalut Register instance
  -out string
    	the output directory for the generated files.
    	Multiple gen runs cannot share the same output dir. (default "unpack_gen/")
  -tag string
    	overwrite the struct tag targeted by the library (default "unpack")
  -target string
    	the target directory to use as the scan root
  -v	verbose output
```

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
