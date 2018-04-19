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
	code = `
package validation
type Person struct {
    Age int ` + "`validate:\"lt=30\"` // limitation for `Age` lt than 30." + `
    Name string
}

type Animal struct {
    Age int ` + "`validate:\"gt=1,lt=3\"` // limitation for animal `Age`." + `
}`
)

func printCode(node ast.Node) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	printer.Fprint(&buf, fset, node)

	return buf.String()
}

func main() {
	fset := token.NewFileSet()

	// get struct tag, and generate codes
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// SelectorExpr token means it is external package caller
		if strukt, ok := n.(*ast.StructType); ok {
			fields := strukt.Fields.List
			for _, field := range fields {
				if field.Tag != nil {
					tag := strings.Trim(field.Tag.Value, "`")
					fmt.Println("tag: ", tag)

					// TODO: handle the rule and generate validation codes.
				}
			}
		}

		return true
	})
}
