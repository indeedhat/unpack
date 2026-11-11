package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
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

type GeneratorConfig struct {
	TargetPath string
	Export     bool
	OutputRoot string
	TagName    string
	Verbose    bool
}

type Generator struct {
	files map[string]*File

	cwd       string
	targetDir string
	modRoot   string
	cfg       GeneratorConfig

	Log logger
}

func NewUnpacker(cfg *GeneratorConfig) (*Generator, error) {
	buildConfig(cfg)

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	data, err := os.ReadFile("go.mod")
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod file: %w", err)
	}

	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod file: %w", err)
	}

	var l logger
	if cfg.Verbose {
		l = consoleLogger{}
	} else {
		l = noopLogger{}
	}

	return &Generator{
		cwd:       cwd,
		targetDir: filepath.Join(cwd, cfg.TargetPath),
		modRoot:   mod.Module.Mod.Path,
		cfg:       *cfg,
		Log:       l,
	}, nil
}

func (g *Generator) GatherFiles() error {
	g.Log.Log("Gathering files...")
	g.files = make(map[string]*File)

	err := filepath.WalkDir(g.targetDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		tokens := token.NewFileSet()
		af, err := parser.ParseFile(tokens, path, nil, 0)
		if err != nil {
			return fmt.Errorf("failed to parse source file: %w", err)
		}

		g.files[path] = &File{tokens, af}
		return nil
	})

	g.Log.Logf("Found %d files", len(g.files))

	if err != nil {
		return fmt.Errorf("gathering files: %w", err)
	}

	return nil
}

func (g *Generator) GenerateFiles() error {
	handlerArgs := handlerFileArgs{
		Mods:   make(map[string]handlerFileMod),
		Export: g.cfg.Export,
	}

	g.Log.WithArgs("Cleaning gen directory...", "dir", g.cfg.OutputRoot)
	if err := ensureCleanTargetDir(g.cfg.OutputRoot); err != nil {
		return fmt.Errorf("cleaning target directory: %w", err)
	}

	for path, tokens := range g.files {
		g.Log.WithArgs("Generating file...", "path", path)
		if err := g.generateMapperFile(path, tokens); err != nil {
			return fmt.Errorf("generating mapper file (%s): %w", path, err)
		}

		g.Log.Log("Collecting handler symbols...")
		g.buildHandlerFileArgs(&handlerArgs, path, tokens)
	}

	var handlerBuf bytes.Buffer
	handlerPath := filepath.Join(g.cfg.OutputRoot, handlerFileName)
	g.Log.WithArgs("Building handler...", "path", handlerPath)
	if err := buildHandlerFile(&handlerBuf, handlerArgs); err != nil {
		return fmt.Errorf("building handler file: %w", err)
	}

	g.Log.Log("Formatting...")
	handlerBytes, err := format.Source(handlerBuf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting handler file: %w", err)
	}

	g.Log.Log("Writing...")

	if err = os.WriteFile(handlerPath, handlerBytes, os.ModePerm); err != nil {
		return fmt.Errorf("writing handler file: %w", err)
	}

	return nil
}

func (g *Generator) buildHandlerFileArgs(args *handlerFileArgs, path string, file *File) error {
	dirFromCwd := strings.TrimPrefix(filepath.Dir(path), g.cwd)
	modPath := filepath.Join(g.modRoot, dirFromCwd)

	mod, found := args.Mods[modPath]
	if !found {
		mod = handlerFileMod{
			Alias:   len(args.Mods),
			GenPath: filepath.Join(g.modRoot, g.cfg.OutputRoot, dirFromCwd),
			ModPath: modPath,
		}
	}

	structs, err := extractStructs(file)
	if err != nil {
		return fmt.Errorf("extracting structs: %w", err)
	}

	for _, tspec := range structs {
		mod.Structs = append(mod.Structs, tspec.Name.Name)
	}

	if len(mod.Structs) > 0 {
		args.Mods[modPath] = mod
	}

	return nil
}

func (g *Generator) generateMapperFile(path string, file *File) error {
	var output bytes.Buffer
	var write bool

	g.Log.Log("Collecting symbols...")
	dirFromCwd := strings.TrimPrefix(filepath.Dir(path), g.cwd)
	pathFromCwd := strings.TrimPrefix(path, g.cwd)

	structs, err := extractStructs(file)
	if err != nil {
		return fmt.Errorf("extracting structs: %w", err)
	}

	args := mapperFileArgs{
		Package:    file.File.Name.Name,
		TargetPath: filepath.Join(g.modRoot, dirFromCwd),
		Structs:    []mapperFileStruct{},
	}

	for _, tspec := range structs {
		sarg := mapperFileStruct{
			Name:   tspec.Name.Name,
			Fields: []mapperFileField{},
		}

		fields := g.extractFields(tspec.Type.(*ast.StructType))
		if len(fields) == 0 {
			continue
		}

		write = true
		for _, field := range fields {
			ft, err := fieldType(field, file.Tokens)
			if err != nil {
				return fmt.Errorf("building field type: %w", err)
			}

			tag := newTagParser(g.cfg.TagName).Parse(field.Tag.Value)

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

	g.Log.Log("Nothing to write...")
	if write {
		g.Log.Log("Building file contents...")
		if err := buildMapperFile(&output, args); err != nil {
			return fmt.Errorf("building mapper file: %w", err)
		}

		g.Log.Log("Formatting...")
		final, err := format.Source(output.Bytes())
		if err != nil {
			return fmt.Errorf("formatting mapper file: %w", err)
		}

		g.Log.Log("Writing...")
		_ = os.MkdirAll(filepath.Join(g.cfg.OutputRoot, dirFromCwd), os.ModePerm)
		if err := os.WriteFile(filepath.Join(g.cfg.OutputRoot, pathFromCwd), final, os.ModePerm); err != nil {
			return fmt.Errorf("writing mapper file: %w", err)
		}

		return nil
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

func (g *Generator) extractFields(s *ast.StructType) []*ast.Field {
	var fields []*ast.Field

	for _, field := range s.Fields.List {
		if field.Tag == nil {
			continue
		}

		tag := strings.Trim(field.Tag.Value, "`")
		st := reflect.StructTag(tag)
		if _, found := st.Lookup(g.cfg.TagName); !found {
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

func buildConfig(cfg *GeneratorConfig) {
	if cfg == nil {
		cfg = &GeneratorConfig{}
	}

	if cfg.OutputRoot == "" {
		cfg.OutputRoot = DefaultOutputRoot
	}

	if cfg.TagName == "" {
		cfg.TagName = DefaultUnpackTag
	}
}
