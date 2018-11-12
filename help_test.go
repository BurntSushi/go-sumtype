package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func setupPackage(t *testing.T, code string) (string, *packages.Package) {
	tmpdir, err := ioutil.TempDir("", "go-test-sumtype-")
	if err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(tmpdir, "src.go")
	if err := ioutil.WriteFile(srcPath, []byte(code), 0666); err != nil {
		t.Fatal(err)
	}
	pkgs, errs := loadPackages([]string{srcPath})
	if len(errs) > 0 {
		t.Fatal(errs[0])
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages returned by loadPackages()")
	}
	if len(pkgs) > 1 {
		t.Fatal("more than one package returned by loadPackages()")
	}
	return tmpdir, pkgs[0]
}

func teardownPackage(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}
