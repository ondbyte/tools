package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go/ast"
)

type Handler struct {
	Raw    string
	Params map[string]bool
}

// handlerDecor implements HandlerDecorExpr.
func (e *Handler) handlerDecor() {
	panic("unimplemented")
}

var endpointRegex = regexp.MustCompile(`endpoint\("([^"]+)"\)`)
var endpointContentRegex = regexp.MustCompile(`{([^}]*)}`)

type Path struct {
	Raw    string
	Params map[string]ast.Expr
}

var pathRegex = regexp.MustCompile(`path\("([^"]+)"\)`)

type Query struct {
	Raw    string
	Params map[string]string
}

var queryRegex = regexp.MustCompile(`query\("([^"]+)"\)`)

type Header struct {
	Raw    string
	Params map[string]string
}

var headerRegex = regexp.MustCompile(`header\("([^"]+)"\)`)

type Body struct {
	Raw    string
	Params map[string]string
}

var bodyRegex = regexp.MustCompile(`body\("([^"]+)"\)`)

type Description struct {
	Data string
}

// handlerDecor implements HandlerDecorExpr.
func (d *Description) handlerDecor() {
	panic("unimplemented")
}

var descriptionRegex = regexp.MustCompile(`description\("([^"]+)"\)`)

func ParseHandler(ep *ast.CallExpr) (*Handler, error) {
	if i, ok := ep.Fun.(*ast.Ident); ok {
		if i.Name != "endpoint" {
			return nil, nil
		}
		if len(ep.Args) == 0 {
			return nil, fmt.Errorf("endpoint requires 1 string argument")
		}
		if i, ok := ep.Args[0].(*ast.BasicLit); ok {
			us, err := strconv.Unquote(i.Value)
			if err != nil {
				return nil, fmt.Errorf("error unquoting string '%v' : %v", i.Value, err)
			}
			matches := endpointContentRegex.FindAllStringSubmatch(us, -1)
			params := map[string]bool{}
			for _, match := range matches {
				if params[match[1]] {
					return nil, fmt.Errorf("two path params with same name %v", match[1])
				}
				params[match[1]] = true
			}
			return &Handler{
				Params: params,
			}, nil
		}
	}
	return nil, nil
}

func ProcessPath(pathX *ast.CallExpr, paramType ast.Expr) (*Path, error) {
	if i, ok := pathX.Fun.(*ast.Ident); ok {
		if i.Name != "path" {
			return nil, nil
		}
		if len(pathX.Args) == 0 {
			return nil, fmt.Errorf("path requires 1 string argument")
		}
		if i, ok := pathX.Args[0].(*ast.BasicLit); ok {
			us, err := strconv.Unquote(i.Value)
			if err != nil {
				return nil, fmt.Errorf("error unquoting string '%v' : %v", i.Value, err)
			}
			return &Path{
				Params: map[string]ast.Expr{
					us: paramType,
				},
			}, nil
		}
	}
	return nil, nil
}

func ParseDescr(ep *ast.CallExpr) (*Description, error) {
	if i, ok := ep.Fun.(*ast.Ident); ok {
		if i.Name != "description" {
			return nil, nil
		}
		if len(ep.Args) == 0 {
			return nil, fmt.Errorf("description requires 1 string argument")
		}
		if i, ok := ep.Args[0].(*ast.BasicLit); ok {
			us, err := strconv.Unquote(i.Value)
			if err != nil {
				return nil, fmt.Errorf("error unquoting string '%v' : %v", i.Value, err)
			}
			return &Description{
				Data: us,
			}, nil
		}
	}
	return nil, nil
}

func ProcessParamCmt(cmt *ast.Comment, paramType ast.Expr) (any, error) {
	text := cmt.Text
	text = strings.Trim(text, "/")
	if x, err := ParseExpr(text); err != nil {
		return nil, err
	} else if xx, ok := x.(*ast.CallExpr); !ok {
		return nil, nil
	} else if ep, err := ProcessPath(xx, paramType); err != nil {
		return nil, err
	} else if ep != nil {
		return ep, nil
	}
	return nil, nil
}

type HandlerDecorExpr interface {
	handlerDecor()
}

func ParseHandlerComment(cmt *ast.Comment) (HandlerDecorExpr, error) {
	text := cmt.Text
	text = strings.Trim(text, "/")
	if x, err := ParseExpr(text); err != nil {
		return nil, err
	} else if xx, ok := x.(*ast.CallExpr); !ok {
		return nil, nil
	} else if ep, err := ParseHandler(xx); err != nil {
		return nil, err
	} else if ep != nil {
		return ep, nil
	} else if descr, err := ParseDescr(xx); err != nil {
		return nil, err
	} else if descr != nil {
		return descr, nil
	}
	return nil, nil
}

type HandlerDecors []HandlerDecorExpr

func (p *parser) ParseHandlerDecorators(fnName *ast.Ident, fnParams, fnResults *ast.FieldList) (decors HandlerDecors) {
	decors = make(HandlerDecors, 0)
	var isHandler = false
	if p.leadComment != nil && len(p.leadComment.List) > 0 {
		for _, cmt := range p.leadComment.List {
			decor, err := ParseHandlerComment(cmt)
			if err != nil {
				p.error(fnName.NamePos, err.Error())
			}
			if !isHandler {
				_, isHandler = decor.(*Handler)
				if !isHandler {
					// dont bother to parse comments for this fn, as the required 'handler' decor is not found
					return nil
				}
			}
			decors = append(decors, decor)
		}
	}
	return
}

func (p *parser) ParseParamDecorators() {

}
