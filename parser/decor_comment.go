package parser

import (
	"go/ast"
	"strings"
)

type DecorComment struct {
	astCmt *ast.Comment
	offset int
}

// parses a ast.Comment to a decor, returns true and non nil if its a decor
func NewDecorComment(c *ast.Comment) (*DecorComment, *DecorationErr) {
	txt, found := strings.CutPrefix(c.Text, "//")
	if !found {
		return nil, nil
	}
	offset := 2
	for {
		txt, found = strings.CutPrefix(txt, " ")
		if !found {
			break
		}
		offset++
	}

	x, _ := ParseExpr(txt)
	cx, ok := x.(*ast.CallExpr)
	if !ok {
		return nil, nil
	}
	ParseFnDecorator()
	return &DecorComment{
		astCmt: c,
		offset: offset,
	}, nil
}

// returns fun position of the
func (dc *DecorComment) FnPos()
