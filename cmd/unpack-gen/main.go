package main

import (
	"errors"

	"github.com/indeedhat/unpack/internal"
)

var ErrNotImplemented = errors.New("not implemented")

func main() {
	u, err := internal.NewUnpacker()
	if err != nil {
		panic(err)
	}

	if err = u.GatherFiles(); err != nil {
		panic(err)
	}

	for path, tokens := range u.Files {
		if u.GenerateFileFromTemplate(path, tokens); err != nil {
			panic(err)
		}
	}
}
