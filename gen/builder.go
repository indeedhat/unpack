package main

import (
	_ "embed"
	"html/template"
	"io"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

//go:embed template/mapper.go.tmpl
var mapperFileTemplate string

func buildFromTemplate(w io.Writer, args any) error {
	spew.Dump(mapperFileTemplate)
	t, err := template.New("unpack").Funcs(template.FuncMap{
		"escape": func(v string) string {
			return strings.ReplaceAll(strings.ReplaceAll(v, `\`, `\\`), "'", `\'`)
		},
	}).Parse(mapperFileTemplate)
	log.Print("to here")

	if err != nil {
		return err
	}
	log.Print("to here 2")

	return t.Execute(w, args)
}
