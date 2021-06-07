package sumtype

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func missingNames(objs []types.Object) []string {
	var list []string
	for _, o := range objs {
		list = append(list, o.Name())
	}
	sort.Strings(list)
	return list
}

// checkSwitch performs an exhaustiveness check on the given type switch
// statement. If the type switch is used on a sum type and does not cover
// all variants of that sum type, then an error is returned indicating which
// variants were missed.
//
// Note that if the type switch contains a non-panicing default case, then
// exhaustiveness checks are disabled.
func checkSwitch(
	pass *analysis.Pass,
	defs []sumTypeDef,
	swtch *ast.TypeSwitchStmt,
) {
	def, missing := missingVariantsInSwitch(pass, defs, swtch)
	if len(missing) > 0 {
		pass.Reportf(
			swtch.Pos(),
			"exhaustiveness check failed for sum type '%s': missing cases for %s",
			def.Decl.TypeName, strings.Join(missingNames(missing), ", "))
	}
}

// missingVariantsInSwitch returns a list of missing variants corresponding to
// the given switch statement. The corresponding sum type definition is also
// returned. (If no sum type definition could be found, then no exhaustiveness
// checks are performed, and therefore, no missing variants are returned.)
func missingVariantsInSwitch(
	pass *analysis.Pass,
	defs []sumTypeDef,
	swtch *ast.TypeSwitchStmt,
) (*sumTypeDef, []types.Object) {
	asserted := findTypeAssertExpr(swtch)
	ty := pass.TypesInfo.TypeOf(asserted)
	def := findDef(defs, ty)
	if def == nil {
		return nil, nil
	}

	variantExprs, hasDefault := switchVariants(swtch)
	if hasDefault && !defaultClauseAlwaysPanics(swtch) {
		// A catch-all case defeats all exhaustiveness checks.
		return def, nil
	}

	var variantTypes []types.Type
	for _, expr := range variantExprs {
		variantTypes = append(variantTypes, pass.TypesInfo.TypeOf(expr))
	}

	return def, def.missing(variantTypes)
}

// switchVariants returns all case expressions found in a type switch. This
// includes expressions from cases that have a list of expressions.
func switchVariants(swtch *ast.TypeSwitchStmt) (exprs []ast.Expr, hasDefault bool) {
	for _, stmt := range swtch.Body.List {
		clause := stmt.(*ast.CaseClause)
		if clause.List == nil {
			hasDefault = true
		} else {
			exprs = append(exprs, clause.List...)
		}
	}
	return
}

// defaultClauseAlwaysPanics returns true if the given switch statement has a
// default clause that always panics. Note that this is done on a best-effort
// basis. While there will never be any false positives, there may be false
// negatives.
//
// If the given switch statement has no default clause, then this function
// panics.
func defaultClauseAlwaysPanics(swtch *ast.TypeSwitchStmt) bool {
	var clause *ast.CaseClause
	for _, stmt := range swtch.Body.List {
		c := stmt.(*ast.CaseClause)
		if c.List == nil {
			clause = c
			break
		}
	}
	if clause == nil {
		panic("switch statement has no default clause")
	}
	if len(clause.Body) != 1 {
		return false
	}
	exprStmt, ok := clause.Body[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	callExpr, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	fun, ok := callExpr.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	return fun.Name == "panic"
}

// findTypeAssertExpr extracts the expression that is being type asserted from a
// type swtich statement.
func findTypeAssertExpr(swtch *ast.TypeSwitchStmt) ast.Expr {
	var expr ast.Expr
	if assign, ok := swtch.Assign.(*ast.AssignStmt); ok {
		expr = assign.Rhs[0]
	} else {
		expr = swtch.Assign.(*ast.ExprStmt).X
	}
	return expr.(*ast.TypeAssertExpr).X
}

// findDef returns the sum type definition corresponding to the given type. If
// no such sum type definition exists, then nil is returned.
func findDef(defs []sumTypeDef, needle types.Type) *sumTypeDef {
	for i := range defs {
		def := &defs[i]
		if types.Identical(needle.Underlying(), def.Ty) {
			return def
		}
	}
	return nil
}
