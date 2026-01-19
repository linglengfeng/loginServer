package shell

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// tryResolveLocalLiteral 解析形如 pkg.NAME 的本地包 const/var（仅支持基础字面量：string/int/bool）
func tryResolveLocalLiteral(expr string) (string, bool) {
	re := regexp.MustCompile(`^([a-zA-Z_]\w*)\.([A-Za-z_]\w*)$`)
	m := re.FindStringSubmatch(strings.TrimSpace(expr))
	if len(m) != 3 {
		return "", false
	}
	pkg := m[1]
	name := m[2]
	if !isLocalTopPackageDir(pkg) {
		return "", false
	}
	val, ok := resolveLiteralFromPackageDir(filepath.Join(modRoot, pkg), name)
	return val, ok
}

func resolveLiteralFromPackageDir(dir string, ident string) (string, bool) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		n := fi.Name()
		return strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go")
	}, 0)
	if err != nil {
		return "", false
	}
	for _, p := range pkgs {
		for _, f := range p.Files {
			var prevConstExpr ast.Expr
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				if gen.Tok != token.CONST && gen.Tok != token.VAR {
					continue
				}
				for _, spec := range gen.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, n := range vs.Names {
						if n.Name != ident {
							if gen.Tok == token.CONST {
								if len(vs.Values) > 0 {
									j := i
									if j >= len(vs.Values) {
										j = len(vs.Values) - 1
									}
									if j >= 0 {
										prevConstExpr = vs.Values[j]
									}
								}
							}
							continue
						}

						var e ast.Expr
						if len(vs.Values) > 0 {
							j := i
							if j >= len(vs.Values) {
								j = len(vs.Values) - 1
							}
							if j >= 0 {
								e = vs.Values[j]
							}
						} else if gen.Tok == token.CONST {
							e = prevConstExpr
						}
						if e == nil {
							continue
						}
						if s, ok := evalBasicLiteral(e); ok {
							return s, true
						}
						return "", false
					}
				}
			}
		}
	}
	return "", false
}

func evalBasicLiteral(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.ParenExpr:
		return evalBasicLiteral(v.X)
	case *ast.UnaryExpr:
		if v.Op == token.SUB || v.Op == token.ADD {
			if s, ok := evalBasicLiteral(v.X); ok {
				if v.Op == token.SUB {
					if strings.HasPrefix(s, "-") {
						return s, true
					}
					return "-" + s, true
				}
				return s, true
			}
		}
		return "", false
	case *ast.BasicLit:
		if v.Kind == token.STRING {
			return strings.Trim(v.Value, "\"`"), true
		}
		return v.Value, true
	case *ast.Ident:
		if v.Name == "true" || v.Name == "false" {
			return v.Name, true
		}
	}
	return "", false
}

