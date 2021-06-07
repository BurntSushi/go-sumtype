package sumtype

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

// sumTypeDecl is a declaration of a sum type in a Go source file.
type sumTypeDecl struct {
	// The package path that contains this decl.
	Package *types.Package
	// The type named by this decl.
	TypeName string
	// Position of the declaration
	Pos token.Pos
}

type filesToPkg map[*ast.File]*types.Package

// findSumTypeDecls searches every package given for sum type declarations of
// the form `go-sumtype:decl ...`.
func findSumTypeDecls(pass *analysis.Pass, ftp filesToPkg) []sumTypeDecl {
	var decls []sumTypeDecl
	for file, pkg := range ftp {
		pos := pass.Fset.Position(file.Pos())
		filename := pos.Filename
		if filepath.Base(filename) == "C" {
			// ignore (fake?) cgo files
			continue
		}

		fileDecls, err := sumTypeDeclSearch(filename)
		if err != nil {
			pass.Reportf(
				file.Pos(),
				"unknown error reading file '%s': %v",
				file.Name.String(), err)
			return nil
		}
		for i := range fileDecls {
			fileDecls[i].Package = pkg
			obj := pkg.Scope().Lookup(fileDecls[i].TypeName)
			if obj == nil {
				// TODO(ifross89): need to figure out how to create a more accurate position
				fileDecls[i].Pos = file.Pos()
			} else {
				fileDecls[i].Pos = obj.Pos()
			}
		}
		decls = append(decls, fileDecls...)
	}
	return decls
}

// sumTypeDeclSearch searches the given file for sum type declarations of the
// form `go-sumtype:decl ...`.
func sumTypeDeclSearch(path string) ([]sumTypeDecl, error) {
	var decls []sumTypeDecl

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lineNum := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if !isSumTypeDecl(line) {
			continue
		}
		ty := parseSumTypeDecl(line)
		if len(ty) == 0 {
			continue
		}
		decls = append(decls, sumTypeDecl{
			TypeName: ty,
		})
	}
	if err := scanner.Err(); err != nil {
		// A scanner can puke if it hits a line that is too long.
		// We assume such files won't contain any future decls and
		// otherwise move on.
		log.Printf("scan error reading '%s': %s", path, err)
	}
	return decls, f.Close()
}

var reParseSumTypeDecl = regexp.MustCompile(`^//go-sumtype:decl\s+(\S+)\s*$`)

// parseSumTypeDecl parses the type name out of a sum type decl.
//
// If no such decl could be found, then this returns an empty string.
func parseSumTypeDecl(line []byte) string {
	caps := reParseSumTypeDecl.FindSubmatch(line)
	if len(caps) < 2 {
		return ""
	}
	return string(caps[1])
}

// isSumTypeDecl returns true if and only if this line in a Go source file
// is a sum type decl.
func isSumTypeDecl(line []byte) bool {
	variant1, variant2 := []byte("//go-sumtype:decl "), []byte("//go-sumtype:decl\t")
	return bytes.HasPrefix(line, variant1) || bytes.HasPrefix(line, variant2)
}
