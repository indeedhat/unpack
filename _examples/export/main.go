package main

import (
	"errors"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/indeedhat/unpack"
	"github.com/indeedhat/unpack-example/internal"

	gen "github.com/indeedhat/unpack-example/unpack_gen"
)

//go:generate ../../unpack-gen -export

func main() {
	ex := internal.Example{}
	reg := gen.Register
	reg.AddMutator("trim", trim)
	reg.AddMutator("not_empty", notEmpty)
	reg.AddMutator("clamp", clamp)

	err := reg.Unpack(&ex, map[string]any{
		"string_data":  "datas ",
		"uncapped_int": 12,
		"nested": map[string]any{
			"data": "nested data",
		},
	})

	spew.Dump(ex)
	log.Print(err)
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
