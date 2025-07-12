package generator

import (
	"os"

	"go/ast"
	"go/parser"
	"go/token"
)

// FuncExists checks if a method with the given name exists for the specified struct
func FuncExists(filePath, structName, methodName string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return false
	}

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// Check if this is a method (has a receiver)
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				// Check receiver type
				if recv := funcDecl.Recv.List[0]; recv.Type != nil {
					var receiverType string
					switch t := recv.Type.(type) {
					case *ast.StarExpr:
						if ident, ok := t.X.(*ast.Ident); ok {
							receiverType = ident.Name
						}
					case *ast.Ident:
						receiverType = t.Name
					}

					// Check if this method belongs to our struct and has the right name
					if receiverType == structName && funcDecl.Name.Name == methodName {
						return true
					}
				}
			}
		}
	}

	return false
}
