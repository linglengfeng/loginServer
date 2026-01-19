package shell

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"regexp"
)

func autoImportForCode(code string) {
	if interpreter == nil || modName == "" {
		return
	}
	aliases := extractSelectorIdents(code)
	for alias := range aliases {
		if alias == "fmt" || alias == "reflect" || alias == "shell" {
			continue
		}
		if !isLocalTopPackageDir(alias) {
			continue
		}
		ensureImportedLocalAlias(alias)
	}
}

func extractSelectorIdents(code string) map[string]struct{} {
	out := make(map[string]struct{})
	e, err := parser.ParseExpr(code)
	if err != nil {
		return out
	}
	ast.Inspect(e, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		out[id.Name] = struct{}{}
		return true
	})
	return out
}

func ensureImportedLocalAlias(alias string) {
	importMu.Lock()
	defer importMu.Unlock()
	if importedAlias[alias] {
		return
	}
	stmt := fmt.Sprintf(`import %s "%s/%s"`, alias, modName, alias)
	if _, err := interpreter.Eval(stmt); err == nil {
		importedAlias[alias] = true
	}
}

func markImported(alias string) {
	importMu.Lock()
	importedAlias[alias] = true
	importMu.Unlock()
}

func isLocalTopPackageDir(pkg string) bool {
	if modRoot == "" {
		return false
	}
	p := filepath.Join(modRoot, pkg)
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// tryRunLocalCallWithGo 里会用到同一个模式，这里复用正则更稳定
var localCallRe = regexp.MustCompile(`^([a-zA-Z_]\w*)\.([A-Za-z_]\w*)\((.*)\)$`)

