package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// isExternalPackageWithModule is a variant that takes a module path directly
func isExternalPackageWithModule(pkgPath, modulePath string) bool {
	// Special cases for internal packages
	if pkgPath == "main" {
		return false
	}

	// If the package path starts with our module name, it's internal
	if modulePath != "" && (pkgPath == modulePath || strings.HasPrefix(pkgPath, modulePath+"/")) {
		return false
	}

	// Skip standard library (packages without any path separators and dots)
	if !strings.Contains(pkgPath, "/") && !strings.Contains(pkgPath, ".") {
		return true
	}

	// Skip standard library and common external packages
	standardLibPrefixes := []string{
		"golang.org/x/",
		"gopkg.in/",
		"go.uber.org/",
	}

	for _, prefix := range standardLibPrefixes {
		if strings.HasPrefix(pkgPath, prefix) {
			return true
		}
	}

	// If it has a domain-like structure, it's likely external
	parts := strings.Split(pkgPath, "/")
	if len(parts) > 0 && strings.Contains(parts[0], ".") {
		return true
	}

	return false
}

// calculateComplexity calculates cyclomatic complexity for a function
func calculateComplexity(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 1
	}

	complexity := 1
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.FuncLit:
			complexity += 2
		}
		return true
	})

	return complexity
}

// extractFunctionSignature extracts a readable function signature
func extractFunctionSignature(pkg *packages.Package, fn *ast.FuncDecl) string {
	var parts []string

	// Add receiver for methods
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		receiverType := typeToString(pkg, recv.Type)
		parts = append(parts, fmt.Sprintf("(%s)", receiverType))
	}

	// Function name
	funcName := ""
	if fn.Name != nil {
		funcName = fn.Name.Name
	}

	// Parameters
	params := extractParameters(pkg, fn.Type.Params)

	// Return types
	returns := extractParameters(pkg, fn.Type.Results)

	// Build signature
	signature := funcName
	if len(parts) > 0 {
		signature = strings.Join(parts, " ") + " " + funcName
	}

	signature += "(" + params + ")"

	if returns != "" {
		if strings.Contains(returns, ",") {
			signature += " (" + returns + ")"
		} else {
			signature += " " + returns
		}
	}

	return signature
}

// extractParameters extracts parameter list as a string
func extractParameters(pkg *packages.Package, fieldList *ast.FieldList) string {
	if fieldList == nil || len(fieldList.List) == 0 {
		return ""
	}

	var params []string
	for _, field := range fieldList.List {
		typeStr := typeToString(pkg, field.Type)

		// Handle named parameters
		if len(field.Names) > 0 {
			names := make([]string, len(field.Names))
			for i, name := range field.Names {
				names[i] = name.Name
			}
			params = append(params, strings.Join(names, ", ")+" "+typeStr)
		} else {
			// Unnamed parameter
			params = append(params, typeStr)
		}
	}

	return strings.Join(params, ", ")
}

// typeToString converts an AST expression to a readable type string
func typeToString(pkg *packages.Package, expr ast.Expr) string {
	if t, ok := pkg.TypesInfo.Types[expr]; ok {
		return types.TypeString(t.Type, func(pkg *types.Package) string {
			if pkg == nil {
				return ""
			}
			// Use simple package name instead of full path
			parts := strings.Split(pkg.Path(), "/")
			return parts[len(parts)-1]
		})
	}

	// Fallback: try to extract from AST node directly
	switch n := expr.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.StarExpr:
		return "*" + typeToString(pkg, n.X)
	case *ast.SelectorExpr:
		return typeToString(pkg, n.X) + "." + n.Sel.Name
	case *ast.ArrayType:
		if n.Len == nil {
			return "[]" + typeToString(pkg, n.Elt)
		}
		return "[...]" + typeToString(pkg, n.Elt)
	case *ast.MapType:
		return "map[" + typeToString(pkg, n.Key) + "]" + typeToString(pkg, n.Value)
	case *ast.ChanType:
		switch n.Dir {
		case ast.SEND:
			return "chan<- " + typeToString(pkg, n.Value)
		case ast.RECV:
			return "<-chan " + typeToString(pkg, n.Value)
		default:
			return "chan " + typeToString(pkg, n.Value)
		}
	case *ast.FuncType:
		params := extractParameters(pkg, n.Params)
		results := extractParameters(pkg, n.Results)
		sig := "func(" + params + ")"
		if results != "" {
			if strings.Contains(results, ",") {
				sig += " (" + results + ")"
			} else {
				sig += " " + results
			}
		}
		return sig
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

// ExtractDocumentation extracts documentation comments from an AST node
func ExtractDocumentation(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}

	var lines []string
	for _, comment := range cg.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}

	return strings.Join(lines, " ")
}

// GetReceiverType extracts the receiver type name from a method
func GetReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	recvType := recv.List[0].Type
	// Remove pointer if present
	if star, ok := recvType.(*ast.StarExpr); ok {
		recvType = star.X
	}

	// Get the type name
	switch t := recvType.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	}

	return ""
}

// ExtractConstValue extracts the value from a constant declaration
func ExtractConstValue(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch v := expr.(type) {
	case *ast.BasicLit:
		return v.Value
	case *ast.Ident:
		return v.Name
	case *ast.UnaryExpr:
		return v.Op.String() + ExtractConstValue(v.X)
	case *ast.BinaryExpr:
		return ExtractConstValue(v.X) + " " + v.Op.String() + " " + ExtractConstValue(v.Y)
	default:
		return ""
	}
}
