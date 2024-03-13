package parser

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestNewDecorComment(t *testing.T) {
	_, f, _ := ParseFile(token.NewFileSet(),
		"xyz.go",
		`package xyz
	// @decor1("yadu",1)
	// @
	//       	@decor2("")
	func yadu(){}
	`, ParseComments)
	for _, decl := range f.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if ok && decl.Name.Name == "yadu" {
			dc, dErr := NewDecorComment(decl.Doc.List[0])
			if dErr != nil {
				t.Fatal("expected no err but got", dErr)
			}
			if dc.DecorName.Name!="decor1"{
				t.Fatal("expected decor name to be decor1 but got ",dc.DecorName.Name)
			}
			if len(dc.Args) != 2 {
				t.Fatal("expected 2 args but got ", len(dc.Args))
			}
			_, dErr = NewDecorComment(decl.Doc.List[1])
			if dErr == nil {
				t.Fatal("expected err but got", nil)
			}
			dc, dErr = NewDecorComment(decl.Doc.List[2])
			if dErr != nil {
				t.Fatal("expected no err but got", dErr)
			}
			if dc.DecorName.Name!="decor2"{
				t.Fatal("expected decor name to be decor2 but got ",dc.DecorName.Name)
			}
			if len(dc.Args)!=1 {
				t.Fatal("expected 1 args but got ", len(dc.Args))
			}
		}
	}
}

