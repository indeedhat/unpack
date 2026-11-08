# Unpack

## Schema
```go
type Example struct {
    Field string `unpack:"data_name;calllback,arg1=value,ag2=value;callback2"`
    Range int `unpack:"uncapped_int;between,min=1,max=10"`
}
```
