package internal

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/mod/modfile"
)

type File struct {
	Tokens *token.FileSet
	File   *ast.File
}

type Unpacker struct {
	Files map[string]*File

	cwd     string
	modRoot string
}

func NewUnpacker() (*Unpacker, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile("go.mod")
	if err != nil {
		return nil, err
	}

	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, err
	}

	return &Unpacker{
		cwd:     cwd,
		modRoot: mod.Module.Mod.Path,
	}, nil
}

func (u *Unpacker) GatherFiles() error {
	u.Files = make(map[string]*File)

	err := filepath.WalkDir(u.cwd, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		tokens := token.NewFileSet()
		af, err := parser.ParseFile(tokens, path, nil, 0)
		if err != nil {
			return err
		}

		u.Files[path] = &File{tokens, af}
		return nil
	})

	return err
}

func (u *Unpacker) GenerateFileFromTemplate(path string, file *File) error {
	var output bytes.Buffer
	var write bool

	dirFromCwd := strings.TrimPrefix(filepath.Dir(path), u.cwd)
	pathFromCwd := strings.TrimPrefix(path, u.cwd)

	structs, err := extractStructs(file)
	if err != nil {
		return err
	}

	args := map[string]any{
		"package":    file.File.Name.Name,
		"targetPath": filepath.Join(u.modRoot, dirFromCwd),
		"structs":    make([]map[string]any, 0),
	}

	for _, tspec := range structs {
		sarg := map[string]any{
			"name":   tspec.Name.Name,
			"fields": make([]map[string]any, 0),
		}

		fields := extractFields(tspec.Type.(*ast.StructType))
		if len(fields) == 0 {
			continue
		}

		args["structs"] = append(args["structs"].([]map[string]any), sarg)

		write = true
		for _, field := range fields {
			ft, err := fieldType(field, file.Tokens)
			if err != nil {
				return err
			}

			tag := newTagParser().Parse(field.Tag.Value)

			sarg["fields"] = append(sarg["fields"].([]map[string]any), map[string]any{
				"names": field.Names,
				"type":  stripPtr(ft),
				"dive":  tag.HasDive(),
				"tag":   tag,
				"ptr":   isPtr(ft),
			})
		}
	}

	spew.Dump(args)

	if write {
		if err := buildFromTemplate(&output, args); err != nil {
			log.Print(err)
			return err
		}
		log.Print(output.String())

		final, err := format.Source(output.Bytes())
		if err != nil {
			log.Print(err)
			return err
		}

		if err := os.MkdirAll(filepath.Join(u.cwd, OutputRoot, dirFromCwd), os.ModePerm); err != nil {
			log.Print(err)
			return err
		}

		err = os.WriteFile(filepath.Join(OutputRoot, pathFromCwd), final, os.ModePerm)
		log.Print(err)
		return err
	}

	return nil

}

func extractStructs(file *File) ([]*ast.TypeSpec, error) {
	var structs []*ast.TypeSpec

	for _, decl := range file.File.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}

			structs = append(structs, ts)
		}
	}

	return structs, nil
}

func extractFields(s *ast.StructType) []*ast.Field {
	var fields []*ast.Field

	for _, field := range s.Fields.List {
		if field.Tag == nil {
			continue
		}

		tag := strings.Trim(field.Tag.Value, "`")
		st := reflect.StructTag(tag)
		if _, found := st.Lookup(UnpackTag); !found {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

func fieldType(field *ast.Field, fset *token.FileSet) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, field.Type)
	return buf.String(), err
}
