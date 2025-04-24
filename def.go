package main

import (
	"fmt"
	"go/ast"
	"go/types"
)

// unsealedError corresponds to a declared sum type whose interface is not
// sealed. A sealed interface requires at least one unexported method.
type unsealedError struct {
	Decl sumTypeDecl
}

func (e unsealedError) Error() string {
	return fmt.Sprintf(
		"%s: interface '%s' is not sealed "+
			"(sealing requires at least one unexported method)",
		e.Decl.Location(), e.Decl.TypeName)
}

// notFoundError corresponds to a declared sum type whose type definition
// could not be found in the same Go package.
type notFoundError struct {
	Decl sumTypeDecl
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("%s: type '%s' is not defined", e.Decl.Location(), e.Decl.TypeName)
}

// notInterfaceError corresponds to a declared sum type that does not
// correspond to an interface.
type notInterfaceError struct {
	Decl sumTypeDecl
}

func (e notInterfaceError) Error() string {
	return fmt.Sprintf("%s: type '%s' is not an interface", e.Decl.Location(), e.Decl.TypeName)
}

// noVariantsError corresponds to a declared sum type for which no variants
// could be found in the same Go package.
type noVariantsError struct {
	Decl sumTypeDecl
}

func (e noVariantsError) Error() string {
	return fmt.Sprintf("%s: type '%s' has no variants", e.Decl.Location(), e.Decl.TypeName)
}

// sumTypeDef corresponds to the definition of a Go interface that is
// interpreted as a sum type. Its variants are determined by finding all types
// that implement said interface in the same package.
type sumTypeDef struct {
	Decl     sumTypeDecl
	Ty       *types.Interface
	Variants []types.Object
}

// findSumTypeDefs attempts to find a Go type definition for each of the given
// sum type declarations. If no such sum type definition could be found for
// any of the given declarations, then an error is returned.
func findSumTypeDefs(decls []sumTypeDecl) ([]sumTypeDef, []error) {
	var defs []sumTypeDef
	var errs []error
	for _, decl := range decls {
		def, err := newSumTypeDef(decl.Package.Types, decl)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if def == nil {
			errs = append(errs, notFoundError{decl})
			continue
		}
		defs = append(defs, *def)
	}
	return defs, errs
}

// newSumTypeDef attempts to extract a sum type definition from a single
// package. If no such type corresponds to the given decl, then this function
// returns a nil def and a nil error.
//
// If the decl corresponds to a type that isn't an interface containing at
// least one unexported method, then this returns an error.
func newSumTypeDef(pkg *types.Package, decl sumTypeDecl) (*sumTypeDef, error) {
	obj := pkg.Scope().Lookup(decl.TypeName)
	if obj == nil {
		return nil, nil
	}
	iface, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, notInterfaceError{decl}
	}
	hasUnexported := false
	for i := 0; i < iface.NumMethods(); i++ {
		if !iface.Method(i).Exported() {
			hasUnexported = true
			break
		}
	}
	if !hasUnexported {
		return nil, unsealedError{decl}
	}
	def := &sumTypeDef{
		Decl: decl,
		Ty:   iface,
	}
	for _, name := range pkg.Scope().Names() {
		obj, ok := pkg.Scope().Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		ty := obj.Type()
		if types.Identical(ty.Underlying(), iface) {
			continue
		}
		_, ok = ty.Underlying().(*types.Struct)
		if !ok {
			continue
		}
		if implementsBackwardsCompatible(ty, iface) {
			def.Variants = append(def.Variants, obj)
		}
	}
	if len(def.Variants) == 0 {
		return nil, noVariantsError{decl}
	}
	return def, nil
}

func implementsBackwardsCompatible(ty types.Type, iface *types.Interface) bool {
	// the old logic, persisted so that we don't accidentally break users
	if types.Implements(ty, iface) || types.Implements(types.NewPointer(ty), iface) {
		return true
	}

	if implementsUnifying(ty, iface) || implementsUnifying(types.NewPointer(ty), iface) {
		return true
	}

	return false
}

func (def *sumTypeDef) String() string {
	return def.Decl.TypeName
}

// missing returns a list of variants in this sum type that are not in the
// given list of types.
func (def *sumTypeDef) missing(tys []types.Type) []types.Object {
	// TODO(ag): This is O(n^2). Fix that. /shrug
	var missing []types.Object
	for _, v := range def.Variants {
		found := false
		varty := mustNamedObjType(v.Type())
		for _, ty := range tys {
			ty, ok := namedObjType(ty)
			if !ok {
				// worst case we'll cause a spurious error to the user,
				// almost as bad as panicking
				continue
			}
			if types.Identical(varty, ty) {
				found = true
			}
		}
		if !found {
			missing = append(missing, v)
		}
	}
	return missing
}

func namedObjType(ty types.Type) (types.Type, bool) {
	ty = indirect(ty)
	if ty, ok := ty.(*types.Named); ok {
		return ty.Obj().Type(), true
	}
	return nil, false
}

func mustNamedObjType(ty types.Type) types.Type {
	ty, ok := namedObjType(ty)
	if !ok {
		panic(fmt.Sprintf("expected named type, got %T (%+v)", ty, ty))
	}
	return ty
}

// indirect dereferences through an arbitrary number of pointer types.
func indirect(ty types.Type) types.Type {
	if ty, ok := ty.(*types.Pointer); ok {
		return indirect(ty.Elem())
	}
	return ty
}

func pkgFiles(pkg *ast.Package) []*ast.File {
	var files []*ast.File
	for _, file := range pkg.Files {
		files = append(files, file)
	}
	return files
}
