package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

var (
	codes = []string{`
package A

import (
    "B"
    "errors"
)

// DoSthReplyOnB reply on B.Flag
func DoSthReplyOnB() error {
    if B.Flag {
        return nil
    } else {
        return errors.New("bad flag")
    }
}
`, `
package B

import (
    "A"
    "fmt"
)

const (
    Flag = true
)

func innerFunction() int {
    return 0
}

// DoSthReplyOnA reply on DoSthReplyOnB
func DoSthReplyOnA() error {
    innerFunction()
    result := A.DoSthReplyOnB()
    return result
}
`}
)

func main() {
	fset := token.NewFileSet()

	// sample: do cycle call check
	// imports collect imported package name as key, and dependency packages as values
	imports := make(map[string][]string)
	hasCycleImport := false
	for _, code := range codes {
		node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		// cycleImport collect conflicts
		var cycleImports []string

		// collect imports table and find the ring
		self := node.Name.Name
		for _, i := range node.Imports {
			// TODO: support SRC_PKG as PKG_NAME grammar
			sanitizedPackageName := strings.Trim(i.Path.Value, "\"")
			imports[self] = append(imports[self], sanitizedPackageName)
			if depImports, exists := imports[sanitizedPackageName]; exists {
				for _, depI := range depImports {
					if depI == self {
						cycleImports = append(cycleImports, sanitizedPackageName)
					}
				}
			}
		}

		if len(cycleImports) != 0 {
			hasCycleImport = true
			ast.Inspect(node, func(n ast.Node) bool {
				// Find CallExpr
				if caller, ok := n.(*ast.CallExpr); ok {
					// SelectorExpr token means it is package func caller
					if fun, ok := caller.Fun.(*ast.SelectorExpr); ok {
						funcName := fun.Sel.Name
						packageName := fun.X.(*ast.Ident).Name
						fmt.Printf("[CYCLE IMPORT] function call found on pkg %s line %d: \n\t%s.%s\n", self, fset.Position(caller.Lparen).Line, packageName, funcName)
					}
				}

				return true
			})
		}
	}

	if hasCycleImport {
		fmt.Println("Boom!")
	} else {
		fmt.Println("Good Job!")
	}
}
