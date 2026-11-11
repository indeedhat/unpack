package main

import (
	"errors"
	"strings"

	"github.com/davecgh/go-spew/spew"

	unpack "github.com/indeedhat/unpack"
	gen1 "github.com/indeedhat/unpack-example/gen/m1"
	gen2 "github.com/indeedhat/unpack-example/gen/m2"
	"github.com/indeedhat/unpack-example/internal/mod_1"
	"github.com/indeedhat/unpack-example/internal/mod_2"
)

//go:generate ../../unpack-gen -out gen/m1 -target internal/mod_1 -export -v
//go:generate ../../unpack-gen -out gen/m2 -target internal/mod_2 -export -v

func main() {
	ex1 := mod_1.Example{}
	ex2 := mod_2.Example2{}

	g1 := gen1.Register
	g2 := gen2.Register

	g1.AddMutator("trim", trim)
	g1.AddMutator("not_empty", notEmpty)
	g1.AddMutator("clamp", clamp)

	g2.AddMutator("trim", trim)
	g2.AddMutator("not_empty", notEmpty)
	g2.AddMutator("clamp", clamp)

	err1 := g1.Unpack(&ex1, map[string]any{
		"string_data":  "data1 ",
		"uncapped_int": 5,
		"nested": map[string]any{
			"data": "nested data 1",
		},
	})
	err2 := g2.Unpack(&ex2, map[string]any{
		"string_data":  "data2 ",
		"uncapped_int": 15,
		"nested": map[string]any{
			"data": "nested data 2",
		},
	})

	spew.Dump(err1, ex1)
	spew.Dump(err2, ex2)
}

func trim(v string, args unpack.MutatorArgs) (string, error) {
	return strings.TrimSpace(v), nil
}

func notEmpty(v string, args unpack.MutatorArgs) (string, error) {
	if v == "" {
		return "", errors.New("value cannot be empty")
	}
	return v, nil
}

func clamp(v int, args unpack.MutatorArgs) (int, error) {
	if err := args.Require("min", "max"); err != nil {
		return 0, err
	}
	min := args.Int("min")
	max := args.Int("max")

	if v > max {
		return max, nil
	}
	if v < min {
		return min, nil
	}

	return v, nil
}
