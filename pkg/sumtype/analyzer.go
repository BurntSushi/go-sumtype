package sumtype

import (
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "gosumtype",
	Doc:      "run exhaustiveness checks on type switch statements for sum types",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// pass.ResultOf[inspect.Analyzer] will be set if we've added inspect.Analyzer to Requires.
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.TypeSwitchStmt)(nil),
	}

	var (
		filesToPkg = map[*ast.File]*types.Package{}
		switches   []*ast.TypeSwitchStmt
	)

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		switch v := node.(type) {
		case *ast.File:
			filesToPkg[v] = pass.Pkg

		case *ast.TypeSwitchStmt:
			switches = append(switches, v)
		}
	})

	decls := findSumTypeDecls(pass, filesToPkg)
	if len(decls) == 0 {
		return nil, nil
	}

	defs := findSumTypeDefs(pass, decls)
	if len(defs) == 0 {
		return nil, nil
	}

	for _, swtch := range switches {
		checkSwitch(pass, defs, swtch)
	}

	return nil, nil
}
