package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
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

// printCode print the ast node source code.
func printCode(node ast.Node) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	printer.Fprint(&buf, fset, node)

	return buf.String()
}

func main() {
	fset := token.NewFileSet()

	// sample: do cycle call check
	// imports collect imported package name as key, and dependency packages as values
	// calls collect package's call, exported fields or func call
	imports := make(map[string][]string)
	calls := make(map[string]map[string][]string)
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
		}

		ast.Inspect(node, func(n ast.Node) bool {
			// SelectorExpr token means it is external package caller
			if expr, ok := n.(*ast.SelectorExpr); ok {
				callCode := printCode(n)
				if calls[self] == nil {
					calls[self] = make(map[string][]string)
				}

				pkg := expr.X.(*ast.Ident).Name
				calls[self][pkg] = append(calls[self][pkg], callCode)
				if len(cycleImports) != 0 {
					for _, ci := range cycleImports {
						fmt.Printf("[CYCLE IMPORT] function call found on pkg %s line %d: %s\n", self, fset.Position(expr.Pos()).Line, callCode)
						if refered, exists := calls[ci][self]; exists {
							fmt.Printf("refered pkg %s usage: %s\n", ci, refered)
						}
					}
				}
			}

			return true
		})
	}

	if hasCycleImport {
		fmt.Println("Boom!")
	} else {
		fmt.Println("Good Job!")
	}
}
