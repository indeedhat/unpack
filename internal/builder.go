package internal

import (
	_ "embed"
	"html/template"
	"io"
	"strings"
)

//go:embed template/mapper.go.tmpl
var mapperFileTemplate string

func buildFromTemplate(w io.Writer, args any) error {
	t, err := template.New("unpack").Funcs(template.FuncMap{
		"escape": func(v string) string {
			return strings.ReplaceAll(strings.ReplaceAll(v, `\`, `\\`), "'", `\'`)
		},
	}).Parse(mapperFileTemplate)

	if err != nil {
		return err
	}

	return t.Execute(w, args)
}
