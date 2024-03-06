package parser

import (
	"fmt"
	"go/ast"
)

func StringifiedType(x ast.Expr) (t string, err error) {
	_sx, ok := x.(*ast.StarExpr)
	if ok {
		x = _sx.X
		t += "*"
	}
	_x, ok := x.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("only pointer type or value type can be parameters, example: *int/int")
	}
	t += _x.Name
	return
}
