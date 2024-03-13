package parser

import (
	"go/ast"
	"go/token"
)

type DecorComment struct {
	DecorName *ast.Ident      // example: 'handler' part of the handler("arg1","arg2")
	Gap       token.Pos       // distance from the start of the '//' of a comment to the actual name of the decor
	Args      []*ast.BasicLit //arguments passed to any decorator like handler("arg1","arg2")
}

// where is this decorator example FUNC
type ON int

const (
	FUNC ON = iota
	PARAM
)

const decorSymbol = '@'

// parses a ast.Comment to a decor, returns nil,nil if its not a decor, returns error if its decor and it has error
func NewDecorComment(c *ast.Comment) (*DecorComment, *DecorationErr) {
	gap := 0
	for _, v := range c.Text {
		if v == '/' || v == ' ' || v == '	' {
			gap++
			continue
		}
		if v == decorSymbol {
			gap++
			break
		}
		// if anythng else this is not a decor
		return nil, nil
	}
	txt := c.Text[gap:]

	x, _ := ParseExpr(txt)
	cx, ok := x.(*ast.CallExpr)
	if !ok {
		return nil, &DecorationErr{pos: c.Slash + token.Pos(gap), msg: "unable to parse decorator"}
	}
	fnIdent, ok := cx.Fun.(*ast.Ident)
	if !ok {
		return nil, &DecorationErr{pos: c.Slash + token.Pos(gap), msg: "expected an identifier"}
	}
	args := []*ast.BasicLit{}
	for _, v := range cx.Args {
		if id, ok := v.(*ast.BasicLit); ok {
			args = append(args, id)
		} else {
			return nil, &DecorationErr{pos: v.Pos() + token.Pos(gap), msg: "unexpected"}
		}
	}
	return &DecorComment{
		DecorName: fnIdent,
		Gap:       token.Pos(gap),
		Args:      args,
	}, nil
}

type FuncDecorGroup map[string]DecorComment

func NewFuncDecorGroup(c *ast.CommentGroup) (fdg FuncDecorGroup, errs []*DecorationErr) {
	if c == nil {
		return nil, nil
	}
	fdg = FuncDecorGroup{}
	errs = []*DecorationErr{}
	for _, v := range c.List {
		dc, err := NewDecorComment(v)
		if err != nil {
			errs = append(errs, err)
		}
		if dc != nil {
			fdg[dc.DecorName.Name] = *dc
		}
	}
	return
}
