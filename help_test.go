package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func setupPackages(t *testing.T, code string) (string, []*packages.Package) {
	tmpdir, err := ioutil.TempDir("", "go-test-sumtype-")
	if err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(tmpdir, "src.go")
	if err := ioutil.WriteFile(srcPath, []byte(code), 0666); err != nil {
		t.Fatal(err)
	}
	pkgs, err := tycheckAll([]string{srcPath})
	if err != nil {
		t.Fatal(err)
	}
	return tmpdir, pkgs
}

func teardownPackage(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}
