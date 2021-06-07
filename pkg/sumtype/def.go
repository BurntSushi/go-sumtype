package sumtype

import (
	"go/types"

	"golang.org/x/tools/go/analysis"
)

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
// any of the given declarations an error is reported and it is not added to
// the returned slice
func findSumTypeDefs(pass *analysis.Pass, decls []sumTypeDecl) []sumTypeDef {
	var defs []sumTypeDef
	for _, decl := range decls {
		def := newSumTypeDef(pass, decl.Package, decl)
		if def == nil {
			continue
		}
		defs = append(defs, *def)
	}
	return defs
}

// newSumTypeDef attempts to extract a sum type definition from a single
// package. If no such type corresponds to the given decl, then this function
// returns a nil def and an error is reported
//
// If the decl corresponds to a type that isn't an interface containing at
// least one unexported method, an error is reported
func newSumTypeDef(pass *analysis.Pass, pkg *types.Package, decl sumTypeDecl) *sumTypeDef {
	obj := pkg.Scope().Lookup(decl.TypeName)
	if obj == nil {
		pass.Reportf(decl.Pos, "type '%s' is not defined", decl.TypeName)
		return nil
	}
	iface, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		pass.Reportf(decl.Pos, "type '%s' is not an interface", decl.TypeName)
		return nil
	}
	hasUnexported := false
	for i := 0; i < iface.NumMethods(); i++ {
		if !iface.Method(i).Exported() {
			hasUnexported = true
			break
		}
	}
	if !hasUnexported {
		pass.Reportf(decl.Pos, "interface '%s' is not sealed "+
			"(sealing requires at least one unexported method)",
			decl.TypeName)
		return nil
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
		if types.Implements(ty, iface) || types.Implements(types.NewPointer(ty), iface) {
			def.Variants = append(def.Variants, obj)
		}
	}
	return def
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
		varty := indirect(v.Type())
		for _, ty := range tys {
			ty = indirect(ty)
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

// indirect dereferences through an arbitrary number of pointer types.
func indirect(ty types.Type) types.Type {
	if ty, ok := ty.(*types.Pointer); ok {
		return indirect(ty.Elem())
	}
	return ty
}
