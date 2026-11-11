package main

import (
	"flag"
	"log"
	"os"

	"github.com/indeedhat/unpack/generator"
)

func main() {
	log.Print(os.Args)
	cfg := &generator.GeneratorConfig{}

	flag.BoolVar(&cfg.Export, "export", false, "Export a new 'Register' from the handler.go file rather than applying the handler to the defalut Register instance")
	flag.StringVar(&cfg.TagName, "tag", generator.DefaultUnpackTag, "overwrite the struct tag targeted by the library")
	flag.StringVar(&cfg.OutputRoot, "out", generator.DefaultOutputRoot, "the output directory for the generated files.\nMultiple gen runs cannot share the same output dir.")
	flag.StringVar(&cfg.TargetPath, "target", "", "the target directory to use as the scan root")
	flag.BoolVar(&cfg.Verbose, "v", false, "verbose output")
	flag.Parse()

	gen, err := generator.NewUnpacker(cfg)
	if err != nil {
		log.Fatalf("setup failed: %s", err)
	}

	if err = gen.GatherFiles(); err != nil {
		log.Fatalf("gather failed: %s", err)
	}

	if err := gen.GenerateFiles(); err != nil {
		log.Fatalf("generate failed: %s", err)
	}

	gen.Log.Log("generation complete")
}
