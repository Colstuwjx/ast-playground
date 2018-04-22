package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"regexp"
	"strings"
)

var (
	ErrUnknownTag = errors.New("unknown tag")

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

// unit validate func
type ValidationFunc func(v interface{}, param string) error

// tag represents one of the tag items
type tag struct {
	Name  string         // name of the tag
	Fn    ValidationFunc // validation function to call
	Param string         // parameter to send to the validation function
}

// separate by no escaped commas
var sepPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*),`)

// splitUnescapedComma splits tag expr using regexp
func splitUnescapedComma(str string) []string {
	ret := []string{}
	indexes := sepPattern.FindAllStringIndex(str, -1)
	last := 0
	for _, is := range indexes {
		ret = append(ret, str[last:is[1]-1])
		last = is[1]
	}
	ret = append(ret, str[last:])
	return ret
}

// parseTags parses all individual tags found within a struct tag.
func parseTags(t string) ([]tag, error) {
	tl := splitUnescapedComma(t)
	tags := make([]tag, 0, len(tl))
	for _, i := range tl {
		i = strings.Replace(i, `\,`, ",", -1)
		tg := tag{}
		v := strings.SplitN(i, "=", 2)
		tg.Name = strings.Trim(v[0], " ")
		if tg.Name == "" {
			return []tag{}, ErrUnknownTag
		}
		if len(v) > 1 {
			tg.Param = strings.Trim(v[1], " ")
		}

		// TODO: fillout validate funcs.
		tags = append(tags, tg)
	}
	return tags, nil
}

// printCode print the ast node source code.
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
					tags, err := parseTags(tag)
					if err != nil {
						panic(err)
					}

					fmt.Println("tags: ", tags)

					// TODO: generate the validation codes.
				}
			}
		}

		return true
	})
}
