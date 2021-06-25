package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/simon-engledew/gosrc/walk"
	"go/build"
	"golang.org/x/mod/modfile"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func printSources(w io.Writer, rel string) func(pkg *build.Package) error {
	seen := make(map[string]struct{})

	return func(pkg *build.Package) error {
		for _, src := range pkg.GoFiles {
			abspath := filepath.Join(pkg.Dir, src)
			if _, ok := seen[abspath]; ok {
				continue
			}
			seen[abspath] = struct{}{}
			relpath, err := filepath.Rel(rel, abspath)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, relpath); err != nil {
				return err
			}
		}
		return nil
	}
}

func main() {
	log.SetFlags(0)

	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("usage: %s <PATH>", os.Args[0])
	}

	dir := flag.Arg(0)

	data, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}

	modpath := modfile.ModulePath(data)

	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pkg, err := build.ImportDir(dir, build.ImportComment)
	if err != nil {
		panic(err)
	}
	pkg.Dir = filepath.Join(workdir, dir)

	isLocal := func(name string) bool {
		return strings.HasPrefix(name, modpath)
	}

	if err := walk.Walk(context.Background(), pkg, printSources(os.Stdout, workdir), isLocal); err != nil {
		panic(err)
	}
}
