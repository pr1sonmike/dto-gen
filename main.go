package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
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
		log.Fatal("Usage: dto-gen --input file.go --output file_dto.go --type StructName")
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *input, nil, parser.AllErrors)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	fields, packageName := findStructFields(node, *structName)
	if fields == nil {
		log.Fatalf("Struct %s not found in file %s", *structName, *input)
	}

	data := TemplateData{
		PackageName:  packageName,
		DTOName:      *structName + "DTO",
		OriginalName: *structName,
		MapperFunc:   "To" + *structName + "DTO",
		Fields:       fields,
	}

	t, err := template.New("dto").Parse(tpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	out, err := os.Create(*output)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer out.Close()

	if err := t.Execute(out, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println("DTO generated at", *output)
}

func findStructFields(node *ast.File, structName string) ([]Field, string) {
	var fields []Field
	packageName := node.Name.Name

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			tspec, ok := spec.(*ast.TypeSpec)
			if !ok || tspec.Name.Name != structName {
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
	return fields, packageName
}

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
