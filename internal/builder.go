package internal

import (
	_ "embed"
	"html/template"
	"io"
	"strings"
)

//go:embed template/mapper.go.tmpl
var mapperTemplate string

//go:embed template/handler.go.tmpl
var handlerTemplate string

type mapperFileArgs struct {
	Package    string
	TargetPath string
	Structs    []mapperFileStruct
}

type mapperFileStruct struct {
	Name   string
	Fields []mapperFileField
}

type mapperFileField struct {
	Names any
	Type  string
	Dive  bool
	Tag   tag
	Ptr   bool
}

type handlerFileArgs struct {
	Mods map[string]handlerFileMod
}

type handlerFileMod struct {
	Alias   int
	GenPath string
	ModPath string
	Structs []string
}

func buildMapperFile(w io.Writer, args mapperFileArgs) error {
	t, err := template.New("unpack").Funcs(template.FuncMap{
		"escape": func(v string) string {
			return strings.ReplaceAll(strings.ReplaceAll(v, `\`, `\\`), "'", `\'`)
		},
	}).Parse(mapperTemplate)

	if err != nil {
		return err
	}

	return t.Execute(w, args)
}

func buildHandlerFile(w io.Writer, args handlerFileArgs) error {
	t, err := template.New("unpack").Parse(handlerTemplate)

	if err != nil {
		return err
	}

	return t.Execute(w, args)
}
