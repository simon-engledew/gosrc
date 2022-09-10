package main

import (
	"flag"
	"fmt"
	"go/build"
	"golang.org/x/mod/modfile"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func walk(pkg *build.Package, root string, out map[string][]string) error {
	if pkg.Goroot {
		return nil
	}

	if _, ok := out[pkg.ImportPath]; ok {
		return nil
	}

	found := pkg.GoFiles

	for n, src := range found {
		abspath := filepath.Join(pkg.Dir, src)
		relpath, err := filepath.Rel(pkg.Root, abspath)
		if err != nil {
			return err
		}
		found[n] = relpath
	}

	out[pkg.ImportPath] = found

	for _, name := range pkg.Imports {
		if !strings.HasPrefix(name, root) {
			continue
		}
		if _, ok := out[name]; ok {
			continue
		}

		dep, err := build.Import(name, ".", build.ImportComment)
		if err != nil {
			return fmt.Errorf("failed to import %s: %w", name, err)
		}

		if err := walk(dep, root, out); err != nil {
			return err
		}
	}

	return nil
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
	pkg.Root = workdir

	found := make(map[string][]string)

	if err := walk(pkg, modpath, found); err != nil {
		panic(err)
	}

	for _, paths := range found {
		for _, path := range paths {
			fmt.Println(path)
		}
	}
}
