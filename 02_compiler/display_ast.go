package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func main() {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "main.go", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	ast.Fprint(os.Stdout, fset, f, nil)
}
