package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func init() {
	file, err := os.ReadFile("tools/cmd/eventhandler/event_handlers.gotext")
	if err != nil {
		log.Fatal("could not find template file", err)
	}

	textTemplate = string(file)
}

var textTemplate string

func main() {
	var buf bytes.Buffer
	dir := filepath.Dir(".")

	fs := token.NewFileSet()

	file, err := parser.ParseFile(fs, "event_types.go", nil, 0)
	if err != nil {
		log.Fatalf("warning: internal error: could not parse %s: %s", "event_types.go", err)
		return
	}

	types := make(map[string][]string)
	for _, v := range file.Decls {
		ts := astDeclToTypeSpec(v)
		if ts == nil {
			continue
		}
		fields := getFuncFieldsFromType(ts)
		if fields == nil {
			continue
		}
		types[ts.Name.String()] = fields
	}
	eventHandlerTmpl, err := template.New("eventHandlers").Funcs(
		template.FuncMap{
			"formatTypeAssert": formatTypeAssert,
			"join":             strings.Join,
			"hasEllipsis":      hasEllipsis,
			"getEllipsisType":  getEllipsisType,
		}).Parse(textTemplate)
	if err != nil {
		log.Fatal("couldn't create a new template", err)
	}

	err = eventHandlerTmpl.Execute(&buf, types)
	if err != nil {
		log.Println(err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Println("internal error: invalid Go generated:", err)
		src = buf.Bytes()
	}

	err = os.WriteFile(filepath.Join(dir, "event_handlers.go"), src, 0644)
	if err != nil {
		log.Fatal(buf, "writing output: %s", err)
	}
}

func formatTypeAssert(s []string) string {
	var assertions []string
	for i, t := range s {
		if strings.HasPrefix(t, "...") {
			assertions = append(assertions, "ellipsis...")
		} else {
			assertions = append(assertions, fmt.Sprintf("a[%d].(%s)", i, t))
		}
	}

	return strings.Join(assertions, ", ")
}

func hasEllipsis(s []string) bool {
	if len(s) == 0 {
		return false
	}

	return strings.HasPrefix(s[len(s)-1], "...")
}

func getEllipsisType(s []string) string {
	if len(s) == 0 {
		return "interface{}"
	}

	return strings.Replace(s[len(s)-1], "...", "", 1)
}

func astDeclToTypeSpec(d ast.Decl) *ast.TypeSpec {
	gd, ok := d.(*ast.GenDecl)
	if !ok {
		return nil
	}

	ts, ok := gd.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil
	}
	return ts
}

func getFuncFieldsFromType(d *ast.TypeSpec) (params []string) {
	params = []string{}
	ft, ok := d.Type.(*ast.FuncType)
	if !ok {
		return nil
	}
	for _, v := range ft.Params.List {
		var ident string

		switch et := v.Type.(type) {
		case *ast.StarExpr:
			ident = "*" + et.X.(*ast.Ident).Name
		case *ast.Ident:
			ident = et.Name
		case *ast.ArrayType:
			ident = "[ ]" + et.Elt.(*ast.Ident).String()
		case *ast.Ellipsis:
			ident = "..." + et.Elt.(*ast.Ident).String()
		}
		params = append(params, ident)
	}

	return params
}
