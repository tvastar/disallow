package disallow

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Names returns an analyzer that disallows specific names
//
// The namesmap defines the diagnostic message associated with each name.
// Names can be any identifier (such as "panic") or any name within a
// specific package name (such as "context.WithValue" or "http.StatusInternalServerError")
func Names(namesmap map[string]string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "disallow",
		Doc:      "checks for disallowed functions and packages",
		Run:      names(namesmap).run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

type names map[string]string

func (n names) run(pass *analysis.Pass) (interface{}, error) {
	// Call runFunc for each Func{Decl,Lit}.
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeTypes := []ast.Node{
		(*ast.FuncLit)(nil),
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeTypes, func(node ast.Node) {
		ast.Inspect(node, func(inner ast.Node) bool {
			for name, diagnostic := range n {
				if match(pass.TypesInfo, name, inner) {
					pass.Reportf(inner.Pos(), diagnostic)
				}
			}
			return true
		})
	})
	return nil, nil
}

func match(info *types.Info, name string, n ast.Node) bool {
	if x, ok := n.(*ast.Ident); ok {
		return x.Name == name
	}

	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}

	if s, ok := n.(*ast.SelectorExpr); ok {
		if s.Sel.Name != parts[1] {
			return false
		}
		x, ok := s.X.(*ast.Ident)
		if !ok {
			return false
		}

		if pkgname, ok := info.Uses[x].(*types.PkgName); ok {
			return pkgname.Imported().Path() == parts[0]
		}
		return false
	}

	return false
}
