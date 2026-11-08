package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/mod/modfile"
)

const (
	OutputRoot    = "unpack_gen/"
	DiveDirective = "dive"
	UnpackTag     = "unpack"
)

var ErrNotImplemented = errors.New("not implemented")

func main() {
	u, err := NewUnpacker()
	if err != nil {
		panic(err)
	}

	if err = u.gatherFiles(); err != nil {
		panic(err)
	}

	for path, tokens := range u.files {
		if u.generateFile(path, tokens); err != nil {
			panic(err)
		}
	}
}

type file struct {
	Tokens *token.FileSet
	File   *ast.File
}

type unpacker struct {
	files map[string]*file

	cwd     string
	modRoot string
}

func NewUnpacker() (*unpacker, error) {
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

	return &unpacker{
		cwd:     cwd,
		modRoot: mod.Module.Mod.Path,
	}, nil
}

func (u *unpacker) gatherFiles() error {
	u.files = make(map[string]*file)

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

		u.files[path] = &file{tokens, af}
		return nil
	})

	return err
}

func (u *unpacker) generateFile(path string, file *file) error {
	var output bytes.Buffer
	var write bool

	dirFromCwd := strings.TrimPrefix(filepath.Dir(path), u.cwd)
	pathFromCwd := strings.TrimPrefix(path, u.cwd)

	buildFileOpen(&output, file.File.Name.Name, filepath.Join(u.modRoot, dirFromCwd))

	structs, err := extractStructs(file)
	if err != nil {
		return err
	}

	for _, tspec := range structs {
		fields := extractFields(tspec.Type.(*ast.StructType))
		if len(fields) == 0 {
			continue
		}

		write = true
		if err := buildMapper(&output, tspec, fields, file.Tokens); err != nil {
			return err
		}
		if err := buildMutator(&output, tspec, fields, file.Tokens); err != nil {
			return err
		}
	}

	log.Print(write)
	log.Print(output.String())
	if write {
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

func extractStructs(file *file) ([]*ast.TypeSpec, error) {
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

func buildMapper(w io.Writer, ts *ast.TypeSpec, fields []*ast.Field, fset *token.FileSet) error {
	buildMapperOpen(w, ts.Name.Name)
	buildEnsureNotNil(w, ts.Name.Name)
	buildMutatorCall(w, ts.Name.Name)

	for _, field := range fields {
		tag := newTagParser().Parse(field.Tag.Value)
		for _, name := range field.Names {
			ft, err := fieldType(field, fset)
			if err != nil {
				return err
			}

			if tag.HasDive() {
				buildNestedAssgnment(w, tag.Key, name.Name, ft)
			} else {
				buildLeteralAssignement(w, tag.Key, name.Name, ft)
			}
		}
	}

	buildMapperClose(w)

	return nil
}

func buildMutator(w io.Writer, ts *ast.TypeSpec, fields []*ast.Field, fset *token.FileSet) error {
	buildMutatorOpen(w, ts.Name.Name)

	for _, field := range fields {
		ft, err := fieldType(field, fset)
		if err != nil {
			return err
		}

		tag := newTagParser().Parse(field.Tag.Value)
		for _, mu := range tag.Mutators {
			buildMutatorApply(w, tag.Key, ft, mu)
		}
	}

	buildMutatorClose(w)

	return nil
}
