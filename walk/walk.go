package walk

import (
	"context"
	"fmt"
	"go/build"
	"golang.org/x/sync/errgroup"
	"sync"
)

func walk(pkg *build.Package, fn func(pkg *build.Package) error, group *errgroup.Group, predicate Predicate) {
	group.Go(func() error {
		return fn(pkg)
	})

	for _, name := range pkg.Imports {
		if !predicate(name) {
			continue
		}
		name := name
		group.Go(func() error {
			dep, err := build.Import(name, ".", build.ImportComment)
			if err != nil {
				return fmt.Errorf("failed to import module required by %s: %w", pkg.Name, err)
			}
			if dep.Goroot {
				return nil
			}

			walk(dep, fn, group, predicate)

			return nil
		})
	}
}

type Predicate func(name string) bool

func combine(predicates []Predicate) Predicate {
	return func(name string) bool {
		for _, predicate := range predicates {
			if !predicate(name) {
				return false
			}
		}
		return true
	}
}

func Walk(ctx context.Context, pkg *build.Package, fn func(pkg *build.Package) error, predicates ...Predicate) error {
	seen := make(map[string]struct{})

	var mutex sync.Mutex
	hasSeen := func(name string) bool {
		mutex.Lock()
		defer mutex.Unlock()
		if _, ok := seen[name]; ok {
			return false
		}
		seen[name] = struct{}{}
		return true
	}

	group, _ := errgroup.WithContext(ctx)
	walk(pkg, fn, group, combine(append([]Predicate{hasSeen}, predicates...)))
	return group.Wait()
}
