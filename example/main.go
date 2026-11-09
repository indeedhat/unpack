package main

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/indeedhat/unpack"
	"github.com/indeedhat/unpack-example/internal"

	_ "github.com/indeedhat/unpack-example/unpack_gen"
)

//go:generate ../unpack-gen

func main() {
	ex := internal.Example{}
	unpack.RegisterMutator("trim", func(v string, args map[string]string) (string, error) {
		return strings.TrimSpace(v), nil
	})
	unpack.RegisterMutator("not_empty", func(v string, args map[string]string) (string, error) {
		if v == "" {
			return "", errors.New("value cannot be empty")
		}
		return v, nil
	})
	unpack.RegisterMutator("clamp", func(v int, args map[string]string) (int, error) {
		if _, ok := args["min"]; !ok {
			return -1, errors.New("min is a required argument")
		}
		if _, ok := args["max"]; !ok {
			return -1, errors.New("max is a required argument")
		}

		min, err := strconv.Atoi(args["min"])
		if err != nil {
			return -1, err
		}
		max, err := strconv.Atoi(args["max"])
		if err != nil {
			return -1, err
		}

		if v > max {
			return max, nil
		}
		if v < min {
			return min, nil
		}

		return v, nil
	})

	err := unpack.Struct(&ex, map[string]any{
		"string_data":  "datas ",
		"uncapped_int": 12,
		"nested": map[string]any{
			"data": "nested data",
		},
	})

	spew.Dump(ex)
	log.Print(err)
}
