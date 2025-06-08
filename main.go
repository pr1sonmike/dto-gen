package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"
)

type Field struct {
	Name    string
	Type    string
	TagName string
}

type TemplateData struct {
	PackageName  string
	DTOName      string
	OriginalName string
	MapperFunc   string
	Fields       []Field
}

const tpl = `package {{.PackageName}}

type {{.DTOName}} struct {
{{- range .Fields }}
	{{ .Name }} {{ .Type }} ` + "`json:\"{{ .TagName }}\"`" + `
{{- end }}
}

func {{.MapperFunc}}(in {{.OriginalName}}) {{.DTOName}} {
	return {{.DTOName}}{
{{- range .Fields }}
		{{ .Name }}: in.{{ .Name }},
{{- end }}
	}
}
`

func main() {
	input := flag.String("input", "", "Input Go file")
	output := flag.String("output", "", "Output Go file")
	structName := flag.String("type", "", "Struct name to convert")
	flag.Parse()

	if *input == "" || *output == "" || *structName == "" {
		fmt.Println("Usage: dto-gen --input file.go --output file_dto.go --type StructName")
		os.Exit(1)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *input, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	var fields []Field
	var packageName = node.Name.Name

	// Find the struct
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			tspec := spec.(*ast.TypeSpec)
			if tspec.Name.Name != *structName {
				continue
			}
			stype, ok := tspec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range stype.Fields.List {
				typeStr := exprToString(field.Type)
				for _, name := range field.Names {
					fields = append(fields, Field{
						Name:    name.Name,
						Type:    typeStr,
						TagName: strings.ToLower(name.Name),
					})
				}
			}
		}
	}

	data := TemplateData{
		PackageName:  packageName,
		DTOName:      *structName + "DTO",
		OriginalName: *structName,
		MapperFunc:   "To" + *structName + "DTO",
		Fields:       fields,
	}

	t := template.Must(template.New("dto").Parse(tpl))
	out, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = t.Execute(out, data)
	if err != nil {
		panic(err)
	}

	fmt.Println("DTO generated at", *output)
}

// Helper: convert type expression to string
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}
