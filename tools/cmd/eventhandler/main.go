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

type FuncField struct {
	Name       string
	IsEllipsis bool
	IsPointer  bool
	IsArray    bool
}

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

	file, err := parser.ParseFile(fs, "eventTypes.go", nil, 0)
	if err != nil {
		log.Fatalf("warning: internal error: could not parse %s: %s", "eventTypes.go", err)
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
			"isEllipsis":       isEllipsis,
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

	err = os.WriteFile(filepath.Join(dir, "eventHandlers.go"), src, 0644)
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
			assertions = append(assertions, fmt.Sprintf("a%d", i))
		}
	}

	return strings.Join(assertions, ", ")
}

func isEllipsis(s string) bool {
	return strings.HasPrefix(s, "...")
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
		params = append(params, resolveType(v.Type, ""))
	}
	return params
}

func resolveType(t ast.Expr, stringType string) string {
	switch et := t.(type) {
	case *ast.Ident:
		stringType += et.String()
		return stringType
	case *ast.ArrayType:
		stringType += "[]"
		return resolveType(et.Elt, stringType)
	case *ast.StarExpr:
		stringType += "*"
		return resolveType(et.X, stringType)
	case *ast.Ellipsis:
		stringType += "..."
		return resolveType(et.Elt, stringType)
	case *ast.ChanType:
		stringType += "chan "
		return resolveType(et.Value, stringType)
	case *ast.FuncType:
		var types []string
		stringType += "func("
		for _, v := range et.Params.List {
			types = append(types, resolveType(v.Type, ""))
		}
		stringType = stringType + strings.Join(types, ", ") + ")"
		return stringType
	case *ast.InterfaceType:
		// Anonymous interfaces are not usable, hence we don't parse methods for it
		stringType += "interface{}"
		return stringType
	case *ast.StructType:
		fields := make([]string, len(et.Fields.List))
		stringType += "struct{"
		for i, v := range et.Fields.List {
			name := make([]string, len(v.Names))
			for j, n := range v.Names {
				name[j] = n.String()
			}
			fields[i] = strings.Join(name, ", ") + " " + resolveType(v.Type, "")
		}
		stringType += strings.Join(fields, "\n") + "}"
	case *ast.MapType:
		stringType += fmt.Sprintf("map[%s]%s", resolveType(et.Key, ""), resolveType(et.Value, ""))
		return stringType
	default:
		return stringType
	}
	return stringType
}
