package generator

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"regexp"
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
	Mods   map[string]handlerFileMod
	Export bool
}

type handlerFileMod struct {
	Alias   int
	GenPath string
	ModPath string
	Structs []string
}

var rawArgRe = regexp.MustCompile(`^(true|false|\d+|\d+\.\d+)$`)

func buildMapperFile(w io.Writer, args mapperFileArgs) error {
	t, err := template.New("unpack").Funcs(template.FuncMap{
		"parse_arg": func(v string) string {
			if rawArgRe.Match([]byte(v)) {
				return v
			}
			return fmt.Sprintf(`"%s"`, v)
		},
	}).Parse(mapperTemplate)

	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	return t.Execute(w, args)
}

func buildHandlerFile(w io.Writer, args handlerFileArgs) error {
	t, err := template.New("unpack").Parse(handlerTemplate)

	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	return t.Execute(w, args)
}
