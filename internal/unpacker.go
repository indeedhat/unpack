package internal

import (
	"bytes"
	"errors"
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

	"golang.org/x/mod/modfile"
)

type File struct {
	Tokens *token.FileSet
	File   *ast.File
}

type Unpacker struct {
	files map[string]*File

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
	u.files = make(map[string]*File)

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

		u.files[path] = &File{tokens, af}
		return nil
	})

	return err
}

func (u Unpacker) GenerateFiles() error {
	handlerArgs := handlerFileArgs{
		Mods: make(map[string]handlerFileMod),
	}

	for path, tokens := range u.files {
		if err := u.generateMapperFile(path, tokens); err != nil {
			return err
		}
		u.buildHandlerFileArgs(&handlerArgs, path, tokens)
	}

	var handlerBuf bytes.Buffer
	if err := buildHandlerFile(&handlerBuf, handlerArgs); err != nil {
		return err
	}

	log.Print(handlerBuf.String())
	handlerBytes, err := format.Source(handlerBuf.Bytes())
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(OutputRoot, HandlerFileName), handlerBytes, os.ModePerm)
}

func (u *Unpacker) buildHandlerFileArgs(args *handlerFileArgs, path string, file *File) error {
	dirFromCwd := strings.TrimPrefix(filepath.Dir(path), u.cwd)
	modPath := filepath.Join(u.modRoot, dirFromCwd)

	mod, found := args.Mods[modPath]
	if !found {
		mod = handlerFileMod{
			Alias:   len(args.Mods),
			GenPath: filepath.Join(u.modRoot, OutputRoot, dirFromCwd),
			ModPath: modPath,
		}
	}

	structs, err := extractStructs(file)
	if err != nil {
		return err
	}

	for _, tspec := range structs {
		mod.Structs = append(mod.Structs, tspec.Name.Name)
	}

	if len(mod.Structs) > 0 {
		args.Mods[modPath] = mod
	}

	return nil
}

func (u *Unpacker) generateMapperFile(path string, file *File) error {
	var output bytes.Buffer
	var write bool

	dirFromCwd := strings.TrimPrefix(filepath.Dir(path), u.cwd)
	pathFromCwd := strings.TrimPrefix(path, u.cwd)

	structs, err := extractStructs(file)
	if err != nil {
		return err
	}

	args := mapperFileArgs{
		Package:    file.File.Name.Name,
		TargetPath: filepath.Join(u.modRoot, dirFromCwd),
		Structs:    []mapperFileStruct{},
	}

	for _, tspec := range structs {
		sarg := mapperFileStruct{
			Name:   tspec.Name.Name,
			Fields: []mapperFileField{},
		}

		fields := extractFields(tspec.Type.(*ast.StructType))
		if len(fields) == 0 {
			continue
		}

		write = true
		for _, field := range fields {
			ft, err := fieldType(field, file.Tokens)
			if err != nil {
				return err
			}

			tag := newTagParser().Parse(field.Tag.Value)

			sarg.Fields = append(sarg.Fields, mapperFileField{
				Names: field.Names,
				Type:  stripPtr(ft),
				Dive:  tag.HasDive(),
				Tag:   tag,
				Ptr:   isPtr(ft),
			})
		}

		args.Structs = append(args.Structs, sarg)
	}

	if write {
		if err := buildMapperFile(&output, args); err != nil {
			return err
		}

		final, err := format.Source(output.Bytes())
		if err != nil {
			return err
		}

		if err := ensureCleanTargetDir(filepath.Join(u.cwd, OutputRoot, dirFromCwd)); err != nil {
			return err
		}

		return os.WriteFile(filepath.Join(OutputRoot, pathFromCwd), final, os.ModePerm)
	}

	return nil

}

func ensureCleanTargetDir(path string) error {
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		if err != os.RemoveAll(path) {
			return err
		}
	}

	return os.MkdirAll(path, os.ModePerm)
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
