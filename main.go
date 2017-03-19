package main

import (
	"go/ast"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/loader"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go-sumtype <args>\n%s", loader.FromArgsUsage)
	}
	pkgpaths := os.Args[1:]
	prog, err := tycheckAll(pkgpaths)
	if err != nil {
		log.Fatal(err)
	}
	if errs := run(prog); len(errs) > 0 {
		var list []string
		for _, err := range errs {
			list = append(list, err.Error())
		}
		log.Fatal(strings.Join(list, "\n"))
	}
}

func run(prog *loader.Program) []error {
	var errs []error

	decls, err := findSumTypeDecls(prog)
	if err != nil {
		return []error{err}
	}

	defs, defErrs := findSumTypeDefs(prog, decls)
	errs = append(errs, defErrs...)
	if len(defs) == 0 {
		return errs
	}

	for _, pkg := range prog.InitialPackages() {
		if pkgErrs := check(prog, defs, pkg); pkgErrs != nil {
			errs = append(errs, pkgErrs...)
		}
	}
	return errs
}

func tycheckAll(pkgpaths []string) (*loader.Program, error) {
	conf := &loader.Config{
		AfterTypeCheck: func(info *loader.PackageInfo, files []*ast.File) {
		},
	}
	if _, err := conf.FromArgs(pkgpaths, true); err != nil {
		return nil, err
	}
	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}
	return prog, nil
}
