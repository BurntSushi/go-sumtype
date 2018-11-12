package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatalf(`Usage: go-sumtype PATTERN ...

PATTERN may be a file name or a Go package pattern such as './things/...'.
`)
	}

	pkgs, errs := loadPackages(os.Args[1:])
	if len(errs) > 0 {
		exitWithErrors(errs)
	}
	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "go-sumtype: warning: No Go files or packages matched.\n")
		os.Exit(0)
	}
	for _, pkg := range pkgs {
		if errs := run(pkg); len(errs) > 0 {
			exitWithErrors(errs)
		}
	}
}

func run(pkg *packages.Package) []error {
	var errs []error

	decls, err := findSumTypeDecls(pkg)
	if err != nil {
		return []error{err}
	}

	defs, defErrs := findSumTypeDefs(pkg, decls)
	errs = append(errs, defErrs...)
	if len(defs) == 0 {
		return errs
	}

	if pkgErrs := check(pkg, defs); pkgErrs != nil {
		errs = append(errs, pkgErrs...)
	}
	return errs
}

func loadPackages(specs []string) ([]*packages.Package, []error) {
	var patterns []string
	for _, spec := range specs {
		patterns = append(patterns, fmt.Sprintf("pattern=%s", spec))
	}

	conf := packages.Config{
		Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(&conf, patterns...)
	if err != nil {
		return nil, []error{err}
	}
	var errs []error
	var result []*packages.Package
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			errs = append(errs, err)
		}
		if len(pkg.GoFiles) > 0 {
			result = append(result, pkg)
		}
	}
	return result, errs
}

func exitWithErrors(errs []error) {
	var list []string
	for _, err := range errs {
		list = append(list, err.Error())
	}
	log.Fatal(strings.Join(list, "\n"))
}
